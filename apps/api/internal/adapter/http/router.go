package http

import (
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/danzt/daas/api/internal/adapter/http/handler"
	mw "github.com/danzt/daas/api/internal/adapter/http/middleware"
	"github.com/danzt/daas/api/internal/app"
)

// RouterConfig holds dependencies needed to build the Echo router.
type RouterConfig struct {
	SupabaseURL    string
	ServiceRoleKey string
	Pool           *pgxpool.Pool
}

// NewRouter creates a new Echo instance with standard middleware configured
// and all routes registered.
func NewRouter() *echo.Echo {
	return NewRouterWithConfig(RouterConfig{
		SupabaseURL:    os.Getenv("SUPABASE_URL"),
		ServiceRoleKey: os.Getenv("SUPABASE_SERVICE_ROLE_KEY"),
	})
}

// NewRouterWithConfig builds the router with explicit dependencies.
// Used in production (from main.go) and in integration tests.
func NewRouterWithConfig(cfg RouterConfig) *echo.Echo { //nolint:funlen
	e := echo.New()
	e.HideBanner = true

	// Structured JSON logger (zerolog) — outputs to stdout
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	// Middleware
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:   true,
		LogURI:      true,
		LogMethod:   true,
		LogLatency:  true,
		LogRemoteIP: true,
		LogError:    true,
		HandleError: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			event := logger.Info()
			if v.Error != nil {
				event = log.Error().Err(v.Error)
			}
			event.
				Str("method", v.Method).
				Str("uri", v.URI).
				Int("status", v.Status).
				Dur("latency", v.Latency).
				Str("remote_ip", v.RemoteIP).
				Time("time", time.Now()).
				Msg("request")
			return nil
		},
	}))
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAuthorization,
		},
	}))

	// Build handlers
	authMW := mw.NewAuthMiddleware(cfg.SupabaseURL)

	tenantSvc := app.NewTenantService(cfg.Pool, cfg.SupabaseURL, cfg.ServiceRoleKey)
	tenantHandler := handler.NewTenantHandler(tenantSvc)

	authHandler := handler.NewAuthHandler(cfg.SupabaseURL)
	userHandler := handler.NewUserHandler(cfg.Pool, cfg.SupabaseURL, cfg.ServiceRoleKey)
	integrationHandler := handler.NewIntegrationHandler(cfg.Pool)
	meHandler := handler.NewMeHandler(cfg.Pool)
	productHandler := handler.NewProductHandler(cfg.Pool)

	// Routes
	registerRoutes(e, authMW, cfg.Pool, tenantHandler, authHandler, userHandler, integrationHandler, meHandler, productHandler)

	return e
}

func registerRoutes(
	e *echo.Echo,
	authMW *mw.AuthMiddleware,
	pool *pgxpool.Pool,
	tenantHandler *handler.TenantHandler,
	authHandler *handler.AuthHandler,
	userHandler *handler.UserHandler,
	integrationHandler *handler.IntegrationHandler,
	meHandler *handler.MeHandler,
	productHandler *handler.ProductHandler,
) {
	// Health check — unauthenticated
	e.GET("/health", healthHandler)

	// Auth routes (no JWT middleware — these create/verify identity)
	auth := e.Group("/api/v1/auth")
	auth.POST("/register", tenantHandler.Register)
	auth.POST("/login", authHandler.Login)
	auth.POST("/logout", authHandler.Logout)

	// Protected routes — require valid JWT + tenant context
	// TenantMiddleware is optional here when pool is nil (dev/test without DB).
	var apiMiddlewares []echo.MiddlewareFunc
	apiMiddlewares = append(apiMiddlewares, authMW.Handle())
	if pool != nil {
		tenantMW := mw.NewTenantMiddleware(pool)
		apiMiddlewares = append(apiMiddlewares, tenantMW.Handle())
	}

	api := e.Group("/api/v1", apiMiddlewares...)

	// Tenant
	api.GET("/tenants/me", meHandler.GetMe)
	api.GET("/users", userHandler.List)
	api.POST("/users/invite", userHandler.Invite, mw.OwnerGuard())
	api.PATCH("/users/:id", userHandler.UpdateActive, mw.OwnerGuard())
	api.PUT("/tenant/integrations/fiscal", integrationHandler.UpsertFiscal, mw.OwnerGuard())
	api.GET("/tenant/integrations/fiscal", integrationHandler.GetFiscal)

	// Products — NOTE: /products/categories must be registered BEFORE /products/:id
	// so Echo's router doesn't match "categories" as a UUID parameter.
	api.GET("/products/categories", productHandler.ListCategories)
	api.POST("/products/categories", productHandler.CreateCategory)
	api.GET("/products", productHandler.ListProducts)
	api.POST("/products", productHandler.CreateProduct)
	api.GET("/products/:id", productHandler.GetProduct)
	api.PUT("/products/:id", productHandler.UpdateProduct)
	api.DELETE("/products/:id", productHandler.DeleteProduct)
}

// healthHandler responds with a simple status OK payload.
// Used by load balancers and uptime monitors.
func healthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status": "ok",
	})
}
