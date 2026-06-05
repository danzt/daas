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
	"github.com/danzt/daas/api/internal/adapter/storage"
	"github.com/danzt/daas/api/internal/app"
	"github.com/danzt/daas/api/internal/domain/notification"
)

// RouterConfig holds dependencies needed to build the Echo router.
type RouterConfig struct {
	SupabaseURL    string
	AnonKey        string
	ServiceRoleKey string
	Pool           *pgxpool.Pool
	// NotificationSvc + StorefrontURL — used for transactional emails
	// sent to customers when their order changes state. Optional; when nil,
	// state transitions still succeed but no email is dispatched.
	NotificationSvc notification.NotificationService
	StorefrontURL   string
	// Storage is used by the payment-proof upload endpoint.
	// When nil, the endpoint is still registered but upload calls will fail.
	Storage storage.Storage
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
func NewRouterWithConfig(cfg RouterConfig) *echo.Echo { //nolint:funlen,cyclop
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

	authHandler := handler.NewAuthHandler(cfg.SupabaseURL, cfg.AnonKey)
	userHandler := handler.NewUserHandler(cfg.Pool, cfg.SupabaseURL, cfg.ServiceRoleKey)
	integrationHandler := handler.NewIntegrationHandler(cfg.Pool)
	meHandler := handler.NewMeHandler(cfg.Pool)
	productHandler := handler.NewProductHandler(cfg.Pool)
	if cfg.Storage != nil {
		productHandler.SetStorage(cfg.Storage)
	}
	inventoryHandler := handler.NewInventoryHandler(cfg.Pool)
	// Create InternalInvoiceService once — shared between InvoiceHandler and auto-invoice.
	internalInvoiceSvc := app.NewInternalInvoiceService(cfg.Pool)
	invoiceHandler := handler.NewInvoiceHandlerWithService(internalInvoiceSvc)
	fiscalInvoiceHandler := handler.NewFiscalInvoiceHandler(cfg.Pool)
	supplierHandler := handler.NewSupplierHandler(cfg.Pool)
	reportHandler := handler.NewReportHandler(cfg.Pool)
	saleHandler := handler.NewSaleHandler(cfg.Pool)

	// Build lifecycle email notifier from injected NotificationSvc.
	// nil-safe: if the notification service is not configured, the notifier
	// stays nil and ShopOrderService.SetNotifier(nil) skips all sends.
	var shopNotifier *app.ShopOrderNotifier
	if cfg.NotificationSvc != nil && cfg.Pool != nil {
		shopNotifier = app.NewShopOrderNotifier(cfg.NotificationSvc, cfg.Pool, cfg.StorefrontURL)
	}
	shopOrderHandler := handler.NewShopOrderHandler(cfg.Pool, shopNotifier, cfg.Storage)
	if cfg.Pool != nil {
		shopOrderHandler.SetInvoiceService(internalInvoiceSvc)
	}
	paymentMethodHandler := handler.NewPaymentMethodHandler(cfg.Pool)

	// Routes
	registerRoutes(e, authMW, cfg.Pool, tenantHandler, authHandler, userHandler, integrationHandler, meHandler, productHandler, inventoryHandler, invoiceHandler, fiscalInvoiceHandler, supplierHandler, reportHandler, saleHandler, shopOrderHandler, paymentMethodHandler)

	// Storefront public route group — no auth required.
	// Rate limiter runs first (fast reject), then tenant resolver activates RLS.
	// Handlers for catalog endpoints are wired in S6-T8.
	if cfg.Pool != nil {
		// Serve uploaded payment proofs (and any other stored files) as static assets.
		// Files are written by LocalFSAdapter under ./storage, so the URL
		// http://localhost:8080/files/payment-proofs/... maps to ./storage/payment-proofs/...
		e.Static("/files", "./storage")

		rateLimiter := mw.NewRateLimiter()
		pathResolver := mw.NewPathResolver(cfg.Pool)
		publicTenantMW := mw.NewPublicTenantMiddleware(pathResolver, cfg.Pool)

		storefront := e.Group("/t/:tenantSlug/shop/v1")
		storefront.Use(rateLimiter.Handle())
		storefront.Use(publicTenantMW.Handle())

		// Smoke-check healthz endpoint — verifies the middleware chain is wired
		// correctly without requiring the full catalog handler (S6-T8).
		storefront.GET("/healthz", func(c echo.Context) error {
			return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
		})

		// Public catalog endpoints (S6-T8).
		shopSvc := app.NewShopService(cfg.Pool)
		shopHandler := handler.NewShopHandler(shopSvc)
		storefront.GET("/products", shopHandler.ListProducts)
		storefront.GET("/products/:id", shopHandler.GetProduct)

		// Public checkout + order tracking endpoints (S7-PR2).
		storefront.POST("/checkout", shopOrderHandler.Checkout)
		storefront.GET("/orders/:id", shopOrderHandler.GetOrder)
		storefront.POST("/orders/:id/cancel", shopOrderHandler.CancelOrder)

		// Public payment methods (S8-PR1).
		storefront.GET("/payment-methods", paymentMethodHandler.PublicList)

		// Payment proof upload (S8-PR2). Body limit slightly above 5MB to allow
		// multipart framing overhead; the service enforces exact 5MB on the file itself.
		storefront.POST("/orders/:id/payment-proof",
			shopOrderHandler.UploadPaymentProof,
			middleware.BodyLimit("6M"),
		)
	}

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
	inventoryHandler *handler.InventoryHandler,
	invoiceHandler *handler.InvoiceHandler,
	fiscalInvoiceHandler *handler.FiscalInvoiceHandler,
	supplierHandler *handler.SupplierHandler,
	reportHandler *handler.ReportHandler,
	saleHandler *handler.SaleHandler,
	shopOrderHandler *handler.ShopOrderHandler,
	paymentMethodHandler *handler.PaymentMethodHandler,
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
	api.DELETE("/products/categories/:id", productHandler.DeleteCategory, mw.OwnerGuard())
	api.GET("/products", productHandler.ListProducts)
	api.POST("/products", productHandler.CreateProduct)
	api.GET("/products/:id", productHandler.GetProduct)
	api.PUT("/products/:id", productHandler.UpdateProduct)
	api.DELETE("/products/:id", productHandler.DeleteProduct)
	api.POST("/products/:id/image", productHandler.UploadProductImage, middleware.BodyLimit("6M"))
	api.DELETE("/products/:id/image", productHandler.DeleteProductImage)

	// Inventory
	api.GET("/inventory/stock", inventoryHandler.ListStock)
	api.GET("/inventory/stock/:product_id", inventoryHandler.GetProductStock)
	api.GET("/inventory/movements", inventoryHandler.ListMovements)
	api.GET("/inventory/movements/:id", inventoryHandler.GetMovement)
	api.POST("/inventory/adjustments", inventoryHandler.CreateAdjustment)

	// Internal invoices — action routes registered before /:id to avoid conflicts
	api.POST("/invoices/internal", invoiceHandler.Create)
	api.GET("/invoices/internal", invoiceHandler.List)
	api.GET("/invoices/internal/:id", invoiceHandler.GetByID)
	api.POST("/invoices/internal/:id/issue", invoiceHandler.Issue)
	api.POST("/invoices/internal/:id/cancel", invoiceHandler.Cancel)

	// Fiscal invoices (SENIAT) — same ordering convention
	api.POST("/invoices/fiscal", fiscalInvoiceHandler.Create)
	api.GET("/invoices/fiscal", fiscalInvoiceHandler.List)
	api.GET("/invoices/fiscal/:id", fiscalInvoiceHandler.GetByID)
	api.POST("/invoices/fiscal/:id/issue", fiscalInvoiceHandler.Issue)
	api.POST("/invoices/fiscal/:id/cancel", fiscalInvoiceHandler.Cancel)
	api.POST("/invoices/fiscal/:id/retry", fiscalInvoiceHandler.Retry)

	// Sales Orders — action routes before /:id
	api.POST("/sales-orders", saleHandler.CreateOrder)
	api.GET("/sales-orders", saleHandler.ListOrders)
	api.GET("/sales-orders/:id", saleHandler.GetOrder)
	api.POST("/sales-orders/:id/confirm", saleHandler.ConfirmOrder)
	api.POST("/sales-orders/:id/invoice", saleHandler.InvoiceOrder)
	api.POST("/sales-orders/:id/cancel", saleHandler.CancelOrder)

	// Reports — JSON
	api.GET("/reports/sales", reportHandler.SalesReport)
	api.GET("/reports/inventory", reportHandler.InventoryReport)
	api.GET("/reports/purchases", reportHandler.PurchaseReport)

	// Reports — CSV export
	api.GET("/reports/sales/export", reportHandler.ExportSalesCSV)
	api.GET("/reports/inventory/export", reportHandler.ExportInventoryCSV)
	api.GET("/reports/purchases/export", reportHandler.ExportPurchasesCSV)
	// Suppliers
	api.GET("/suppliers", supplierHandler.ListSuppliers)
	api.POST("/suppliers", supplierHandler.CreateSupplier)
	api.GET("/suppliers/:id", supplierHandler.GetSupplier)
	api.PUT("/suppliers/:id", supplierHandler.UpdateSupplier)

	// Purchase Orders
	api.GET("/purchase-orders", supplierHandler.ListPOs)
	api.POST("/purchase-orders", supplierHandler.CreatePO)
	api.GET("/purchase-orders/:id", supplierHandler.GetPO)
	api.POST("/purchase-orders/:id/order", supplierHandler.OrderPO)
	api.POST("/purchase-orders/:id/receive", supplierHandler.ReceivePO)
	api.POST("/purchase-orders/:id/cancel", supplierHandler.CancelPO)

	// Shop Orders — tenant admin lifecycle (S7-PR2)
	api.GET("/shop-orders", shopOrderHandler.AdminList)
	api.GET("/shop-orders/:id", shopOrderHandler.AdminGet)
	api.POST("/shop-orders/:id/mark-paid", shopOrderHandler.MarkPaid)
	api.POST("/shop-orders/:id/mark-fulfilled", shopOrderHandler.MarkFulfilled)
	api.POST("/shop-orders/:id/mark-delivered", shopOrderHandler.MarkDelivered)
	api.POST("/shop-orders/:id/cancel", shopOrderHandler.AdminCancel)

	// Payment Methods — tenant settings (S8-PR1)
	api.GET("/payment-methods", paymentMethodHandler.AdminList)
	api.POST("/payment-methods", paymentMethodHandler.AdminCreate)
	api.PUT("/payment-methods/:id", paymentMethodHandler.AdminUpdate)
	api.DELETE("/payment-methods/:id", paymentMethodHandler.AdminDelete)
}

// healthHandler responds with a simple status OK payload.
// Used by load balancers and uptime monitors.
func healthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status": "ok",
	})
}
