package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

// DashboardHandler exposes the combined dashboard KPI endpoint.
type DashboardHandler struct {
	pool *pgxpool.Pool
}

func NewDashboardHandler(pool *pgxpool.Pool) *DashboardHandler {
	return &DashboardHandler{pool: pool}
}

// ─── Response types ───────────────────────────────────────────────────────────

type DashboardResponse struct {
	Today            PeriodStats        `json:"today"`
	Month            PeriodStats        `json:"month"`
	Pending          PendingStats       `json:"pending"`
	Inventory        InventoryStats     `json:"inventory"`
	RecentShopOrders []ShopOrderSummary `json:"recent_shop_orders"`
}

type PeriodStats struct {
	B2BRevenue   float64 `json:"b2b_revenue"`
	B2CRevenue   float64 `json:"b2c_revenue"`
	InvoiceCount int     `json:"invoice_count"`
}

type PendingStats struct {
	ShopOrders  int `json:"shop_orders"`
	SalesOrders int `json:"sales_orders"`
}

type InventoryStats struct {
	LowStock   int `json:"low_stock"`
	OutOfStock int `json:"out_of_stock"`
}

type ShopOrderSummary struct {
	ID            string  `json:"id"`
	CustomerName  string  `json:"customer_name"`
	CustomerEmail string  `json:"customer_email"`
	Total         float64 `json:"total"`
	Status        string  `json:"status"`
	CreatedAt     string  `json:"created_at"`
}

// ─── Handler ──────────────────────────────────────────────────────────────────

// Get handles GET /api/v1/dashboard
//
// Returns a single combined JSON payload with all dashboard KPIs for the
// authenticated tenant. Uses a single query with multiple CTEs.
func (h *DashboardHandler) Get(c echo.Context) error {
	tenantID, err := getTenantID(c)
	if err != nil {
		return WriteProblem(c, http.StatusForbidden, "forbidden", "tenant context missing")
	}

	resp, err := h.query(c.Request().Context(), tenantID.String())
	if err != nil {
		return WriteProblem(c, http.StatusInternalServerError, "internal-error", err.Error())
	}

	return c.JSON(http.StatusOK, resp)
}

// query executes the single CTE-based dashboard query and maps rows.
func (h *DashboardHandler) query(ctx context.Context, tenantID string) (*DashboardResponse, error) {
	const sql = `
WITH
-- ── Today bounds ────────────────────────────────────────────────────────────
today_start AS (
  SELECT CURRENT_DATE::TIMESTAMPTZ AS ts
),

-- ── Month start ─────────────────────────────────────────────────────────────
month_start AS (
  SELECT DATE_TRUNC('month', CURRENT_DATE)::TIMESTAMPTZ AS ts
),

-- ── B2B revenue today ────────────────────────────────────────────────────────
b2b_today AS (
  SELECT COALESCE(SUM(total), 0)::float8 AS revenue,
         COUNT(*)::int                   AS cnt
  FROM internal_invoices
  WHERE tenant_id = $1::uuid
    AND status    = 'issued'
    AND issued_at >= (SELECT ts FROM today_start)
),

-- ── B2B revenue this month ───────────────────────────────────────────────────
b2b_month AS (
  SELECT COALESCE(SUM(total), 0)::float8 AS revenue,
         COUNT(*)::int                   AS cnt
  FROM internal_invoices
  WHERE tenant_id = $1::uuid
    AND status    = 'issued'
    AND issued_at >= (SELECT ts FROM month_start)
),

-- ── B2C revenue today ────────────────────────────────────────────────────────
b2c_today AS (
  SELECT COALESCE(SUM(total), 0)::float8 AS revenue
  FROM shop_orders
  WHERE tenant_id = $1::uuid
    AND status    IN ('paid','fulfilled','delivered')
    AND paid_at  >= (SELECT ts FROM today_start)
),

-- ── B2C revenue this month ───────────────────────────────────────────────────
b2c_month AS (
  SELECT COALESCE(SUM(total), 0)::float8 AS revenue
  FROM shop_orders
  WHERE tenant_id = $1::uuid
    AND status    IN ('paid','fulfilled','delivered')
    AND paid_at  >= (SELECT ts FROM month_start)
),

-- ── Pending shop orders ──────────────────────────────────────────────────────
pending_shop AS (
  SELECT COUNT(*)::int AS cnt
  FROM shop_orders
  WHERE tenant_id = $1::uuid
    AND status    = 'pending'
),

-- ── Pending sales orders (confirmed, not yet invoiced) ───────────────────────
pending_sales AS (
  SELECT COUNT(*)::int AS cnt
  FROM sales_orders
  WHERE tenant_id = $1::uuid
    AND status    = 'confirmed'
),

-- ── Inventory health ─────────────────────────────────────────────────────────
inv_stats AS (
  SELECT
    COUNT(*) FILTER (WHERE quantity_on_hand > 0 AND quantity_on_hand < 5)::int AS low_stock,
    COUNT(*) FILTER (WHERE quantity_on_hand = 0)::int                           AS out_of_stock
  FROM product_stock
  WHERE tenant_id = $1::uuid
)

SELECT
  (SELECT revenue FROM b2b_today)  AS b2b_today_rev,
  (SELECT cnt     FROM b2b_today)  AS b2b_today_cnt,
  (SELECT revenue FROM b2c_today)  AS b2c_today_rev,
  (SELECT revenue FROM b2b_month)  AS b2b_month_rev,
  (SELECT cnt     FROM b2b_month)  AS b2b_month_cnt,
  (SELECT revenue FROM b2c_month)  AS b2c_month_rev,
  (SELECT cnt     FROM pending_shop)  AS pending_shop_cnt,
  (SELECT cnt     FROM pending_sales) AS pending_sales_cnt,
  (SELECT low_stock   FROM inv_stats) AS low_stock,
  (SELECT out_of_stock FROM inv_stats) AS out_of_stock
`

	var (
		b2bTodayRev  float64
		b2bTodayCnt  int
		b2cTodayRev  float64
		b2bMonthRev  float64
		b2bMonthCnt  int
		b2cMonthRev  float64
		pendingShop  int
		pendingSales int
		lowStock     int
		outOfStock   int
	)

	row := h.pool.QueryRow(ctx, sql, tenantID)
	if err := row.Scan(
		&b2bTodayRev, &b2bTodayCnt, &b2cTodayRev,
		&b2bMonthRev, &b2bMonthCnt, &b2cMonthRev,
		&pendingShop, &pendingSales,
		&lowStock, &outOfStock,
	); err != nil {
		return nil, err
	}

	// Recent shop orders — separate simple query (LIMIT 5)
	recent, err := h.recentShopOrders(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	return &DashboardResponse{
		Today: PeriodStats{
			B2BRevenue:   b2bTodayRev,
			B2CRevenue:   b2cTodayRev,
			InvoiceCount: b2bTodayCnt,
		},
		Month: PeriodStats{
			B2BRevenue:   b2bMonthRev,
			B2CRevenue:   b2cMonthRev,
			InvoiceCount: b2bMonthCnt,
		},
		Pending: PendingStats{
			ShopOrders:  pendingShop,
			SalesOrders: pendingSales,
		},
		Inventory: InventoryStats{
			LowStock:   lowStock,
			OutOfStock: outOfStock,
		},
		RecentShopOrders: recent,
	}, nil
}

// recentShopOrders fetches the last 5 shop orders for the tenant.
func (h *DashboardHandler) recentShopOrders(ctx context.Context, tenantID string) ([]ShopOrderSummary, error) {
	const sql = `
SELECT id::text, customer_name, customer_email, total, status, created_at
FROM shop_orders
WHERE tenant_id = $1::uuid
ORDER BY created_at DESC
LIMIT 5
`

	rows, err := h.pool.Query(ctx, sql, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []ShopOrderSummary
	for rows.Next() {
		var s ShopOrderSummary
		var createdAt time.Time
		if err := rows.Scan(&s.ID, &s.CustomerName, &s.CustomerEmail, &s.Total, &s.Status, &createdAt); err != nil {
			return nil, err
		}
		s.CreatedAt = createdAt.UTC().Format(time.RFC3339)
		result = append(result, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if result == nil {
		result = []ShopOrderSummary{}
	}
	return result, nil
}
