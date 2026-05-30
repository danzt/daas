package handler

import (
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"

	"github.com/danzt/daas/api/internal/app"
	"github.com/danzt/daas/api/internal/domain/report"
)

// ReportHandler exposes read-only reporting endpoints.
type ReportHandler struct {
	svc *app.ReportService
}

func NewReportHandler(pool *pgxpool.Pool) *ReportHandler {
	return &ReportHandler{svc: app.NewReportService(pool)}
}

// parseDateRange reads ?from= and ?to= query params.
// Defaults: from = 30 days ago, to = now.
func parseDateRange(c echo.Context) report.DateRange {
	now := time.Now().UTC()
	defaultFrom := now.AddDate(0, 0, -30)

	dr := report.DateRange{From: defaultFrom, To: now}

	if f := c.QueryParam("from"); f != "" {
		if t, err := time.Parse("2006-01-02", f); err == nil {
			dr.From = t.UTC()
		}
	}
	if t := c.QueryParam("to"); t != "" {
		if parsed, err := time.Parse("2006-01-02", t); err == nil {
			dr.To = parsed.Add(24*time.Hour - time.Second).UTC() // inclusive end-of-day
		}
	}

	return dr
}

// SalesReport godoc
// GET /api/v1/reports/sales?from=YYYY-MM-DD&to=YYYY-MM-DD
func (h *ReportHandler) SalesReport(c echo.Context) error {
	tenantID, err := getTenantID(c)
	if err != nil {
		return WriteProblem(c, http.StatusForbidden, "forbidden", "tenant context missing")
	}
	dr := parseDateRange(c)
	result, err := h.svc.SalesReport(c.Request().Context(), tenantID, dr)
	if err != nil {
		return WriteProblem(c, http.StatusInternalServerError, "internal-error", err.Error())
	}
	return c.JSON(http.StatusOK, result)
}

// InventoryReport godoc
// GET /api/v1/reports/inventory
func (h *ReportHandler) InventoryReport(c echo.Context) error {
	tenantID, err := getTenantID(c)
	if err != nil {
		return WriteProblem(c, http.StatusForbidden, "forbidden", "tenant context missing")
	}
	result, err := h.svc.InventorySnapshot(c.Request().Context(), tenantID)
	if err != nil {
		return WriteProblem(c, http.StatusInternalServerError, "internal-error", err.Error())
	}
	return c.JSON(http.StatusOK, result)
}

// PurchaseReport godoc
// GET /api/v1/reports/purchases?from=YYYY-MM-DD&to=YYYY-MM-DD
func (h *ReportHandler) PurchaseReport(c echo.Context) error {
	tenantID, err := getTenantID(c)
	if err != nil {
		return WriteProblem(c, http.StatusForbidden, "forbidden", "tenant context missing")
	}
	dr := parseDateRange(c)
	result, err := h.svc.PurchaseReport(c.Request().Context(), tenantID, dr)
	if err != nil {
		return WriteProblem(c, http.StatusInternalServerError, "internal-error", err.Error())
	}
	return c.JSON(http.StatusOK, result)
}
