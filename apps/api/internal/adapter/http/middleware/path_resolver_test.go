package middleware_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/labstack/echo/v4"

	mw "github.com/danzt/daas/api/internal/adapter/http/middleware"
)

// mockSlugLookup implements mw.SlugLookup for unit tests.
// It returns a fixed (tenantID, err) pair and counts how many times it was called.
type mockSlugLookup struct {
	tenantID string
	err      error
	calls    atomic.Int32
}

func (m *mockSlugLookup) LookupTenantBySlug(_ context.Context, slug string) (string, error) {
	m.calls.Add(1)
	_ = slug
	return m.tenantID, m.err
}

// newEchoCtxWithSlug creates an Echo context with ":tenantSlug" path param set.
func newEchoCtxWithSlug(t *testing.T, slug string) echo.Context {
	t.Helper()
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/t/"+slug+"/shop/v1/products", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("tenantSlug")
	c.SetParamValues(slug)
	return c
}

// TestPathResolver_HappyPath verifies STORE-RESOLVE-2:
// an active tenant slug resolves to the correct UUID.
func TestPathResolver_HappyPath(t *testing.T) {
	const (
		slug     = "tienda-x"
		tenantID = "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
	)
	lookup := &mockSlugLookup{tenantID: tenantID}
	resolver := mw.NewPathResolverWithLookup(lookup, 5*time.Minute)

	got, err := resolver.Resolve(newEchoCtxWithSlug(t, slug))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != tenantID {
		t.Errorf("tenantID: got %q, want %q", got, tenantID)
	}
}

// TestPathResolver_UnknownSlug verifies STORE-RESOLVE-3:
// an unknown slug returns ErrTenantNotFound and is NOT cached.
func TestPathResolver_UnknownSlug(t *testing.T) {
	lookup := &mockSlugLookup{tenantID: "", err: mw.ErrTenantNotFound}
	resolver := mw.NewPathResolverWithLookup(lookup, 5*time.Minute)

	_, err := resolver.Resolve(newEchoCtxWithSlug(t, "unknown-slug"))
	if !errors.Is(err, mw.ErrTenantNotFound) {
		t.Fatalf("expected ErrTenantNotFound, got %v", err)
	}
}

// TestPathResolver_InactiveTenant verifies that inactive tenants behave
// identically to unknown slugs — the SECURITY DEFINER function already
// filters by status='active', so the resolver just sees ErrTenantNotFound.
func TestPathResolver_InactiveTenant(t *testing.T) {
	lookup := &mockSlugLookup{tenantID: "", err: mw.ErrTenantNotFound}
	resolver := mw.NewPathResolverWithLookup(lookup, 5*time.Minute)

	_, err := resolver.Resolve(newEchoCtxWithSlug(t, "suspended-tenant"))
	if !errors.Is(err, mw.ErrTenantNotFound) {
		t.Fatalf("expected ErrTenantNotFound, got %v", err)
	}
}

// TestPathResolver_CacheHit verifies STORE-RESOLVE-4:
// a second resolve within TTL MUST NOT hit the DB again.
func TestPathResolver_CacheHit(t *testing.T) {
	const tenantID = "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
	lookup := &mockSlugLookup{tenantID: tenantID}
	resolver := mw.NewPathResolverWithLookup(lookup, 5*time.Minute)

	slug := "tienda-y"
	c := newEchoCtxWithSlug(t, slug)

	// First call — should hit DB.
	if _, err := resolver.Resolve(c); err != nil {
		t.Fatalf("first resolve error: %v", err)
	}
	// Second call — cache hit, DB should NOT be called again.
	if _, err := resolver.Resolve(c); err != nil {
		t.Fatalf("second resolve error: %v", err)
	}

	if n := lookup.calls.Load(); n != 1 {
		t.Errorf("DB called %d times, expected exactly 1 (cache should have hit)", n)
	}
}

// TestPathResolver_CacheExpiry verifies that an expired cache entry triggers
// a fresh DB lookup.
func TestPathResolver_CacheExpiry(t *testing.T) {
	const tenantID = "cccccccc-cccc-cccc-cccc-cccccccccccc"
	lookup := &mockSlugLookup{tenantID: tenantID}
	// Set a very short TTL so the entry expires immediately.
	resolver := mw.NewPathResolverWithLookup(lookup, 1*time.Nanosecond)

	slug := "tienda-z"
	c := newEchoCtxWithSlug(t, slug)

	if _, err := resolver.Resolve(c); err != nil {
		t.Fatalf("first resolve error: %v", err)
	}
	// Wait a bit beyond the nanosecond TTL.
	time.Sleep(10 * time.Millisecond)

	if _, err := resolver.Resolve(c); err != nil {
		t.Fatalf("second resolve error: %v", err)
	}

	if n := lookup.calls.Load(); n != 2 {
		t.Errorf("DB called %d times, expected 2 (expired entry should re-hit DB)", n)
	}
}

// TestPathResolver_NegativeNotCached verifies STORE-RESOLVE-3:
// negative results are NOT cached — two calls for unknown slugs hit DB twice.
func TestPathResolver_NegativeNotCached(t *testing.T) {
	lookup := &mockSlugLookup{tenantID: "", err: mw.ErrTenantNotFound}
	resolver := mw.NewPathResolverWithLookup(lookup, 5*time.Minute)

	for i := 0; i < 2; i++ {
		_, err := resolver.Resolve(newEchoCtxWithSlug(t, "no-such-tenant"))
		if !errors.Is(err, mw.ErrTenantNotFound) {
			t.Fatalf("call %d: expected ErrTenantNotFound, got %v", i+1, err)
		}
	}

	if n := lookup.calls.Load(); n != 2 {
		t.Errorf("DB called %d times, expected 2 (negatives must not be cached)", n)
	}
}

// TestPathResolver_Concurrent verifies STORE-RESOLVE-6 (thread safety):
// 100 goroutines resolving the same slug must not data-race.
// Run with -race to detect sync.Map or cache struct races.
func TestPathResolver_Concurrent(t *testing.T) {
	const tenantID = "dddddddd-dddd-dddd-dddd-dddddddddddd"
	lookup := &mockSlugLookup{tenantID: tenantID}
	resolver := mw.NewPathResolverWithLookup(lookup, 5*time.Minute)

	const goroutines = 100
	var wg sync.WaitGroup
	wg.Add(goroutines)
	errs := make(chan error, goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			c := newEchoCtxWithSlug(t, "concurrent-slug")
			got, err := resolver.Resolve(c)
			if err != nil {
				errs <- err
				return
			}
			if got != tenantID {
				errs <- errors.New("wrong tenantID")
			}
		}()
	}

	wg.Wait()
	close(errs)
	for err := range errs {
		t.Errorf("goroutine error: %v", err)
	}
}
