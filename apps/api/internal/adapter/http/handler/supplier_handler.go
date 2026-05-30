package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"

	"github.com/danzt/daas/api/internal/app"
	"github.com/danzt/daas/api/internal/domain/supplier"
)

// SupplierHandler exposes HTTP endpoints for supplier management and purchase orders.
type SupplierHandler struct {
	svc *app.SupplierService
}

func NewSupplierHandler(pool *pgxpool.Pool) *SupplierHandler {
	return &SupplierHandler{svc: app.NewSupplierService(pool)}
}

// ─── Request DTOs ─────────────────────────────────────────────────────────────

type createSupplierReq struct {
	Name        string `json:"name"`
	RIF         string `json:"rif"`
	ContactName string `json:"contact_name"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Address     string `json:"address"`
	Notes       string `json:"notes"`
}

type updateSupplierReq struct {
	Name        *string `json:"name"`
	RIF         *string `json:"rif"`
	ContactName *string `json:"contact_name"`
	Email       *string `json:"email"`
	Phone       *string `json:"phone"`
	Address     *string `json:"address"`
	Notes       *string `json:"notes"`
	Active      *bool   `json:"active"`
}

type createPOLineReq struct {
	ProductID   string  `json:"product_id"`
	Description string  `json:"description"`
	Quantity    float64 `json:"quantity"`
	UnitCost    float64 `json:"unit_cost"`
}

type createPOReq struct {
	SupplierID string            `json:"supplier_id"`
	Notes      string            `json:"notes"`
	Lines      []createPOLineReq `json:"lines"`
}

// ─── Supplier handlers ────────────────────────────────────────────────────────

func (h *SupplierHandler) CreateSupplier(c echo.Context) error {
	tenantID, err := getTenantID(c)
	if err != nil {
		return WriteProblem(c, http.StatusForbidden, "forbidden", "tenant context missing")
	}
	var body createSupplierReq
	if err := c.Bind(&body); err != nil {
		return WriteProblem(c, http.StatusBadRequest, "bad-request", "invalid body")
	}
	sup, err := h.svc.CreateSupplier(c.Request().Context(), tenantID, supplier.CreateSupplierRequest{
		Name: body.Name, RIF: body.RIF, ContactName: body.ContactName,
		Email: body.Email, Phone: body.Phone, Address: body.Address, Notes: body.Notes,
	})
	if err != nil {
		return mapSupplierError(c, err)
	}
	return c.JSON(http.StatusCreated, sup)
}

func (h *SupplierHandler) ListSuppliers(c echo.Context) error {
	tenantID, err := getTenantID(c)
	if err != nil {
		return WriteProblem(c, http.StatusForbidden, "forbidden", "tenant context missing")
	}
	activeOnly := c.QueryParam("active") != "false"
	suppliers, err := h.svc.ListSuppliers(c.Request().Context(), tenantID, activeOnly)
	if err != nil {
		return WriteProblem(c, http.StatusInternalServerError, "internal-error", err.Error())
	}
	if suppliers == nil {
		suppliers = []*supplier.Supplier{}
	}
	return c.JSON(http.StatusOK, suppliers)
}

func (h *SupplierHandler) GetSupplier(c echo.Context) error {
	tenantID, err := getTenantID(c)
	if err != nil {
		return WriteProblem(c, http.StatusForbidden, "forbidden", "tenant context missing")
	}
	supplierID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return WriteProblem(c, http.StatusBadRequest, "bad-request", "invalid supplier id")
	}
	sup, err := h.svc.GetSupplier(c.Request().Context(), tenantID, supplierID)
	if err != nil {
		return mapSupplierError(c, err)
	}
	return c.JSON(http.StatusOK, sup)
}

func (h *SupplierHandler) UpdateSupplier(c echo.Context) error {
	tenantID, err := getTenantID(c)
	if err != nil {
		return WriteProblem(c, http.StatusForbidden, "forbidden", "tenant context missing")
	}
	supplierID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return WriteProblem(c, http.StatusBadRequest, "bad-request", "invalid supplier id")
	}
	var body updateSupplierReq
	if err := c.Bind(&body); err != nil {
		return WriteProblem(c, http.StatusBadRequest, "bad-request", "invalid body")
	}
	sup, err := h.svc.UpdateSupplier(c.Request().Context(), tenantID, supplierID, supplier.UpdateSupplierRequest{
		Name: body.Name, RIF: body.RIF, ContactName: body.ContactName,
		Email: body.Email, Phone: body.Phone, Address: body.Address,
		Notes: body.Notes, Active: body.Active,
	})
	if err != nil {
		return mapSupplierError(c, err)
	}
	return c.JSON(http.StatusOK, sup)
}

// ─── Purchase Order handlers ──────────────────────────────────────────────────

func (h *SupplierHandler) CreatePO(c echo.Context) error {
	tenantID, err := getTenantID(c)
	if err != nil {
		return WriteProblem(c, http.StatusForbidden, "forbidden", "tenant context missing")
	}
	supabaseUID, err := getSupabaseUID(c)
	if err != nil {
		return WriteProblem(c, http.StatusUnauthorized, "unauthorized", "missing user context")
	}
	var body createPOReq
	if err := c.Bind(&body); err != nil {
		return WriteProblem(c, http.StatusBadRequest, "bad-request", "invalid body")
	}
	supplierID, err := uuid.Parse(body.SupplierID)
	if err != nil {
		return WriteProblem(c, http.StatusBadRequest, "bad-request", "invalid supplier_id")
	}
	lines, err := buildPOLines(body.Lines)
	if err != nil {
		return WriteProblem(c, http.StatusBadRequest, "bad-request", err.Error())
	}
	po, err := h.svc.CreatePO(c.Request().Context(), tenantID, supabaseUID, supplier.CreatePORequest{
		SupplierID: supplierID, Notes: body.Notes, Lines: lines,
	})
	if err != nil {
		return mapSupplierError(c, err)
	}
	return c.JSON(http.StatusCreated, po)
}

func (h *SupplierHandler) ListPOs(c echo.Context) error {
	tenantID, err := getTenantID(c)
	if err != nil {
		return WriteProblem(c, http.StatusForbidden, "forbidden", "tenant context missing")
	}
	var supplierID *uuid.UUID
	if s := c.QueryParam("supplier_id"); s != "" {
		id, err := uuid.Parse(s)
		if err == nil {
			supplierID = &id
		}
	}
	var status *supplier.POStatus
	if s := c.QueryParam("status"); s != "" {
		st := supplier.POStatus(s)
		status = &st
	}
	pos, err := h.svc.ListPOs(c.Request().Context(), tenantID, supplierID, status)
	if err != nil {
		return WriteProblem(c, http.StatusInternalServerError, "internal-error", err.Error())
	}
	if pos == nil {
		pos = []*supplier.PurchaseOrder{}
	}
	return c.JSON(http.StatusOK, pos)
}

func (h *SupplierHandler) GetPO(c echo.Context) error {
	tenantID, err := getTenantID(c)
	if err != nil {
		return WriteProblem(c, http.StatusForbidden, "forbidden", "tenant context missing")
	}
	poID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return WriteProblem(c, http.StatusBadRequest, "bad-request", "invalid po id")
	}
	po, err := h.svc.GetPO(c.Request().Context(), tenantID, poID)
	if err != nil {
		return mapSupplierError(c, err)
	}
	return c.JSON(http.StatusOK, po)
}

func (h *SupplierHandler) OrderPO(c echo.Context) error {
	tenantID, err := getTenantID(c)
	if err != nil {
		return WriteProblem(c, http.StatusForbidden, "forbidden", "tenant context missing")
	}
	supabaseUID, err := getSupabaseUID(c)
	if err != nil {
		return WriteProblem(c, http.StatusUnauthorized, "unauthorized", "missing user context")
	}
	poID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return WriteProblem(c, http.StatusBadRequest, "bad-request", "invalid po id")
	}
	po, err := h.svc.OrderPO(c.Request().Context(), tenantID, poID, supabaseUID)
	if err != nil {
		return mapSupplierError(c, err)
	}
	return c.JSON(http.StatusOK, po)
}

func (h *SupplierHandler) ReceivePO(c echo.Context) error {
	tenantID, err := getTenantID(c)
	if err != nil {
		return WriteProblem(c, http.StatusForbidden, "forbidden", "tenant context missing")
	}
	supabaseUID, err := getSupabaseUID(c)
	if err != nil {
		return WriteProblem(c, http.StatusUnauthorized, "unauthorized", "missing user context")
	}
	poID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return WriteProblem(c, http.StatusBadRequest, "bad-request", "invalid po id")
	}
	po, err := h.svc.ReceivePO(c.Request().Context(), tenantID, poID, supabaseUID)
	if err != nil {
		return mapSupplierError(c, err)
	}
	return c.JSON(http.StatusOK, po)
}

func (h *SupplierHandler) CancelPO(c echo.Context) error {
	tenantID, err := getTenantID(c)
	if err != nil {
		return WriteProblem(c, http.StatusForbidden, "forbidden", "tenant context missing")
	}
	supabaseUID, err := getSupabaseUID(c)
	if err != nil {
		return WriteProblem(c, http.StatusUnauthorized, "unauthorized", "missing user context")
	}
	poID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return WriteProblem(c, http.StatusBadRequest, "bad-request", "invalid po id")
	}
	po, err := h.svc.CancelPO(c.Request().Context(), tenantID, poID, supabaseUID)
	if err != nil {
		return mapSupplierError(c, err)
	}
	return c.JSON(http.StatusOK, po)
}

// ─── Private helpers ──────────────────────────────────────────────────────────

func buildPOLines(raw []createPOLineReq) ([]supplier.CreatePOLineRequest, error) {
	lines := make([]supplier.CreatePOLineRequest, 0, len(raw))
	for _, r := range raw {
		pid, err := uuid.Parse(r.ProductID)
		if err != nil {
			return nil, fmt.Errorf("invalid product_id: %s", r.ProductID)
		}
		lines = append(lines, supplier.CreatePOLineRequest{
			ProductID: pid, Description: r.Description,
			Quantity: r.Quantity, UnitCost: r.UnitCost,
		})
	}
	return lines, nil
}

func mapSupplierError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, supplier.ErrSupplierNotFound),
		errors.Is(err, supplier.ErrPONotFound):
		return WriteProblem(c, http.StatusNotFound, "not-found", err.Error())
	case errors.Is(err, supplier.ErrSupplierNameRequired),
		errors.Is(err, supplier.ErrEmptyPO),
		errors.Is(err, supplier.ErrPOAlreadyOrdered),
		errors.Is(err, supplier.ErrPOAlreadyReceived),
		errors.Is(err, supplier.ErrPOAlreadyCancelled),
		errors.Is(err, supplier.ErrPONotOrdered),
		errors.Is(err, supplier.ErrPONotDraft):
		return WriteProblem(c, http.StatusUnprocessableEntity, "unprocessable", err.Error())
	default:
		return WriteProblem(c, http.StatusInternalServerError, "internal-error", err.Error())
	}
}
