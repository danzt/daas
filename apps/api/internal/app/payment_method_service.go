package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/danzt/daas/api/internal/domain/tenant"
)

// PaymentMethodService manages the payment methods a tenant accepts.
// It uses pgxpool.Pool directly, matching the project convention (no separate repo layer).
type PaymentMethodService struct {
	pool *pgxpool.Pool
}

// NewPaymentMethodService creates a PaymentMethodService backed by the given pool.
func NewPaymentMethodService(pool *pgxpool.Pool) *PaymentMethodService {
	return &PaymentMethodService{pool: pool}
}

// List returns ALL payment methods for a tenant (admin view, including inactive).
func (s *PaymentMethodService) List(ctx context.Context, tenantID uuid.UUID) ([]*tenant.PaymentMethod, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, tenant_id, type, COALESCE(label,''), details,
		       COALESCE(currency,''), active, sort_order, created_at, updated_at
		FROM tenant_payment_methods
		WHERE tenant_id = $1
		ORDER BY sort_order ASC, created_at ASC`,
		tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("list payment methods: %w", err)
	}
	defer rows.Close()

	return scanPaymentMethods(rows)
}

// ListActiveForStorefront returns only active methods, ordered by sort_order (public view).
func (s *PaymentMethodService) ListActiveForStorefront(ctx context.Context, tenantID uuid.UUID) ([]*tenant.PaymentMethod, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, tenant_id, type, COALESCE(label,''), details,
		       COALESCE(currency,''), active, sort_order, created_at, updated_at
		FROM tenant_payment_methods
		WHERE tenant_id = $1 AND active = TRUE
		ORDER BY sort_order ASC, created_at ASC`,
		tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("list active payment methods: %w", err)
	}
	defer rows.Close()

	return scanPaymentMethods(rows)
}

// Get returns a single payment method by tenant + id.
func (s *PaymentMethodService) Get(ctx context.Context, tenantID, id uuid.UUID) (*tenant.PaymentMethod, error) {
	pm, err := scanPaymentMethod(s.pool.QueryRow(ctx, `
		SELECT id, tenant_id, type, COALESCE(label,''), details,
		       COALESCE(currency,''), active, sort_order, created_at, updated_at
		FROM tenant_payment_methods
		WHERE id = $1 AND tenant_id = $2`,
		id, tenantID,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, tenant.ErrPaymentMethodNotFound
		}
		return nil, fmt.Errorf("get payment method: %w", err)
	}
	return pm, nil
}

// Create validates and inserts a new payment method.
func (s *PaymentMethodService) Create(ctx context.Context, tenantID uuid.UUID, req tenant.CreatePaymentMethodRequest) (*tenant.PaymentMethod, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	detailsJSON, err := json.Marshal(req.Details)
	if err != nil {
		return nil, fmt.Errorf("marshal details: %w", err)
	}

	id := uuid.New()
	now := time.Now().UTC()

	_, err = s.pool.Exec(ctx, `
		INSERT INTO tenant_payment_methods
			(id, tenant_id, type, label, details, currency, active, sort_order, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, TRUE, $7, $8, $8)`,
		id, tenantID, string(req.Type), req.Label, detailsJSON, req.Currency, req.SortOrder, now,
	)
	if err != nil {
		return nil, fmt.Errorf("insert payment method: %w", err)
	}

	return s.Get(ctx, tenantID, id)
}

// Update applies partial updates to an existing payment method.
func (s *PaymentMethodService) Update(ctx context.Context, tenantID, id uuid.UUID, req tenant.UpdatePaymentMethodRequest) (*tenant.PaymentMethod, error) {
	pm, err := s.Get(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}

	// Apply updates to local copy.
	if req.Label != nil {
		pm.Label = *req.Label
	}
	if req.Details != nil {
		pm.Details = *req.Details
	}
	if req.Currency != nil {
		pm.Currency = *req.Currency
	}
	if req.Active != nil {
		pm.Active = *req.Active
	}
	if req.SortOrder != nil {
		pm.SortOrder = *req.SortOrder
	}

	detailsJSON, err := json.Marshal(pm.Details)
	if err != nil {
		return nil, fmt.Errorf("marshal details: %w", err)
	}

	now := time.Now().UTC()
	_, err = s.pool.Exec(ctx, `
		UPDATE tenant_payment_methods
		SET label = $1, details = $2, currency = $3, active = $4, sort_order = $5, updated_at = $6
		WHERE id = $7 AND tenant_id = $8`,
		pm.Label, detailsJSON, pm.Currency, pm.Active, pm.SortOrder, now, id, tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("update payment method: %w", err)
	}

	return s.Get(ctx, tenantID, id)
}

// Delete removes a payment method permanently.
func (s *PaymentMethodService) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	result, err := s.pool.Exec(ctx, `
		DELETE FROM tenant_payment_methods
		WHERE id = $1 AND tenant_id = $2`,
		id, tenantID,
	)
	if err != nil {
		return fmt.Errorf("delete payment method: %w", err)
	}
	if result.RowsAffected() == 0 {
		return tenant.ErrPaymentMethodNotFound
	}
	return nil
}

// ─── Scan helpers ─────────────────────────────────────────────────────────────

type pmScanner interface {
	Scan(dest ...any) error
}

func scanPaymentMethod(row pmScanner) (*tenant.PaymentMethod, error) {
	var pm tenant.PaymentMethod
	var pmType string
	var detailsRaw []byte

	if err := row.Scan(
		&pm.ID, &pm.TenantID, &pmType, &pm.Label, &detailsRaw,
		&pm.Currency, &pm.Active, &pm.SortOrder, &pm.CreatedAt, &pm.UpdatedAt,
	); err != nil {
		return nil, err
	}

	pm.Type = tenant.PaymentMethodType(pmType)

	if err := json.Unmarshal(detailsRaw, &pm.Details); err != nil {
		return nil, fmt.Errorf("unmarshal details: %w", err)
	}
	if pm.Details == nil {
		pm.Details = map[string]any{}
	}

	return &pm, nil
}

func scanPaymentMethods(rows pgx.Rows) ([]*tenant.PaymentMethod, error) {
	var methods []*tenant.PaymentMethod
	for rows.Next() {
		pm, err := scanPaymentMethod(rows)
		if err != nil {
			return nil, fmt.Errorf("scan payment method: %w", err)
		}
		methods = append(methods, pm)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("payment methods rows: %w", err)
	}
	if methods == nil {
		methods = []*tenant.PaymentMethod{}
	}
	return methods, nil
}
