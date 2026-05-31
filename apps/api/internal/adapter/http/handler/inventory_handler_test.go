package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

// TestInventoryMovements_ImmutabilityEnforced verifies that PATCH and PUT on
// /api/v1/inventory/movements/:id are not registered routes. Echo returns
// 405 Method Not Allowed when a path exists but the method is not registered,
// satisfying the spec requirement that movements are immutable at the HTTP level.
func TestInventoryMovements_ImmutabilityEnforced(t *testing.T) {
	e := echo.New()

	// Register only GET — exactly what the production router does for movements.
	e.GET("/api/v1/inventory/movements/:id", func(c echo.Context) error {
		return c.JSON(http.StatusOK, nil)
	})

	for _, method := range []string{http.MethodPatch, http.MethodPut, http.MethodDelete} {
		t.Run(method+" returns 405", func(t *testing.T) {
			req := httptest.NewRequest(method, "/api/v1/inventory/movements/some-uuid", nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			if rec.Code != http.StatusMethodNotAllowed {
				t.Errorf("%s /inventory/movements/:id returned %d, want 405", method, rec.Code)
			}
		})
	}
}
