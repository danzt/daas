package handler

import (
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"

	mw "github.com/danzt/daas/api/internal/adapter/http/middleware"
)

// MeHandler handles the GET /api/v1/tenants/me endpoint.
// Returns the current tenant's profile data needed by the frontend store.
type MeHandler struct {
	pool *pgxpool.Pool
}

// NewMeHandler creates a MeHandler backed by the given connection pool.
func NewMeHandler(pool *pgxpool.Pool) *MeHandler {
	return &MeHandler{pool: pool}
}

// tenantMeResponse is the JSON shape returned by GET /api/v1/tenants/me.
type tenantMeResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	FiscalID    string `json:"fiscal_id"`
	CountryCode string `json:"country_code"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
}

// GetMe handles GET /api/v1/tenants/me.
// Returns the authenticated tenant's profile. Uses the tenant_id from the JWT
// (set by auth middleware) to look up the tenant row.
func (h *MeHandler) GetMe(c echo.Context) error {
	tenantID, _ := c.Get(string(mw.ContextKeyTenantID)).(string)
	if tenantID == "" {
		return WriteProblem(c, http.StatusForbidden, "forbidden", "tenant context missing")
	}

	var resp tenantMeResponse
	var fiscalID *string
	var createdAt time.Time

	err := h.pool.QueryRow(c.Request().Context(),
		`SELECT id, name, fiscal_id, country_code, status, created_at
		 FROM tenants
		 WHERE id = $1`,
		tenantID,
	).Scan(&resp.ID, &resp.Name, &fiscalID, &resp.CountryCode, &resp.Status, &createdAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return WriteProblem(c, http.StatusNotFound, "tenant-not-found", "tenant not found")
		}
		return WriteProblem(c, http.StatusInternalServerError, "internal-error", "failed to fetch tenant")
	}

	if fiscalID != nil {
		resp.FiscalID = *fiscalID
	}
	resp.CreatedAt = createdAt.Format(time.RFC3339)

	return c.JSON(http.StatusOK, resp)
}
