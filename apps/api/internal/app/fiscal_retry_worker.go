package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

const (
	retryWorkerInterval  = 2 * time.Minute
	retryWorkerBatchSize = 20
)

// FiscalRetryWorker polls fiscal_invoice_queue and retries failed fiscal invoices
// that are due for retry. Exponential backoff is managed by the queue table:
// each failed attempt pushes next_retry_at further out
// (5 min * 2^attempts). The worker simply processes whatever is due.
type FiscalRetryWorker struct {
	svc      *FiscalInvoiceService
	interval time.Duration
}

// NewFiscalRetryWorker creates a worker backed by the given FiscalInvoiceService.
func NewFiscalRetryWorker(svc *FiscalInvoiceService) *FiscalRetryWorker {
	return &FiscalRetryWorker{
		svc:      svc,
		interval: retryWorkerInterval,
	}
}

// Start polls the fiscal queue on every tick until ctx is cancelled.
// Invoke as a goroutine: go worker.Start(ctx).
//
// It also fires one immediate tick on startup so invoices that failed
// before the last server restart are not delayed a full interval.
func (w *FiscalRetryWorker) Start(ctx context.Context) {
	log.Info().Dur("interval", w.interval).Msg("fiscal-retry-worker: started")
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	// Immediate tick — catch any backlog before the first scheduled tick.
	w.tick(ctx)

	for {
		select {
		case <-ticker.C:
			w.tick(ctx)
		case <-ctx.Done():
			log.Info().Msg("fiscal-retry-worker: stopped")
			return
		}
	}
}

// tick fetches all due queue entries and retries each one.
// Successes are dequeued by issueInternal. Failures increment attempts
// and push next_retry_at back (exponential backoff via the queue upsert).
// Entries that have exhausted max_attempts are not picked up and remain
// in failed status for manual retry via POST /invoices/fiscal/:id/retry.
func (w *FiscalRetryWorker) tick(ctx context.Context) {
	type entry struct {
		invoiceID uuid.UUID
		tenantID  uuid.UUID
	}

	rows, err := w.svc.pool.Query(ctx,
		`SELECT invoice_id, tenant_id
		 FROM fiscal_invoice_queue
		 WHERE next_retry_at <= NOW() AND attempts < max_attempts
		 ORDER BY next_retry_at
		 LIMIT $1`,
		retryWorkerBatchSize,
	)
	if err != nil {
		log.Error().Err(err).Msg("fiscal-retry-worker: failed to poll queue")
		return
	}

	var entries []entry
	for rows.Next() {
		var e entry
		if err := rows.Scan(&e.invoiceID, &e.tenantID); err != nil {
			log.Error().Err(err).Msg("fiscal-retry-worker: scan queue entry")
			continue
		}
		entries = append(entries, e)
	}
	rows.Close()

	if len(entries) == 0 {
		return
	}

	log.Info().Int("count", len(entries)).Msg("fiscal-retry-worker: processing due entries")

	for _, e := range entries {
		_, err := w.svc.issueInternal(ctx, e.tenantID, e.invoiceID)
		if err != nil {
			log.Warn().
				Err(err).
				Str("invoice_id", e.invoiceID.String()).
				Str("tenant_id", e.tenantID.String()).
				Msg("fiscal-retry-worker: retry failed — rescheduled")
		} else {
			log.Info().
				Str("invoice_id", e.invoiceID.String()).
				Str("tenant_id", e.tenantID.String()).
				Msg("fiscal-retry-worker: invoice issued successfully")
		}
	}
}
