package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/labstack/echo/v4"
)

// AuthHandler proxies authentication requests to Supabase Auth.
// The Go API is stateless — no server-side sessions. Clients hold their JWT.
type AuthHandler struct {
	supabaseURL string
	anonKey     string
	httpClient  *http.Client
}

// NewAuthHandler creates an AuthHandler that proxies auth requests to the
// given Supabase project URL.
func NewAuthHandler(supabaseURL, anonKey string) *AuthHandler {
	return &AuthHandler{
		supabaseURL: supabaseURL,
		anonKey:     anonKey,
		httpClient:  &http.Client{Timeout: 15 * time.Second},
	}
}

// loginRequest is the JSON body for POST /api/v1/auth/login.
type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Login handles POST /api/v1/auth/login.
// Proxies to Supabase: POST /auth/v1/token?grant_type=password.
// On success returns the Supabase JWT payload (access_token, refresh_token).
// On invalid credentials returns HTTP 401 with RFC 7807 Problem Details.
func (h *AuthHandler) Login(c echo.Context) error {
	var req loginRequest
	if err := c.Bind(&req); err != nil {
		return WriteProblem(c, http.StatusBadRequest, "bad-request", "invalid JSON body")
	}
	if req.Email == "" || req.Password == "" {
		return WriteProblem(c, http.StatusUnprocessableEntity, "validation-error",
			"email and password are required")
	}

	body, err := json.Marshal(map[string]string{
		"email":    req.Email,
		"password": req.Password,
	})
	if err != nil {
		return WriteProblem(c, http.StatusInternalServerError, "internal-error",
			"failed to build login request")
	}

	url := h.supabaseURL + "/auth/v1/token?grant_type=password"
	proxyReq, err := http.NewRequestWithContext(c.Request().Context(), http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return WriteProblem(c, http.StatusInternalServerError, "internal-error",
			"failed to create proxy request")
	}
	proxyReq.Header.Set("Content-Type", "application/json")
	proxyReq.Header.Set("apikey", h.anonKey)

	resp, err := h.httpClient.Do(proxyReq)
	if err != nil {
		return WriteProblem(c, http.StatusBadGateway, "auth-unavailable",
			"authentication service unavailable")
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return WriteProblem(c, http.StatusInternalServerError, "internal-error",
			"failed to read auth response")
	}

	// Supabase returns 400 for invalid credentials.
	if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusUnauthorized {
		return WriteProblem(c, http.StatusUnauthorized, "invalid-credentials",
			"invalid email or password")
	}
	if resp.StatusCode != http.StatusOK {
		return WriteProblem(c, http.StatusInternalServerError, "internal-error",
			"unexpected response from auth service")
	}

	// Forward the Supabase JWT payload directly to the client.
	// The client is responsible for storing access_token and refresh_token.
	c.Response().Header().Set("Content-Type", "application/json")
	c.Response().WriteHeader(http.StatusOK)
	_, err = c.Response().Write(respBody)
	return err
}

// Logout handles POST /api/v1/auth/logout.
// Calls Supabase POST /auth/v1/logout to invalidate the session server-side.
// The client must also clear its local token storage.
func (h *AuthHandler) Logout(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return WriteProblem(c, http.StatusUnauthorized, "unauthorized",
			"authorization header required for logout")
	}

	url := h.supabaseURL + "/auth/v1/logout"
	proxyReq, err := http.NewRequestWithContext(c.Request().Context(), http.MethodPost, url, nil)
	if err != nil {
		return WriteProblem(c, http.StatusInternalServerError, "internal-error",
			"failed to create logout request")
	}
	proxyReq.Header.Set("Authorization", authHeader)

	resp, err := h.httpClient.Do(proxyReq)
	if err != nil {
		return WriteProblem(c, http.StatusBadGateway, "auth-unavailable",
			"authentication service unavailable")
	}
	defer func() { _ = resp.Body.Close() }()

	return c.JSON(http.StatusOK, map[string]string{"message": "logged out"})
}

// recoverRequest is the JSON body for POST /api/v1/auth/recover.
type recoverRequest struct {
	Email      string `json:"email"`
	RedirectTo string `json:"redirect_to"`
}

// RequestPasswordReset handles POST /api/v1/auth/recover.
// Proxies to Supabase POST /auth/v1/recover, which emails a recovery link.
// The optional redirect_to is validated by Supabase against its allow list, so
// forwarding it from the client is safe (a non-allowed URL falls back to the
// Site URL). Always returns 200 to avoid revealing which emails are registered.
func (h *AuthHandler) RequestPasswordReset(c echo.Context) error {
	var req recoverRequest
	if err := c.Bind(&req); err != nil {
		return WriteProblem(c, http.StatusBadRequest, "bad-request", "invalid JSON body")
	}
	if req.Email == "" {
		return WriteProblem(c, http.StatusUnprocessableEntity, "validation-error",
			"email is required")
	}

	body, err := json.Marshal(map[string]string{"email": req.Email})
	if err != nil {
		return WriteProblem(c, http.StatusInternalServerError, "internal-error",
			"failed to build recover request")
	}

	recoverURL := h.supabaseURL + "/auth/v1/recover"
	if req.RedirectTo != "" {
		recoverURL += "?redirect_to=" + url.QueryEscape(req.RedirectTo)
	}
	proxyReq, err := http.NewRequestWithContext(c.Request().Context(), http.MethodPost, recoverURL, bytes.NewReader(body))
	if err != nil {
		return WriteProblem(c, http.StatusInternalServerError, "internal-error",
			"failed to create recover request")
	}
	proxyReq.Header.Set("Content-Type", "application/json")
	proxyReq.Header.Set("apikey", h.anonKey)

	resp, err := h.httpClient.Do(proxyReq)
	if err != nil {
		return WriteProblem(c, http.StatusBadGateway, "auth-unavailable",
			"authentication service unavailable")
	}
	defer func() { _ = resp.Body.Close() }()

	// Always 200 regardless of Supabase's response — never leak whether the
	// email is registered.
	return c.JSON(http.StatusOK, map[string]string{
		"message": "if the email is registered, a recovery link was sent",
	})
}

// updatePasswordRequest is the JSON body for PUT /api/v1/auth/password.
type updatePasswordRequest struct {
	Password string `json:"password"`
}

// UpdatePassword handles PUT /api/v1/auth/password.
// Requires the recovery access_token (from the email link) in the Authorization
// header. Proxies to Supabase PUT /auth/v1/user to set the new password.
func (h *AuthHandler) UpdatePassword(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return WriteProblem(c, http.StatusUnauthorized, "unauthorized",
			"authorization header required")
	}

	var req updatePasswordRequest
	if err := c.Bind(&req); err != nil {
		return WriteProblem(c, http.StatusBadRequest, "bad-request", "invalid JSON body")
	}
	if len(req.Password) < 8 {
		return WriteProblem(c, http.StatusUnprocessableEntity, "validation-error",
			"password must be at least 8 characters")
	}

	body, err := json.Marshal(map[string]string{"password": req.Password})
	if err != nil {
		return WriteProblem(c, http.StatusInternalServerError, "internal-error",
			"failed to build update request")
	}

	proxyReq, err := http.NewRequestWithContext(c.Request().Context(), http.MethodPut,
		h.supabaseURL+"/auth/v1/user", bytes.NewReader(body))
	if err != nil {
		return WriteProblem(c, http.StatusInternalServerError, "internal-error",
			"failed to create update request")
	}
	proxyReq.Header.Set("Content-Type", "application/json")
	proxyReq.Header.Set("apikey", h.anonKey)
	proxyReq.Header.Set("Authorization", authHeader)

	resp, err := h.httpClient.Do(proxyReq)
	if err != nil {
		return WriteProblem(c, http.StatusBadGateway, "auth-unavailable",
			"authentication service unavailable")
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return WriteProblem(c, http.StatusUnauthorized, "invalid-token",
			"recovery link is invalid or has expired")
	}
	if resp.StatusCode != http.StatusOK {
		return WriteProblem(c, http.StatusInternalServerError, "internal-error",
			"failed to update password")
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "password updated"})
}
