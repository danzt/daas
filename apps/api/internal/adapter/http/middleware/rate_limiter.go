package middleware

import (
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/time/rate"
)

// privateRanges lists the IPv4 CIDR blocks that are considered private /
// loopback and must be filtered out when parsing X-Forwarded-For headers.
var privateRanges = func() []*net.IPNet {
	cidrs := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
	}
	nets := make([]*net.IPNet, 0, len(cidrs))
	for _, cidr := range cidrs {
		_, n, _ := net.ParseCIDR(cidr)
		nets = append(nets, n)
	}
	return nets
}()

// isPrivateIP returns true when the parsed IP falls within any private range.
func isPrivateIP(ip net.IP) bool {
	for _, n := range privateRanges {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

// limiterEntry bundles a token-bucket rate.Limiter with the timestamp of the
// last request, used for background cleanup.
type limiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiter is an in-memory, per-IP token-bucket rate limiter.
// It uses sync.Map for thread safety and starts a background goroutine that
// evicts entries not seen in 10 minutes (ticking every 5 minutes).
type RateLimiter struct {
	limiters sync.Map
	rpm      int
	mu       sync.Mutex // guards lastSeen writes
}

// NewRateLimiter constructs a RateLimiter reading RPM from the environment
// variable PUBLIC_RATE_LIMIT_RPM (default 100).
func NewRateLimiter() *RateLimiter {
	rpm := 100
	if v := os.Getenv("PUBLIC_RATE_LIMIT_RPM"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			rpm = n
		}
	}
	return NewRateLimiterWithRPM(rpm)
}

// NewRateLimiterWithRPM constructs a RateLimiter with an explicit RPM value.
// The burst capacity equals rpm (full minute's quota available as a burst).
// A background goroutine is started to evict stale entries.
func NewRateLimiterWithRPM(rpm int) *RateLimiter {
	rl := &RateLimiter{rpm: rpm}
	go rl.cleanupLoop()
	return rl
}

// cleanupLoop evicts limiter entries that haven't been seen in 10 minutes.
// Runs indefinitely in a goroutine; one goroutine per RateLimiter instance.
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.limiters.Range(func(key, value any) bool {
			entry := value.(*limiterEntry)
			rl.mu.Lock()
			lastSeen := entry.lastSeen
			rl.mu.Unlock()
			if time.Since(lastSeen) > 10*time.Minute {
				rl.limiters.Delete(key)
			}
			return true
		})
	}
}

// extractPublicIP parses the X-Forwarded-For header and returns the first
// non-private IP address found. Falls back to c.RealIP() if no public IP
// is present in the header.
func (rl *RateLimiter) extractPublicIP(c echo.Context) string {
	xff := c.Request().Header.Get("X-Forwarded-For")
	if xff != "" {
		for _, rawIP := range strings.Split(xff, ",") {
			trimmed := strings.TrimSpace(rawIP)
			ip := net.ParseIP(trimmed)
			if ip != nil && !isPrivateIP(ip) {
				return trimmed
			}
		}
	}
	// Fallback: strip port from c.RealIP().
	realIP := c.RealIP()
	if host, _, err := net.SplitHostPort(realIP); err == nil {
		return host
	}
	return realIP
}

// getLimiter retrieves (or lazily creates) the rate.Limiter for the given IP.
func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	if raw, ok := rl.limiters.Load(ip); ok {
		entry := raw.(*limiterEntry)
		rl.mu.Lock()
		entry.lastSeen = time.Now()
		rl.mu.Unlock()
		return entry.limiter
	}
	limiter := rate.NewLimiter(rate.Every(time.Minute/time.Duration(rl.rpm)), rl.rpm)
	entry := &limiterEntry{limiter: limiter, lastSeen: time.Now()}
	// LoadOrStore handles the race where two goroutines create entries for the
	// same IP concurrently.
	actual, _ := rl.limiters.LoadOrStore(ip, entry)
	return actual.(*limiterEntry).limiter
}

// Handle returns the Echo middleware function for rate limiting.
func (rl *RateLimiter) Handle() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip := rl.extractPublicIP(c)
			limiter := rl.getLimiter(ip)

			if !limiter.Allow() {
				delay := limiter.Reserve().Delay()
				seconds := int(delay.Seconds())
				if seconds < 1 {
					seconds = 1
				}
				c.Response().Header().Set("Retry-After", strconv.Itoa(seconds))
				return c.JSON(http.StatusTooManyRequests, map[string]any{
					"error":               "rate_limit_exceeded",
					"retry_after_seconds": seconds,
				})
			}

			return next(c)
		}
	}
}
