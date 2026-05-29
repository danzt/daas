package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// NewRouter creates a new Echo instance with standard middleware configured
// and all routes registered.
func NewRouter() *echo.Echo {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
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

	// Routes
	registerRoutes(e)

	return e
}

func registerRoutes(e *echo.Echo) {
	// Health check — unauthenticated
	e.GET("/health", healthHandler)

	// API v1 group — all business routes go here
	_ = e.Group("/api/v1")
}

// healthHandler responds with a simple status OK payload.
// Used by load balancers and uptime monitors.
func healthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status": "ok",
	})
}
