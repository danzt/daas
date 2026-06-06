package handler

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"

	invoicepdf "github.com/danzt/daas/api/internal/adapter/pdf"
	"github.com/danzt/daas/api/internal/app"
)

// PDFHandler serves PDF downloads for internal and fiscal invoices.
// It is a thin orchestrator: loads invoice data via the services,
// fetches the tenant header info from the pool, and delegates rendering
// to the invoicepdf package.
type PDFHandler struct {
	pool        *pgxpool.Pool
	internalSvc *app.InternalInvoiceService
	fiscalSvc   *app.FiscalInvoiceService
}

// NewPDFHandler creates a PDFHandler.
func NewPDFHandler(pool *pgxpool.Pool, internalSvc *app.InternalInvoiceService, fiscalSvc *app.FiscalInvoiceService) *PDFHandler {
	return &PDFHandler{pool: pool, internalSvc: internalSvc, fiscalSvc: fiscalSvc}
}

// DownloadInternal handles GET /api/v1/invoices/internal/:id/pdf.
// Responds with the invoice rendered as application/pdf.
func (h *PDFHandler) DownloadInternal(c echo.Context) error {
	tenantID, err := getTenantID(c)
	if err != nil {
		return WriteProblem(c, http.StatusForbidden, "forbidden", "tenant context missing")
	}
	invoiceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return WriteProblem(c, http.StatusBadRequest, "bad-request", "invalid invoice id")
	}

	inv, err := h.internalSvc.GetByID(c.Request().Context(), tenantID, invoiceID)
	if err != nil {
		return WriteProblem(c, http.StatusNotFound, "not-found", "invoice not found")
	}

	tenant, err := h.tenantInfo(c, tenantID)
	if err != nil {
		return WriteProblem(c, http.StatusInternalServerError, "internal-error", "failed to load tenant info")
	}

	data, err := invoicepdf.BuildInternalInvoicePDF(tenant, inv)
	if err != nil {
		return WriteProblem(c, http.StatusInternalServerError, "internal-error", "failed to generate PDF")
	}

	filename := fmt.Sprintf("factura-%s.pdf", inv.Correlative)
	if inv.Correlative == "" {
		filename = fmt.Sprintf("factura-%s.pdf", inv.ID.String()[:8])
	}
	return servePDF(c, data, filename)
}

// DownloadFiscal handles GET /api/v1/invoices/fiscal/:id/pdf.
// Responds with the fiscal invoice rendered as application/pdf.
func (h *PDFHandler) DownloadFiscal(c echo.Context) error {
	tenantID, err := getTenantID(c)
	if err != nil {
		return WriteProblem(c, http.StatusForbidden, "forbidden", "tenant context missing")
	}
	invoiceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return WriteProblem(c, http.StatusBadRequest, "bad-request", "invalid invoice id")
	}

	inv, err := h.fiscalSvc.GetByID(c.Request().Context(), tenantID, invoiceID)
	if err != nil {
		return WriteProblem(c, http.StatusNotFound, "not-found", "fiscal invoice not found")
	}

	tenant, err := h.tenantInfo(c, tenantID)
	if err != nil {
		return WriteProblem(c, http.StatusInternalServerError, "internal-error", "failed to load tenant info")
	}

	data, err := invoicepdf.BuildFiscalInvoicePDF(tenant, inv)
	if err != nil {
		return WriteProblem(c, http.StatusInternalServerError, "internal-error", "failed to generate PDF")
	}

	filename := "factura-fiscal"
	if inv.FiscalNumber != nil {
		filename = fmt.Sprintf("factura-fiscal-%s", *inv.FiscalNumber)
	}
	return servePDF(c, data, filename+".pdf")
}

// ─── Private helpers ──────────────────────────────────────────────────────────

func (h *PDFHandler) tenantInfo(c echo.Context, tenantID uuid.UUID) (invoicepdf.TenantInfo, error) {
	var info invoicepdf.TenantInfo
	var fiscalID *string

	err := h.pool.QueryRow(c.Request().Context(),
		`SELECT name, fiscal_id, country_code FROM tenants WHERE id = $1`,
		tenantID,
	).Scan(&info.Name, &fiscalID, &info.Country)
	if err != nil {
		if err == pgx.ErrNoRows {
			return info, fmt.Errorf("tenant not found")
		}
		return info, fmt.Errorf("query tenant: %w", err)
	}
	if fiscalID != nil {
		info.FiscalID = *fiscalID
	}
	return info, nil
}

func servePDF(c echo.Context, data []byte, filename string) error {
	c.Response().Header().Set("Content-Type", "application/pdf")
	c.Response().Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Response().Header().Set("Cache-Control", "private, no-store")
	return c.Blob(http.StatusOK, "application/pdf", data)
}
