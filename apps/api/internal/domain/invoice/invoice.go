// Package invoice contains the domain entities and business rules for
// internal (non-fiscal) invoice management. Zero imports from Echo, pgx,
// or any external framework.
package invoice

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Status represents the lifecycle state of an internal invoice.
type Status string

const (
	StatusDraft     Status = "draft"
	StatusIssued    Status = "issued"
	StatusCancelled Status = "cancelled"
)

// CustomerIDType classifies the customer identification document.
type CustomerIDType string

const (
	CustomerIDTypeCedula    CustomerIDType = "cedula"
	CustomerIDTypeRIF       CustomerIDType = "rif"
	CustomerIDTypePassport  CustomerIDType = "passport"
	CustomerIDTypeAnonymous CustomerIDType = "anonymous"
)

// InternalInvoice is the aggregate root for a non-fiscal sale document.
type InternalInvoice struct {
	ID               uuid.UUID
	TenantID         uuid.UUID
	Correlative      string // empty while status=draft
	CustomerName     string
	CustomerIDType   CustomerIDType
	CustomerIDNumber string
	Status           Status
	Subtotal         float64
	Total            float64
	Notes            string
	IssuedAt         *time.Time
	CreatedBy        uuid.UUID
	CreatedAt        time.Time
	UpdatedAt        time.Time
	Lines            []InvoiceLine
}

// InvoiceLine is a single product line inside an invoice.
type InvoiceLine struct {
	ID          uuid.UUID
	InvoiceID   uuid.UUID
	ProductID   uuid.UUID
	Description string
	Quantity    float64
	UnitPrice   float64
	Subtotal    float64
	SortOrder   int
}

// CreateLineRequest is the input for one invoice line.
type CreateLineRequest struct {
	ProductID   uuid.UUID
	Description string
	Quantity    float64
	UnitPrice   float64
	IsFiscal    bool // populated from product lookup, validated by service
}

// CreateRequest is the full invoice creation payload.
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
		if l.IsFiscal {
			return ErrFiscalProductForbidden
		}
		if l.Quantity <= 0 {
			return fmt.Errorf("quantity must be positive")
		}
		if l.UnitPrice <= 0 {
			return fmt.Errorf("unit_price must be positive")
		}
	}
	return nil
}

// FormatCorrelative builds the display correlative string.
func FormatCorrelative(year, seq int) string {
	return fmt.Sprintf("INT-%d-%05d", year, seq)
}

// Domain errors
var (
	ErrInvoiceNotFound         = errors.New("invoice not found")
	ErrInvoiceAlreadyIssued    = errors.New("invoice is already issued")
	ErrInvoiceAlreadyCancelled = errors.New("invoice is already cancelled")
	ErrFiscalProductForbidden  = errors.New("fiscal products cannot be added to an internal invoice")
	ErrEmptyInvoice            = errors.New("invoice must have at least one line")
	ErrInvoiceNotDraft         = errors.New("invoice is not in draft status")
)
