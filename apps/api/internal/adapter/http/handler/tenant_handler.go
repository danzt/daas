// Package handler contains Echo HTTP handlers. Each handler is a thin
// adapter between HTTP and the application service layer. No business logic
// lives here — handlers parse requests, call services, and format responses.
package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"

	"github.com/danzt/daas/api/internal/app"
	"github.com/danzt/daas/api/internal/domain/tenant"
)

// TenantHandler handles tenant registration and management endpoints.
type TenantHandler struct {
	service *app.TenantService
}

// NewTenantHandler creates a TenantHandler backed by the given service.
func NewTenantHandler(service *app.TenantService) *TenantHandler {
	return &TenantHandler{service: service}
}

// registerRequest is the JSON body for POST /api/v1/auth/register.
type registerRequest struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	CountryCode string `json:"country_code"`
	FiscalID    string `json:"fiscal_id,omitempty"`
}

// registerResponse is returned on successful registration (HTTP 201).
type registerResponse struct {
	TenantID string `json:"tenant_id"`
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
}

// Register handles POST /api/v1/auth/register.
// Creates a new tenant + owner user. Returns 201 on success.
//
// Error responses use RFC 7807 Problem Details format.
func (h *TenantHandler) Register(c echo.Context) error {
	var req registerRequest
	if err := c.Bind(&req); err != nil {
		return WriteProblem(c, http.StatusBadRequest, "bad-request", "invalid JSON body")
	}

	result, err := h.service.Register(c.Request().Context(), app.RegisterRequest{
		Name:        req.Name,
		Email:       req.Email,
		Password:    req.Password,
		CountryCode: req.CountryCode,
		FiscalID:    req.FiscalID,
	})
	if err != nil {
		return MapError(c, err)
	}

	return c.JSON(http.StatusCreated, registerResponse{
		TenantID: result.TenantID.String(),
		UserID:   result.UserID.String(),
		Email:    result.Email,
	})
}

// MapError translates domain/service errors to RFC 7807 HTTP responses.
// Exported so tests can reuse the same mapping logic.
func MapError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, tenant.ErrDuplicateEmail):
		return WriteProblem(c, http.StatusConflict, "duplicate-email", "email already registered")
	case errors.Is(err, tenant.ErrInvalidCountryCode):
		return WriteProblem(c, http.StatusUnprocessableEntity, "invalid-country-code",
			"country_code must be 2 uppercase letters (ISO-3166 alpha-2)")
	case errors.Is(err, tenant.ErrTenantNotFound):
		return WriteProblem(c, http.StatusNotFound, "tenant-not-found", "tenant not found")
	case errors.Is(err, tenant.ErrTenantSuspended):
		return WriteProblem(c, http.StatusForbidden, "tenant-suspended", "tenant account is suspended")
	default:
		log.Error().Err(err).Str("path", c.Request().URL.Path).Msg("unhandled service error")
		return WriteProblem(c, http.StatusInternalServerError, "internal-error", "an unexpected error occurred")
	}
}

// WriteProblem writes an RFC 7807 Problem Details JSON response.
// Exported so tests and other handlers can produce consistent error shapes.
func WriteProblem(c echo.Context, status int, errType, detail string) error {
	return c.JSON(status, map[string]any{
		"type":   "https://daas.app/errors/" + errType,
		"title":  http.StatusText(status),
		"status": status,
		"detail": detail,
	})
}
