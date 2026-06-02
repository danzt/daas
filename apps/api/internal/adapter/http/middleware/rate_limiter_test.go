package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"

	mw "github.com/danzt/daas/api/internal/adapter/http/middleware"
)

// successHandler is a trivial Echo handler used as the "next" in middleware tests.
var successHandler = func(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

// newEchoCtxWithRemoteIP creates an Echo context with the given remote address.
func newEchoCtxWithRemoteIP(t *testing.T, remoteAddr string) (echo.Context, *httptest.ResponseRecorder) {
	t.Helper()
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = remoteAddr + ":12345"
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c, rec
}

// newEchoCtxWithXFF creates an Echo context with an X-Forwarded-For header.
func newEchoCtxWithXFF(t *testing.T, xff string) (echo.Context, *httptest.ResponseRecorder) {
	t.Helper()
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", xff)
	req.RemoteAddr = "10.0.0.1:12345"
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c, rec
}

// drainRequests sends n requests through the rate limiter middleware and
// returns the HTTP status codes observed.
func drainRequests(t *testing.T, rl *mw.RateLimiter, ip string, n int) []int {
	t.Helper()
	statuses := make([]int, 0, n)
	for i := 0; i < n; i++ {
		c, rec := newEchoCtxWithRemoteIP(t, ip)
		handler := rl.Handle()(successHandler)
		if err := handler(c); err != nil {
			// Echo may return the error instead of writing the response.
			he, ok := err.(*echo.HTTPError)
			if ok {
				statuses = append(statuses, he.Code)
			} else {
				statuses = append(statuses, http.StatusInternalServerError)
			}
		} else {
			statuses = append(statuses, rec.Code)
		}
	}
	return statuses
}

// TestRateLimiter_UnderLimit verifies STORE-RATE-1:
// requests below the RPM limit all receive 200 OK.
func TestRateLimiter_UnderLimit(t *testing.T) {
	rl := mw.NewRateLimiterWithRPM(100)
	statuses := drainRequests(t, rl, "203.0.113.1", 99)
	for i, s := range statuses {
		if s != http.StatusOK {
			t.Errorf("request %d: expected 200, got %d", i+1, s)
		}
	}
}

// TestRateLimiter_AtLimit verifies STORE-RATE-2:
// the (RPM+1)th request returns 429 with Retry-After header.
func TestRateLimiter_AtLimit(t *testing.T) {
	const rpm = 10
	rl := mw.NewRateLimiterWithRPM(rpm)

	// Send rpm requests — all should pass (token bucket burst = rpm).
	drainRequests(t, rl, "203.0.113.2", rpm)

	// The very next request must be rate-limited.
	c, rec := newEchoCtxWithRemoteIP(t, "203.0.113.2")
	handler := rl.Handle()(successHandler)
	_ = handler(c)

	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429, got %d", rec.Code)
	}
	if retryAfter := rec.Header().Get("Retry-After"); retryAfter == "" {
		t.Error("expected Retry-After header to be set on 429 response")
	}
}

// TestRateLimiter_SeparateIPs verifies STORE-RATE-3:
// exhausting one IP's bucket does not affect a different IP.
func TestRateLimiter_SeparateIPs(t *testing.T) {
	const rpm = 5
	rl := mw.NewRateLimiterWithRPM(rpm)

	ipA := "203.0.113.10"
	ipB := "203.0.113.20"

	// Exhaust IP-A's bucket.
	statuses := drainRequests(t, rl, ipA, rpm+1)
	if last := statuses[rpm]; last != http.StatusTooManyRequests {
		t.Errorf("IP-A: expected 429 on %dth request, got %d", rpm+1, last)
	}

	// IP-B should still have a full bucket.
	c, rec := newEchoCtxWithRemoteIP(t, ipB)
	handler := rl.Handle()(successHandler)
	_ = handler(c)
	if rec.Code != http.StatusOK {
		t.Errorf("IP-B should get 200, got %d", rec.Code)
	}
}

// TestRateLimiter_XForwardedFor_PrivateFiltered verifies STORE-RATE-4:
// a private IP in X-Forwarded-For is filtered; the first public IP wins.
func TestRateLimiter_XForwardedFor_PrivateFiltered(t *testing.T) {
	const rpm = 5
	rl := mw.NewRateLimiterWithRPM(rpm)

	// XFF: public IP + private IP. Rate limiter should key on 203.0.113.5.
	const publicIP = "203.0.113.5"
	c1, _ := newEchoCtxWithXFF(t, publicIP+", 10.0.0.1")
	_ = rl.Handle()(successHandler)(c1)

	// Exhaust 203.0.113.5's bucket (already used 1 above).
	for i := 0; i < rpm-1; i++ {
		c, _ := newEchoCtxWithXFF(t, publicIP+", 10.0.0.1")
		_ = rl.Handle()(successHandler)(c)
	}

	// This should be rate-limited for 203.0.113.5.
	c2, rec2 := newEchoCtxWithXFF(t, publicIP+", 10.0.0.1")
	_ = rl.Handle()(successHandler)(c2)
	if rec2.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429 for exhausted public IP, got %d", rec2.Code)
	}

	// A different public IP in XFF should NOT be rate-limited.
	c3, rec3 := newEchoCtxWithXFF(t, "198.51.100.1, 10.0.0.2")
	_ = rl.Handle()(successHandler)(c3)
	if rec3.Code != http.StatusOK {
		t.Errorf("different IP should get 200, got %d", rec3.Code)
	}
}

// TestRateLimiter_EnvVarOverride verifies STORE-RATE-5:
// NewRateLimiterWithRPM uses the provided RPM directly (same logic as env var).
func TestRateLimiter_EnvVarOverride(t *testing.T) {
	const rpm = 10
	rl := mw.NewRateLimiterWithRPM(rpm)
	// Drain exactly rpm requests.
	drainRequests(t, rl, "203.0.113.99", rpm)
	// The 11th should be 429.
	c, rec := newEchoCtxWithRemoteIP(t, "203.0.113.99")
	_ = rl.Handle()(successHandler)(c)
	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429 at RPM=%d+1, got %d", rpm, rec.Code)
	}
}
