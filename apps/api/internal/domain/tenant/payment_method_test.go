package tenant_test

import (
	"errors"
	"testing"

	"github.com/danzt/daas/api/internal/domain/tenant"
)

// ─── Pago Móvil ──────────────────────────────────────────────────────────────

func TestValidate_PagoMovil_HappyPath(t *testing.T) {
	req := tenant.CreatePaymentMethodRequest{
		Type: tenant.PaymentMethodPagoMovil,
		Details: map[string]any{
			"bank":            "Banesco",
			"document_number": "V-12345678",
			"phone":           "0412-1234567",
		},
	}
	if err := req.Validate(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestValidate_PagoMovil_MissingBank(t *testing.T) {
	req := tenant.CreatePaymentMethodRequest{
		Type: tenant.PaymentMethodPagoMovil,
		Details: map[string]any{
			"document_number": "V-12345678",
			"phone":           "0412-1234567",
		},
	}
	err := req.Validate()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var mfe *tenant.MissingFieldError
	if !errors.As(err, &mfe) {
		t.Fatalf("expected *MissingFieldError, got %T: %v", err, err)
	}
	if mfe.Field != "bank" {
		t.Fatalf("expected field=bank, got %q", mfe.Field)
	}
}

// ─── Transfer Bank ───────────────────────────────────────────────────────────

func TestValidate_TransferBank_HappyPath(t *testing.T) {
	req := tenant.CreatePaymentMethodRequest{
		Type: tenant.PaymentMethodTransferBank,
		Details: map[string]any{
			"bank":           "Mercantil",
			"account_number": "01050123456789012345",
			"account_holder": "Empresa XYZ C.A.",
		},
	}
	if err := req.Validate(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestValidate_TransferBank_MissingAccountNumber(t *testing.T) {
	req := tenant.CreatePaymentMethodRequest{
		Type: tenant.PaymentMethodTransferBank,
		Details: map[string]any{
			"bank":           "Mercantil",
			"account_holder": "Empresa XYZ C.A.",
		},
	}
	err := req.Validate()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var mfe *tenant.MissingFieldError
	if !errors.As(err, &mfe) {
		t.Fatalf("expected *MissingFieldError, got %T: %v", err, err)
	}
	if mfe.Field != "account_number" {
		t.Fatalf("expected field=account_number, got %q", mfe.Field)
	}
}

// ─── Zelle ───────────────────────────────────────────────────────────────────

func TestValidate_Zelle_HappyPath(t *testing.T) {
	req := tenant.CreatePaymentMethodRequest{
		Type: tenant.PaymentMethodZelle,
		Details: map[string]any{
			"email":          "payments@example.com",
			"account_holder": "John Doe",
		},
	}
	if err := req.Validate(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestValidate_Zelle_MissingEmail(t *testing.T) {
	req := tenant.CreatePaymentMethodRequest{
		Type: tenant.PaymentMethodZelle,
		Details: map[string]any{
			"account_holder": "John Doe",
		},
	}
	err := req.Validate()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var mfe *tenant.MissingFieldError
	if !errors.As(err, &mfe) {
		t.Fatalf("expected *MissingFieldError, got %T: %v", err, err)
	}
	if mfe.Field != "email" {
		t.Fatalf("expected field=email, got %q", mfe.Field)
	}
}

// ─── PayPal ──────────────────────────────────────────────────────────────────

func TestValidate_PayPal_RequiresEmail(t *testing.T) {
	// missing email → error
	req := tenant.CreatePaymentMethodRequest{
		Type:    tenant.PaymentMethodPayPal,
		Details: map[string]any{},
	}
	err := req.Validate()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var mfe *tenant.MissingFieldError
	if !errors.As(err, &mfe) {
		t.Fatalf("expected *MissingFieldError, got %T: %v", err, err)
	}
	if mfe.Field != "email" {
		t.Fatalf("expected field=email, got %q", mfe.Field)
	}

	// with email → ok
	req.Details["email"] = "me@paypal.com"
	if err := req.Validate(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// ─── USDT ─────────────────────────────────────────────────────────────────────

func TestValidate_USDT_RequiresWalletAndNetwork(t *testing.T) {
	// missing both
	req := tenant.CreatePaymentMethodRequest{
		Type:    tenant.PaymentMethodUSDT,
		Details: map[string]any{},
	}
	err := req.Validate()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// with both → ok
	req.Details["wallet_address"] = "TXYZ1234"
	req.Details["network"] = "TRC20"
	if err := req.Validate(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// ─── Cash ─────────────────────────────────────────────────────────────────────

func TestValidate_Cash_NoRequiredFields(t *testing.T) {
	req := tenant.CreatePaymentMethodRequest{
		Type:    tenant.PaymentMethodCash,
		Details: map[string]any{},
	}
	if err := req.Validate(); err != nil {
		t.Fatalf("expected no error for cash with empty details, got %v", err)
	}
}

// ─── Other ────────────────────────────────────────────────────────────────────

func TestValidate_Other_RequiresNotes(t *testing.T) {
	// missing notes → error
	req := tenant.CreatePaymentMethodRequest{
		Type:    tenant.PaymentMethodOther,
		Details: map[string]any{},
	}
	err := req.Validate()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var mfe *tenant.MissingFieldError
	if !errors.As(err, &mfe) {
		t.Fatalf("expected *MissingFieldError, got %T: %v", err, err)
	}
	if mfe.Field != "notes" {
		t.Fatalf("expected field=notes, got %q", mfe.Field)
	}

	// with notes → ok
	req.Details["notes"] = "Pago en efectivo al recibir"
	if err := req.Validate(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// ─── Invalid type ─────────────────────────────────────────────────────────────

func TestValidate_InvalidType(t *testing.T) {
	req := tenant.CreatePaymentMethodRequest{
		Type:    tenant.PaymentMethodType("bitcoin_lightning"),
		Details: map[string]any{},
	}
	err := req.Validate()
	if !errors.Is(err, tenant.ErrInvalidPaymentMethodType) {
		t.Fatalf("expected ErrInvalidPaymentMethodType, got %v", err)
	}
}

// ─── Empty string treated as missing ─────────────────────────────────────────

func TestValidate_EmptyStringTreatedAsMissing(t *testing.T) {
	req := tenant.CreatePaymentMethodRequest{
		Type: tenant.PaymentMethodPagoMovil,
		Details: map[string]any{
			"bank":            "", // empty string — must be treated as missing
			"document_number": "V-12345678",
			"phone":           "0412-1234567",
		},
	}
	err := req.Validate()
	if err == nil {
		t.Fatal("expected error for empty bank, got nil")
	}
	var mfe *tenant.MissingFieldError
	if !errors.As(err, &mfe) {
		t.Fatalf("expected *MissingFieldError, got %T: %v", err, err)
	}
	if mfe.Field != "bank" {
		t.Fatalf("expected field=bank, got %q", mfe.Field)
	}
}
