package handler

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"

	"github.com/danzt/daas/api/internal/adapter/storage"
)

// BrandingHandler exposes HTTP endpoints for tenant storefront branding.
// It holds a DB pool for direct queries (no service layer needed — branding is
// a simple read/write on a single tenants row) and an optional storage adapter
// for logo/banner image uploads.
type BrandingHandler struct {
	pool    *pgxpool.Pool
	storage storage.Storage
}

// NewBrandingHandler creates a BrandingHandler backed by the given pool and
// storage adapter. Both are required for the full feature set; nil storage
// disables image upload/delete endpoints.
func NewBrandingHandler(pool *pgxpool.Pool, st storage.Storage) *BrandingHandler {
	return &BrandingHandler{pool: pool, storage: st}
}

// ─── DTOs ─────────────────────────────────────────────────────────────────────

// BrandingResponse is the public-facing DTO returned by GetPublic and GetAdmin.
type BrandingResponse struct {
	StoreName             string `json:"store_name"`
	Tagline               string `json:"tagline,omitempty"`
	LogoURL               string `json:"logo_url,omitempty"`
	BannerURL             string `json:"banner_url,omitempty"`
	PrimaryColor          string `json:"primary_color,omitempty"`
	NotificationsWhatsApp string `json:"notifications_whatsapp_phone,omitempty"`
}

type updateBrandingReq struct {
	StoreName             *string `json:"store_name"`
	Tagline               *string `json:"tagline"`
	PrimaryColor          *string `json:"primary_color"`
	NotificationsWhatsApp *string `json:"notifications_whatsapp_phone"`
}

// ─── Public endpoint ──────────────────────────────────────────────────────────

// GetPublic handles GET /t/:tenantSlug/shop/v1/branding
//
// No auth required. Returns the branding fields for the resolved tenant.
// Uses the tenant_id already placed in context by PublicTenantMiddleware.
func (h *BrandingHandler) GetPublic(c echo.Context) error {
	tenantID, err := getPublicTenantID(c)
	if err != nil {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "tenant_context_missing"})
	}

	row := h.pool.QueryRow(c.Request().Context(),
		`SELECT name, branding_store_name, branding_tagline,
		        branding_logo_url, branding_banner_url, branding_primary_color
		 FROM tenants WHERE id = $1`, tenantID)

	var name, storeName, tagline, logoURL, bannerURL, primaryColor *string
	if err := row.Scan(&name, &storeName, &tagline, &logoURL, &bannerURL, &primaryColor); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal_error"})
	}

	resp := BrandingResponse{}
	if storeName != nil && *storeName != "" {
		resp.StoreName = *storeName
	} else if name != nil {
		resp.StoreName = *name
	}
	if tagline != nil {
		resp.Tagline = *tagline
	}
	if logoURL != nil {
		resp.LogoURL = *logoURL
	}
	if bannerURL != nil {
		resp.BannerURL = *bannerURL
	}
	if primaryColor != nil {
		resp.PrimaryColor = *primaryColor
	}
	return c.JSON(http.StatusOK, resp)
}

// ─── Admin endpoints ──────────────────────────────────────────────────────────

// GetAdmin handles GET /api/v1/branding
//
// Returns the full branding configuration for the authenticated tenant,
// including the WhatsApp notification phone number.
func (h *BrandingHandler) GetAdmin(c echo.Context) error {
	tenantID, err := getTenantID(c)
	if err != nil {
		return WriteProblem(c, http.StatusForbidden, "forbidden", "tenant context missing")
	}

	row := h.pool.QueryRow(c.Request().Context(),
		`SELECT name, branding_store_name, branding_tagline,
		        branding_logo_url, branding_banner_url, branding_primary_color,
		        notifications_whatsapp_phone
		 FROM tenants WHERE id = $1`, tenantID)

	var name, storeName, tagline, logoURL, bannerURL, primaryColor, waPhone *string
	if err := row.Scan(&name, &storeName, &tagline, &logoURL, &bannerURL, &primaryColor, &waPhone); err != nil {
		return WriteProblem(c, http.StatusInternalServerError, "internal-error", "failed to load branding")
	}

	resp := BrandingResponse{}
	if storeName != nil && *storeName != "" {
		resp.StoreName = *storeName
	} else if name != nil {
		resp.StoreName = *name
	}
	if tagline != nil {
		resp.Tagline = *tagline
	}
	if logoURL != nil {
		resp.LogoURL = *logoURL
	}
	if bannerURL != nil {
		resp.BannerURL = *bannerURL
	}
	if primaryColor != nil {
		resp.PrimaryColor = *primaryColor
	}
	if waPhone != nil {
		resp.NotificationsWhatsApp = *waPhone
	}
	return c.JSON(http.StatusOK, resp)
}

// Update handles PUT /api/v1/branding
//
// Updates text-only branding fields: store_name, tagline, primary_color.
// Image fields (logo/banner) are managed via dedicated upload/delete endpoints.
func (h *BrandingHandler) Update(c echo.Context) error {
	tenantID, err := getTenantID(c)
	if err != nil {
		return WriteProblem(c, http.StatusForbidden, "forbidden", "tenant context missing")
	}

	var req updateBrandingReq
	if err := c.Bind(&req); err != nil {
		return WriteProblem(c, http.StatusBadRequest, "bad-request", "invalid JSON body")
	}

	// Validate primary_color if provided.
	if req.PrimaryColor != nil && *req.PrimaryColor != "" {
		if !isValidHexColor(*req.PrimaryColor) {
			return WriteProblem(c, http.StatusUnprocessableEntity, "validation-error",
				"primary_color must be a 6-character hex color starting with # (e.g. #7C3AED)")
		}
	}

	_, err = h.pool.Exec(c.Request().Context(),
		`UPDATE tenants
		 SET branding_store_name             = COALESCE($2, branding_store_name),
		     branding_tagline                = COALESCE($3, branding_tagline),
		     branding_primary_color          = COALESCE($4, branding_primary_color),
		     notifications_whatsapp_phone    = COALESCE($5, notifications_whatsapp_phone),
		     updated_at                      = NOW()
		 WHERE id = $1`,
		tenantID,
		nullableString(req.StoreName),
		nullableString(req.Tagline),
		nullableString(req.PrimaryColor),
		nullableString(req.NotificationsWhatsApp),
	)
	if err != nil {
		return WriteProblem(c, http.StatusInternalServerError, "internal-error", "failed to update branding")
	}

	return h.GetAdmin(c)
}

// ─── Logo endpoints ───────────────────────────────────────────────────────────

// UploadLogo handles POST /api/v1/branding/logo
//
// Multipart field "logo" (required). Accepts image/* content types. Max 2 MB.
func (h *BrandingHandler) UploadLogo(c echo.Context) error {
	if h.storage == nil {
		return WriteProblem(c, http.StatusInternalServerError, "storage-unavailable", "storage adapter not configured")
	}
	return h.uploadBrandingImage(c, "logo", "tenant-branding/%s/logo/%d%s", 2*1024*1024,
		func(tenantID, url string) error {
			_, err := h.pool.Exec(c.Request().Context(),
				`UPDATE tenants SET branding_logo_url = $2, updated_at = NOW() WHERE id = $1`,
				tenantID, url)
			return err
		})
}

// DeleteLogo handles DELETE /api/v1/branding/logo
//
// Clears the logo URL and deletes the stored file. Idempotent.
func (h *BrandingHandler) DeleteLogo(c echo.Context) error {
	if h.storage == nil {
		return WriteProblem(c, http.StatusInternalServerError, "storage-unavailable", "storage adapter not configured")
	}
	return h.deleteBrandingImage(c, "logo", func(tenantID string) (string, error) {
		var url *string
		err := h.pool.QueryRow(c.Request().Context(),
			`UPDATE tenants SET branding_logo_url = NULL, updated_at = NOW()
			 WHERE id = $1 RETURNING branding_logo_url`, tenantID).Scan(&url)
		if url != nil {
			return *url, err
		}
		return "", err
	})
}

// ─── Banner endpoints ─────────────────────────────────────────────────────────

// UploadBanner handles POST /api/v1/branding/banner
//
// Multipart field "banner" (required). Accepts image/* content types. Max 5 MB.
func (h *BrandingHandler) UploadBanner(c echo.Context) error {
	if h.storage == nil {
		return WriteProblem(c, http.StatusInternalServerError, "storage-unavailable", "storage adapter not configured")
	}
	return h.uploadBrandingImage(c, "banner", "tenant-branding/%s/banner/%d%s", 5*1024*1024,
		func(tenantID, url string) error {
			_, err := h.pool.Exec(c.Request().Context(),
				`UPDATE tenants SET branding_banner_url = $2, updated_at = NOW() WHERE id = $1`,
				tenantID, url)
			return err
		})
}

// DeleteBanner handles DELETE /api/v1/branding/banner
//
// Clears the banner URL and deletes the stored file. Idempotent.
func (h *BrandingHandler) DeleteBanner(c echo.Context) error {
	if h.storage == nil {
		return WriteProblem(c, http.StatusInternalServerError, "storage-unavailable", "storage adapter not configured")
	}
	return h.deleteBrandingImage(c, "banner", func(tenantID string) (string, error) {
		var url *string
		err := h.pool.QueryRow(c.Request().Context(),
			`UPDATE tenants SET branding_banner_url = NULL, updated_at = NOW()
			 WHERE id = $1 RETURNING branding_banner_url`, tenantID).Scan(&url)
		if url != nil {
			return *url, err
		}
		return "", err
	})
}

// ─── Shared image helpers ─────────────────────────────────────────────────────

// uploadBrandingImage is a shared upload handler for logo and banner.
// formField is the multipart field name ("logo" or "banner").
// keyFmt is a fmt.Sprintf pattern with %s=tenantID, %d=timestamp, %s=ext.
// maxSize is the byte limit.
// save is called with (tenantID, publicURL) to persist the URL in the DB.
func (h *BrandingHandler) uploadBrandingImage(
	c echo.Context,
	formField string,
	keyFmt string,
	maxSize int64,
	save func(tenantID, url string) error,
) error {
	tenantID, err := getTenantID(c)
	if err != nil {
		return WriteProblem(c, http.StatusForbidden, "forbidden", "tenant context missing")
	}

	fileHeader, err := c.FormFile(formField)
	if err != nil {
		return WriteProblem(c, http.StatusBadRequest, formField+"-required",
			fmt.Sprintf("multipart field '%s' is required", formField))
	}

	// Validate content-type — accept any image/* MIME type.
	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	if idx := strings.Index(contentType, ";"); idx != -1 {
		contentType = strings.TrimSpace(contentType[:idx])
	}
	if !strings.HasPrefix(contentType, "image/") {
		return WriteProblem(c, http.StatusUnsupportedMediaType, "unsupported-media-type",
			"file must be an image (image/jpeg, image/png, image/webp, etc.)")
	}

	// Validate size.
	if fileHeader.Size > maxSize {
		return WriteProblem(c, http.StatusRequestEntityTooLarge, "file-too-large",
			fmt.Sprintf("file must not exceed %d MB", maxSize/1024/1024))
	}

	// Derive extension from filename or content-type.
	ext := filepath.Ext(fileHeader.Filename)
	if ext == "" {
		switch contentType {
		case "image/jpeg":
			ext = ".jpg"
		case "image/png":
			ext = ".png"
		case "image/webp":
			ext = ".webp"
		default:
			ext = ".img"
		}
	}

	file, err := fileHeader.Open()
	if err != nil {
		return WriteProblem(c, http.StatusInternalServerError, "internal-error", "cannot open uploaded file")
	}
	defer func() { _ = file.Close() }()

	storageKey := fmt.Sprintf(keyFmt, tenantID, time.Now().UnixMilli(), ext)
	publicURL, err := h.storage.Save(c.Request().Context(), storageKey, file, contentType)
	if err != nil {
		return WriteProblem(c, http.StatusInternalServerError, "internal-error", "failed to store image")
	}

	if err := save(tenantID.String(), publicURL); err != nil {
		return WriteProblem(c, http.StatusInternalServerError, "internal-error", "failed to update branding record")
	}

	return h.GetAdmin(c)
}

// deleteBrandingImage is a shared delete handler for logo and banner.
// getAndClear fetches the current URL and atomically clears it in the DB.
func (h *BrandingHandler) deleteBrandingImage(
	c echo.Context,
	_ string,
	getAndClear func(tenantID string) (string, error),
) error {
	tenantID, err := getTenantID(c)
	if err != nil {
		return WriteProblem(c, http.StatusForbidden, "forbidden", "tenant context missing")
	}

	oldURL, err := getAndClear(tenantID.String())
	if err != nil {
		return WriteProblem(c, http.StatusInternalServerError, "internal-error", "failed to clear branding image")
	}

	// Best-effort delete from storage.
	if oldURL != "" {
		key := extractStorageKey(oldURL)
		if key != "" {
			_ = h.storage.Delete(c.Request().Context(), key)
		}
	}

	return h.GetAdmin(c)
}

// ─── Utility helpers ──────────────────────────────────────────────────────────

// isValidHexColor returns true when s is a valid 6-character hex color code
// starting with '#' (e.g. "#7C3AED").
func isValidHexColor(s string) bool {
	if len(s) != 7 || s[0] != '#' {
		return false
	}
	for _, r := range s[1:] {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') && (r < 'A' || r > 'F') {
			return false
		}
	}
	return true
}

// nullableString converts a *string pointer to an interface that pgx can use
// as a nullable parameter. A nil pointer is passed as nil (SQL NULL).
func nullableString(s *string) interface{} {
	if s == nil {
		return nil
	}
	return *s
}
