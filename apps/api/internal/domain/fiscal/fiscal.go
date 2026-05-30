// Package fiscal contains the domain entities and business rules for
// fiscal invoice management (SENIAT-compliant documents). Zero external imports.
package fiscal

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Status represents the full lifecycle of a fiscal invoice.
type Status string

const (
	StatusDraft         Status = "draft"
	StatusPendingFiscal Status = "pending_fiscal"
	StatusIssued        Status = "issued"
	StatusFailed        Status = "failed"
	StatusCancelled     Status = "cancelled"
)

// CustomerIDType classifies the customer identification document.
// Shared with the internal invoice domain but redeclared here to keep
// the packages decoupled.
type CustomerIDType string

const (
	CustomerIDTypeCedula    CustomerIDType = "cedula"
	CustomerIDTypeRIF       CustomerIDType = "rif"
	CustomerIDTypePassport  CustomerIDType = "passport"
	CustomerIDTypeAnonymous CustomerIDType = "anonymous"
)

// FiscalInvoice is the aggregate root for a SENIAT-compliant sales document.
type FiscalInvoice struct {
	ID       uuid.UUID
	TenantID uuid.UUID
	// Fiscal fields assigned by the SENIAT machine on successful send
	FiscalNumber  *string
	MachineSerial *string
	ReportZNumber *int
	// Customer
	CustomerName     string
	CustomerIDType   CustomerIDType
	CustomerIDNumber string
	// Financial totals
	SubtotalBase float64 // base imponible (pre-tax)
	TaxAmount    float64 // IVA total
	Total        float64
	// Lifecycle
	Status     Status
	FailReason *string
	RetryCount int
	Notes      string
	IssuedAt   *time.Time
	CreatedBy  uuid.UUID
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Lines      []FiscalInvoiceLine
}

// FiscalInvoiceLine is a single product line inside a fiscal invoice.
type FiscalInvoiceLine struct {
	ID          uuid.UUID
	InvoiceID   uuid.UUID
	ProductID   uuid.UUID
	Description string
	Quantity    float64
	UnitPrice   float64 // fiscal_price from product catalog
	TaxRate     float64 // e.g. 0.16 for 16%
	TaxAmount   float64 // Quantity * UnitPrice * TaxRate
	Subtotal    float64 // Quantity * UnitPrice + TaxAmount
	SortOrder   int
}

// CreateLineRequest is the input for one fiscal invoice line.
type CreateLineRequest struct {
	ProductID   uuid.UUID
	Description string
	Quantity    float64
	UnitPrice   float64
	TaxRate     float64
	IsFiscal    bool // populated from product lookup; must be true
}

// CreateRequest is the full fiscal invoice creation payload.
type CreateRequest struct {
	CustomerName     string
	CustomerIDType   CustomerIDType
	CustomerIDNumber string
	Notes            string
	Lines            []CreateLineRequest
}

// Validate checks the create request for domain-level constraints.
func (r *CreateRequest) Validate() error {
	if len(r.Lines) == 0 {
		return ErrEmptyInvoice
	}
	for _, l := range r.Lines {
		if !l.IsFiscal {
			return ErrNonFiscalProductForbidden
		}
		if l.Quantity <= 0 {
			return fmt.Errorf("quantity must be positive")
		}
		if l.UnitPrice <= 0 {
			return fmt.Errorf("unit_price must be positive")
		}
		if l.TaxRate < 0 {
			return fmt.Errorf("tax_rate cannot be negative")
		}
	}
	return nil
}

// ComputeTotals returns (subtotal_base, tax_amount, total) for a set of lines.
func ComputeTotals(lines []CreateLineRequest) (subtotalBase, taxAmount, total float64) {
	for _, l := range lines {
		base := l.Quantity * l.UnitPrice
		tax := base * l.TaxRate
		subtotalBase += base
		taxAmount += tax
	}
	total = subtotalBase + taxAmount
	return
}

// Domain errors
var (
	ErrInvoiceNotFound           = errors.New("fiscal invoice not found")
	ErrInvoiceAlreadyIssued      = errors.New("fiscal invoice is already issued")
	ErrInvoiceAlreadyCancelled   = errors.New("fiscal invoice is already cancelled")
	ErrNonFiscalProductForbidden = errors.New("non-fiscal products cannot be added to a fiscal invoice")
	ErrEmptyInvoice              = errors.New("fiscal invoice must have at least one line")
	ErrInvoiceNotDraft           = errors.New("fiscal invoice is not in draft status")
	ErrInvoiceCannotRetry        = errors.New("fiscal invoice has exceeded the maximum retry attempts")
)
