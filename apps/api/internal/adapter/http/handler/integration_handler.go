package handler

import (
	"encoding/base64"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"

	mw "github.com/danzt/daas/api/internal/adapter/http/middleware"
	"github.com/danzt/daas/api/internal/platform/crypto"
)

// IntegrationHandler manages tenant fiscal integration configuration.
// API keys are stored AES-256-GCM encrypted. The plaintext is NEVER returned
// in API responses — only a masked version (last 4 characters).
type IntegrationHandler struct {
	pool *pgxpool.Pool
}

// NewIntegrationHandler creates an IntegrationHandler.
func NewIntegrationHandler(pool *pgxpool.Pool) *IntegrationHandler {
	return &IntegrationHandler{pool: pool}
}

// upsertFiscalRequest is the JSON body for PUT /api/v1/tenant/integrations/fiscal.
type upsertFiscalRequest struct {
	APIKey string `json:"api_key"`
}

// UpsertFiscal handles PUT /api/v1/tenant/integrations/fiscal.
// Encrypts the fiscal API key and stores (or updates) it in the DB.
// Returns the masked key. Requires owner role.
func (h *IntegrationHandler) UpsertFiscal(c echo.Context) error {
	tenantID, _ := c.Get(string(mw.ContextKeyTenantID)).(string)
	if tenantID == "" {
		return WriteProblem(c, http.StatusForbidden, "forbidden", "tenant context missing")
	}

	var req upsertFiscalRequest
	if err := c.Bind(&req); err != nil {
		return WriteProblem(c, http.StatusBadRequest, "bad-request", "invalid JSON body")
	}
	if req.APIKey == "" {
		return WriteProblem(c, http.StatusUnprocessableEntity, "validation-error", "api_key is required")
	}

	encKey, err := loadEncryptionKey()
	if err != nil {
		return WriteProblem(c, http.StatusInternalServerError, "configuration-error",
			"encryption key not configured")
	}

	encrypted, err := crypto.Encrypt([]byte(req.APIKey), encKey)
	if err != nil {
		return WriteProblem(c, http.StatusInternalServerError, "internal-error", "failed to encrypt api key")
	}

	_, err = h.pool.Exec(c.Request().Context(),
		`INSERT INTO tenant_integrations (tenant_id, integration_type, config_encrypted, active)
         VALUES ($1, 'fiscal', $2, TRUE)
         ON CONFLICT (tenant_id, integration_type)
         DO UPDATE SET config_encrypted = $2, active = TRUE, updated_at = NOW()`,
		tenantID, encrypted,
	)
	if err != nil {
		return WriteProblem(c, http.StatusInternalServerError, "internal-error",
			"failed to store integration config")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"integration_type": "fiscal",
		"active":           true,
		"api_key_masked":   maskKey(req.APIKey),
		"message":          "fiscal integration configured successfully",
	})
}

// GetFiscal handles GET /api/v1/tenant/integrations/fiscal.
// Returns integration status and masked API key (last 4 chars only).
// The plaintext key is NEVER returned — decrypt only happens to generate the mask.
func (h *IntegrationHandler) GetFiscal(c echo.Context) error {
	tenantID, _ := c.Get(string(mw.ContextKeyTenantID)).(string)
	if tenantID == "" {
		return WriteProblem(c, http.StatusForbidden, "forbidden", "tenant context missing")
	}

	var encrypted []byte
	var active bool
	err := h.pool.QueryRow(c.Request().Context(),
		`SELECT config_encrypted, active
         FROM tenant_integrations
         WHERE tenant_id = $1 AND integration_type = 'fiscal'`,
		tenantID,
	).Scan(&encrypted, &active)
	if err != nil {
		if err == pgx.ErrNoRows {
			return c.JSON(http.StatusOK, map[string]interface{}{
				"integration_type": "fiscal",
				"active":           false,
				"configured":       false,
			})
		}
		return WriteProblem(c, http.StatusInternalServerError, "internal-error",
			"failed to fetch integration config")
	}

	// Decrypt only to generate the masked display value.
	// The plaintext is NOT included in the response.
	maskedKey := ""
	if len(encrypted) > 0 {
		encKey, err := loadEncryptionKey()
		if err == nil {
			if plaintext, err := crypto.Decrypt(encrypted, encKey); err == nil {
				maskedKey = maskKey(string(plaintext))
			}
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"integration_type": "fiscal",
		"active":           active,
		"configured":       len(encrypted) > 0,
		"api_key_masked":   maskedKey,
	})
}

// loadEncryptionKey reads INTEGRATION_ENCRYPTION_KEY from the environment.
// The value must be a base64-encoded 32-byte key.
func loadEncryptionKey() ([]byte, error) {
	raw := os.Getenv("INTEGRATION_ENCRYPTION_KEY")
	if raw == "" {
		return nil, WriteProblemErr("INTEGRATION_ENCRYPTION_KEY environment variable not set")
	}
	key, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil, WriteProblemErr("failed to decode INTEGRATION_ENCRYPTION_KEY: " + err.Error())
	}
	if len(key) != 32 {
		return nil, WriteProblemErr("INTEGRATION_ENCRYPTION_KEY must decode to exactly 32 bytes")
	}
	return key, nil
}

// WriteProblemErr is a helper that creates a plain error from a message.
// Used to propagate configuration errors without the echo.Context.
func WriteProblemErr(msg string) error {
	return errStr(msg)
}

type errStr string

func (e errStr) Error() string { return string(e) }

// maskKey returns a masked version of the key showing only the last 4 characters.
// This is safe to return in API responses for verification without exposing the secret.
func maskKey(key string) string {
	if len(key) <= 4 {
		return "****"
	}
	masked := ""
	for range len(key) - 4 {
		masked += "*"
	}
	return masked + key[len(key)-4:]
}
