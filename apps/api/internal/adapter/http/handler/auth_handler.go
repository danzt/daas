package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
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
