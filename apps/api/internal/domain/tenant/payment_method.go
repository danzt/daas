package tenant

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// PaymentMethodType represents the kind of payment a tenant accepts.
type PaymentMethodType string

const (
	PaymentMethodPagoMovil    PaymentMethodType = "pago_movil"
	PaymentMethodTransferBank PaymentMethodType = "transfer_bank"
	PaymentMethodZelle        PaymentMethodType = "zelle"
	PaymentMethodPayPal       PaymentMethodType = "paypal"
	PaymentMethodUSDT         PaymentMethodType = "usdt"
	PaymentMethodCash         PaymentMethodType = "cash"
	PaymentMethodOther        PaymentMethodType = "other"
)

// PaymentMethod is a payment option the tenant accepts.
// Details is type-specific JSON — validated by Validate().
type PaymentMethod struct {
	ID        uuid.UUID         `json:"id"`
	TenantID  uuid.UUID         `json:"tenant_id"`
	Type      PaymentMethodType `json:"type"`
	Label     string            `json:"label"`
	Details   map[string]any    `json:"details"`
	Currency  string            `json:"currency"`
	Active    bool              `json:"active"`
	SortOrder int               `json:"sort_order"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// CreatePaymentMethodRequest carries the data needed to create a payment method.
type CreatePaymentMethodRequest struct {
	Type      PaymentMethodType
	Label     string
	Details   map[string]any
	Currency  string
	SortOrder int
}

// UpdatePaymentMethodRequest carries the optional fields to update.
type UpdatePaymentMethodRequest struct {
	Label     *string
	Details   *map[string]any
	Currency  *string
	Active    *bool
	SortOrder *int
}

// Validate enforces type-specific required fields in Details.
func (r *CreatePaymentMethodRequest) Validate() error {
	if !isValidType(r.Type) {
		return ErrInvalidPaymentMethodType
	}
	return validateDetailsForType(r.Type, r.Details)
}

func isValidType(t PaymentMethodType) bool {
	switch t {
	case PaymentMethodPagoMovil, PaymentMethodTransferBank, PaymentMethodZelle,
		PaymentMethodPayPal, PaymentMethodUSDT, PaymentMethodCash, PaymentMethodOther:
		return true
	}
	return false
}

// validateDetailsForType checks the required fields per payment type.
// Only fields critical for the customer to pay are required.
func validateDetailsForType(t PaymentMethodType, d map[string]any) error {
	requiredByType := map[PaymentMethodType][]string{
		PaymentMethodPagoMovil:    {"bank", "document_number", "phone"},
		PaymentMethodTransferBank: {"bank", "account_number", "account_holder"},
		PaymentMethodZelle:        {"email", "account_holder"},
		PaymentMethodPayPal:       {"email"},
		PaymentMethodUSDT:         {"wallet_address", "network"},
		PaymentMethodCash:         {}, // no required details — notes optional
		PaymentMethodOther:        {"notes"},
	}
	required, ok := requiredByType[t]
	if !ok {
		return ErrInvalidPaymentMethodType
	}
	for _, field := range required {
		v, present := d[field]
		if !present {
			return &MissingFieldError{Field: field}
		}
		if s, isStr := v.(string); isStr && s == "" {
			return &MissingFieldError{Field: field}
		}
	}
	return nil
}

// Domain errors for payment methods.
var (
	ErrInvalidPaymentMethodType = errors.New("invalid payment method type")
	ErrPaymentMethodNotFound    = errors.New("payment method not found")
)

// MissingFieldError is returned when a required details field is absent or empty.
type MissingFieldError struct {
	Field string
}

func (e *MissingFieldError) Error() string {
	return "missing required field: " + e.Field
}
