package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	fiscaladapter "github.com/danzt/daas/api/internal/adapter/fiscal"
	"github.com/danzt/daas/api/internal/domain/fiscal"
)

// FiscalInvoiceListFilters holds optional query parameters for listing fiscal invoices.
type FiscalInvoiceListFilters struct {
	Status *fiscal.Status
	From   *time.Time
	To     *time.Time
	Limit  int
	Offset int
}

// FiscalInvoiceService handles all fiscal invoicing use cases.
type FiscalInvoiceService struct {
	pool    *pgxpool.Pool
	adapter fiscaladapter.Adapter
}

// NewFiscalInvoiceService creates a service backed by the given pool and SENIAT adapter.
func NewFiscalInvoiceService(pool *pgxpool.Pool, adapter fiscaladapter.Adapter) *FiscalInvoiceService {
	return &FiscalInvoiceService{pool: pool, adapter: adapter}
}

// ─── Create ───────────────────────────────────────────────────────────────────

// Create validates and persists a new draft fiscal invoice with its lines.
// All products MUST be fiscal; non-fiscal products are rejected.
func (s *FiscalInvoiceService) Create(ctx context.Context, tenantID uuid.UUID, createdBySupabaseUID string, req fiscal.CreateRequest) (*fiscal.FiscalInvoice, error) {
	createdBy, err := s.resolveFiscalUserID(ctx, tenantID, createdBySupabaseUID)
	if err != nil {
		return nil, err
	}

	// Resolve product details — validate is_fiscal flag and set UnitPrice/TaxRate.
	for i, line := range req.Lines {
		var isFiscal bool
		var fiscalPrice, taxRate float64
		err := s.pool.QueryRow(ctx,
			`SELECT is_fiscal,
			        COALESCE(fiscal_price, 0),
			        COALESCE(tax_rate, 0) / 100.0
			 FROM products
			 WHERE id = $1 AND tenant_id = $2 AND active = true`,
			line.ProductID, tenantID,
		).Scan(&isFiscal, &fiscalPrice, &taxRate)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, fmt.Errorf("product %s not found or inactive", line.ProductID)
			}
			return nil, fmt.Errorf("lookup product: %w", err)
		}
		if !isFiscal {
			return nil, fiscal.ErrNonFiscalProductForbidden
		}
		req.Lines[i].IsFiscal = true
		if req.Lines[i].UnitPrice == 0 {
			req.Lines[i].UnitPrice = fiscalPrice
		}
		if req.Lines[i].TaxRate == 0 {
			req.Lines[i].TaxRate = taxRate
		}
	}

	if err := req.Validate(); err != nil {
		return nil, err
	}

	subtotalBase, taxAmount, total := fiscal.ComputeTotals(req.Lines)

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	inv := &fiscal.FiscalInvoice{
		ID:               uuid.New(),
		TenantID:         tenantID,
		CustomerName:     req.CustomerName,
		CustomerIDType:   req.CustomerIDType,
		CustomerIDNumber: req.CustomerIDNumber,
		Notes:            req.Notes,
		SubtotalBase:     subtotalBase,
		TaxAmount:        taxAmount,
		Total:            total,
		Status:           fiscal.StatusDraft,
		CreatedBy:        createdBy,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO fiscal_invoices
		    (id, tenant_id, customer_name, customer_id_type, customer_id_number,
		     notes, subtotal_base, tax_amount, total, status, created_by)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		inv.ID, inv.TenantID, inv.CustomerName, inv.CustomerIDType, inv.CustomerIDNumber,
		inv.Notes, inv.SubtotalBase, inv.TaxAmount, inv.Total, inv.Status, inv.CreatedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("insert fiscal invoice: %w", err)
	}

	for i, line := range req.Lines {
		base := line.Quantity * line.UnitPrice
		lineTax := base * line.TaxRate
		lineSubtotal := base + lineTax
		lineID := uuid.New()
		_, err = tx.Exec(ctx,
			`INSERT INTO fiscal_invoice_lines
			    (id, invoice_id, product_id, description, quantity,
			     unit_price, tax_rate, tax_amount, subtotal, sort_order)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
			lineID, inv.ID, line.ProductID, line.Description, line.Quantity,
			line.UnitPrice, line.TaxRate, lineTax, lineSubtotal, i,
		)
		if err != nil {
			return nil, fmt.Errorf("insert fiscal invoice line: %w", err)
		}
		inv.Lines = append(inv.Lines, fiscal.FiscalInvoiceLine{
			ID:          lineID,
			InvoiceID:   inv.ID,
			ProductID:   line.ProductID,
			Description: line.Description,
			Quantity:    line.Quantity,
			UnitPrice:   line.UnitPrice,
			TaxRate:     line.TaxRate,
			TaxAmount:   lineTax,
			Subtotal:    lineSubtotal,
			SortOrder:   i,
		})
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}
	return inv, nil
}

// ─── Issue ────────────────────────────────────────────────────────────────────

// Issue sends the fiscal invoice to the SENIAT machine.
// On success, the invoice transitions to StatusIssued and receives its fiscal number.
// On failure, the invoice transitions to StatusFailed and is queued for retry.
func (s *FiscalInvoiceService) Issue(ctx context.Context, tenantID, invoiceID uuid.UUID, issuedBySupabaseUID string) (*fiscal.FiscalInvoice, error) {
	_, err := s.resolveFiscalUserID(ctx, tenantID, issuedBySupabaseUID)
	if err != nil {
		return nil, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	inv, err := s.loadFiscalForUpdate(ctx, tx, tenantID, invoiceID)
	if err != nil {
		return nil, err
	}
	if inv.Status != fiscal.StatusDraft && inv.Status != fiscal.StatusFailed {
		if inv.Status == fiscal.StatusIssued {
			return nil, fiscal.ErrInvoiceAlreadyIssued
		}
		if inv.Status == fiscal.StatusCancelled {
			return nil, fiscal.ErrInvoiceAlreadyCancelled
		}
		return nil, fiscal.ErrInvoiceNotDraft
	}
	if inv.RetryCount >= 3 {
		return nil, fiscal.ErrInvoiceCannotRetry
	}

	// Mark as pending before calling the adapter
	_, err = tx.Exec(ctx,
		`UPDATE fiscal_invoices SET status='pending_fiscal', updated_at=NOW() WHERE id=$1`,
		invoiceID,
	)
	if err != nil {
		return nil, fmt.Errorf("set pending_fiscal: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit pending tx: %w", err)
	}

	// Build SENIAT send request
	lines, err := s.loadFiscalLines(ctx, invoiceID)
	if err != nil {
		return nil, err
	}
	sendLines := make([]fiscaladapter.SendLine, len(lines))
	for i, l := range lines {
		sendLines[i] = fiscaladapter.SendLine{
			Description: l.Description,
			Quantity:    l.Quantity,
			UnitPrice:   l.UnitPrice,
			TaxRate:     l.TaxRate,
		}
	}
	resp, sendErr := s.adapter.Send(ctx, fiscaladapter.SendRequest{
		InvoiceID:        invoiceID,
		CustomerIDType:   string(inv.CustomerIDType),
		CustomerIDNumber: inv.CustomerIDNumber,
		CustomerName:     inv.CustomerName,
		Lines:            sendLines,
	})

	// Update outcome in a new transaction
	tx2, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin outcome tx: %w", err)
	}
	defer tx2.Rollback(ctx) //nolint:errcheck

	if sendErr != nil {
		// Transition to failed + upsert queue entry
		failReason := sendErr.Error()
		_, err = tx2.Exec(ctx,
			`UPDATE fiscal_invoices
			 SET status='failed', fail_reason=$1, retry_count=retry_count+1, updated_at=NOW()
			 WHERE id=$2`,
			failReason, invoiceID,
		)
		if err != nil {
			return nil, fmt.Errorf("mark failed: %w", err)
		}
		_, err = tx2.Exec(ctx,
			`INSERT INTO fiscal_invoice_queue (invoice_id, tenant_id, attempts, last_error, next_retry_at)
			 VALUES ($1, $2, 1, $3, NOW() + INTERVAL '5 minutes')
			 ON CONFLICT (invoice_id) DO UPDATE
			   SET attempts=fiscal_invoice_queue.attempts+1,
			       last_error=$3,
			       next_retry_at=NOW() + (INTERVAL '5 minutes' * POWER(2, fiscal_invoice_queue.attempts))`,
			invoiceID, tenantID, failReason,
		)
		if err != nil {
			return nil, fmt.Errorf("upsert queue: %w", err)
		}
		if err := tx2.Commit(ctx); err != nil {
			return nil, fmt.Errorf("commit failed tx: %w", err)
		}
		return nil, fmt.Errorf("SENIAT send failed: %w", sendErr)
	}

	// Success — update with fiscal fields
	now := resp.IssuedAt
	_, err = tx2.Exec(ctx,
		`UPDATE fiscal_invoices
		 SET status='issued',
		     fiscal_number=$1,
		     machine_serial=$2,
		     report_z_number=$3,
		     issued_at=$4,
		     fail_reason=NULL,
		     updated_at=NOW()
		 WHERE id=$5`,
		resp.FiscalNumber, resp.MachineSerial, resp.ReportZNumber, now, invoiceID,
	)
	if err != nil {
		return nil, fmt.Errorf("mark issued: %w", err)
	}
	// Remove from queue if present
	_, err = tx2.Exec(ctx, `DELETE FROM fiscal_invoice_queue WHERE invoice_id=$1`, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("dequeue: %w", err)
	}
	if err := tx2.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit issued tx: %w", err)
	}

	return s.GetByID(ctx, tenantID, invoiceID)
}

// ─── Cancel ───────────────────────────────────────────────────────────────────

// Cancel transitions a draft or failed invoice to cancelled status.
// Issued invoices cannot be cancelled via this endpoint (fiscal law constraint).
func (s *FiscalInvoiceService) Cancel(ctx context.Context, tenantID, invoiceID uuid.UUID, cancelledBySupabaseUID, notes string) (*fiscal.FiscalInvoice, error) {
	_, err := s.resolveFiscalUserID(ctx, tenantID, cancelledBySupabaseUID)
	if err != nil {
		return nil, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	inv, err := s.loadFiscalForUpdate(ctx, tx, tenantID, invoiceID)
	if err != nil {
		return nil, err
	}
	if inv.Status == fiscal.StatusCancelled {
		return nil, fiscal.ErrInvoiceAlreadyCancelled
	}
	if inv.Status == fiscal.StatusIssued {
		return nil, fiscal.ErrInvoiceAlreadyIssued
	}

	cancelNote := inv.Notes
	if notes != "" {
		cancelNote = notes
	}
	_, err = tx.Exec(ctx,
		`UPDATE fiscal_invoices
		 SET status='cancelled', notes=$1, updated_at=NOW()
		 WHERE id=$2`,
		cancelNote, invoiceID,
	)
	if err != nil {
		return nil, fmt.Errorf("cancel invoice: %w", err)
	}
	_, err = tx.Exec(ctx, `DELETE FROM fiscal_invoice_queue WHERE invoice_id=$1`, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("dequeue on cancel: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit cancel: %w", err)
	}

	return s.GetByID(ctx, tenantID, invoiceID)
}

// ─── Retry ────────────────────────────────────────────────────────────────────

// Retry re-sends a failed fiscal invoice to the SENIAT machine.
func (s *FiscalInvoiceService) Retry(ctx context.Context, tenantID, invoiceID uuid.UUID, retriedBySupabaseUID string) (*fiscal.FiscalInvoice, error) {
	return s.Issue(ctx, tenantID, invoiceID, retriedBySupabaseUID)
}

// ─── Read ─────────────────────────────────────────────────────────────────────

// GetByID loads a fiscal invoice with its lines.
func (s *FiscalInvoiceService) GetByID(ctx context.Context, tenantID, invoiceID uuid.UUID) (*fiscal.FiscalInvoice, error) {
	inv, err := s.scanFiscalInvoice(s.pool.QueryRow(ctx,
		`SELECT id, tenant_id, fiscal_number, machine_serial, report_z_number,
		        customer_name, customer_id_type, customer_id_number,
		        subtotal_base, tax_amount, total,
		        status, fail_reason, retry_count, notes,
		        issued_at, created_by, created_at, updated_at
		 FROM fiscal_invoices
		 WHERE id=$1 AND tenant_id=$2`,
		invoiceID, tenantID,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fiscal.ErrInvoiceNotFound
		}
		return nil, fmt.Errorf("load fiscal invoice: %w", err)
	}
	lines, err := s.loadFiscalLines(ctx, invoiceID)
	if err != nil {
		return nil, err
	}
	inv.Lines = lines
	return inv, nil
}

// List returns fiscal invoices for a tenant with optional filters.
func (s *FiscalInvoiceService) List(ctx context.Context, tenantID uuid.UUID, f FiscalInvoiceListFilters) ([]*fiscal.FiscalInvoice, error) {
	query := `SELECT id, tenant_id, fiscal_number, machine_serial, report_z_number,
	                 customer_name, customer_id_type, customer_id_number,
	                 subtotal_base, tax_amount, total,
	                 status, fail_reason, retry_count, notes,
	                 issued_at, created_by, created_at, updated_at
	          FROM fiscal_invoices
	          WHERE tenant_id=$1`
	args := []any{tenantID}
	idx := 2

	if f.Status != nil {
		query += fmt.Sprintf(" AND status=$%d", idx)
		args = append(args, *f.Status)
		idx++
	}
	if f.From != nil {
		query += fmt.Sprintf(" AND created_at >= $%d", idx)
		args = append(args, *f.From)
		idx++
	}
	if f.To != nil {
		query += fmt.Sprintf(" AND created_at <= $%d", idx)
		args = append(args, *f.To)
		idx++
	}
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", idx, idx+1)
	args = append(args, f.Limit, f.Offset)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list fiscal invoices: %w", err)
	}
	defer rows.Close()

	var result []*fiscal.FiscalInvoice
	for rows.Next() {
		inv, err := s.scanFiscalInvoice(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, inv)
	}
	return result, nil
}

// ─── Private helpers ──────────────────────────────────────────────────────────

func (s *FiscalInvoiceService) resolveFiscalUserID(ctx context.Context, tenantID uuid.UUID, supabaseUID string) (uuid.UUID, error) {
	var id uuid.UUID
	err := s.pool.QueryRow(ctx,
		`SELECT id FROM tenant_users WHERE tenant_id=$1 AND supabase_uid=$2`,
		tenantID, supabaseUID,
	).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("resolve user: %w", err)
	}
	return id, nil
}

func (s *FiscalInvoiceService) loadFiscalForUpdate(ctx context.Context, tx pgx.Tx, tenantID, invoiceID uuid.UUID) (*fiscal.FiscalInvoice, error) {
	inv, err := s.scanFiscalInvoice(tx.QueryRow(ctx,
		`SELECT id, tenant_id, fiscal_number, machine_serial, report_z_number,
		        customer_name, customer_id_type, customer_id_number,
		        subtotal_base, tax_amount, total,
		        status, fail_reason, retry_count, notes,
		        issued_at, created_by, created_at, updated_at
		 FROM fiscal_invoices
		 WHERE id=$1 AND tenant_id=$2
		 FOR UPDATE`,
		invoiceID, tenantID,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fiscal.ErrInvoiceNotFound
		}
		return nil, fmt.Errorf("load fiscal for update: %w", err)
	}
	return inv, nil
}

func (s *FiscalInvoiceService) loadFiscalLines(ctx context.Context, invoiceID uuid.UUID) ([]fiscal.FiscalInvoiceLine, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, invoice_id, product_id, description, quantity,
		        unit_price, tax_rate, tax_amount, subtotal, sort_order
		 FROM fiscal_invoice_lines
		 WHERE invoice_id=$1
		 ORDER BY sort_order`,
		invoiceID,
	)
	if err != nil {
		return nil, fmt.Errorf("load fiscal lines: %w", err)
	}
	defer rows.Close()

	var lines []fiscal.FiscalInvoiceLine
	for rows.Next() {
		var l fiscal.FiscalInvoiceLine
		if err := rows.Scan(
			&l.ID, &l.InvoiceID, &l.ProductID, &l.Description, &l.Quantity,
			&l.UnitPrice, &l.TaxRate, &l.TaxAmount, &l.Subtotal, &l.SortOrder,
		); err != nil {
			return nil, fmt.Errorf("scan fiscal line: %w", err)
		}
		lines = append(lines, l)
	}
	return lines, nil
}

type fiscalScanner interface {
	Scan(dest ...any) error
}

func (s *FiscalInvoiceService) scanFiscalInvoice(row fiscalScanner) (*fiscal.FiscalInvoice, error) {
	var inv fiscal.FiscalInvoice
	if err := row.Scan(
		&inv.ID, &inv.TenantID,
		&inv.FiscalNumber, &inv.MachineSerial, &inv.ReportZNumber,
		&inv.CustomerName, &inv.CustomerIDType, &inv.CustomerIDNumber,
		&inv.SubtotalBase, &inv.TaxAmount, &inv.Total,
		&inv.Status, &inv.FailReason, &inv.RetryCount, &inv.Notes,
		&inv.IssuedAt, &inv.CreatedBy, &inv.CreatedAt, &inv.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &inv, nil
}
