package product_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/danzt/daas/api/internal/domain/product"
)

func ptr[T any](v T) *T { return &v }

func TestProduct_Validate_FiscalWithoutPrice(t *testing.T) {
	p := product.Product{
		ID:       uuid.New(),
		TenantID: uuid.New(),
		Name:     "Producto Fiscal",
		IsFiscal: true,
		TaxRate:  ptr(16.0),
		// FiscalPrice intentionally omitted
	}
	if err := p.Validate(); err == nil {
		t.Fatal("expected error for fiscal product without price, got nil")
	} else if err != product.ErrFiscalPriceMissing {
		t.Fatalf("expected ErrFiscalPriceMissing, got %v", err)
	}
}

func TestProduct_Validate_FiscalWithZeroPrice(t *testing.T) {
	p := product.Product{
		ID:          uuid.New(),
		TenantID:    uuid.New(),
		Name:        "Producto Fiscal Cero",
		IsFiscal:    true,
		FiscalPrice: ptr(0.0),
		TaxRate:     ptr(16.0),
	}
	if err := p.Validate(); err != product.ErrFiscalPriceMissing {
		t.Fatalf("expected ErrFiscalPriceMissing for zero price, got %v", err)
	}
}

func TestProduct_Validate_FiscalWithoutTaxRate(t *testing.T) {
	p := product.Product{
		ID:          uuid.New(),
		TenantID:    uuid.New(),
		Name:        "Producto Fiscal Sin IVA",
		IsFiscal:    true,
		FiscalPrice: ptr(100.0),
		// TaxRate intentionally omitted
	}
	if err := p.Validate(); err == nil {
		t.Fatal("expected error for fiscal product without tax_rate, got nil")
	} else if err != product.ErrTaxRateMissing {
		t.Fatalf("expected ErrTaxRateMissing, got %v", err)
	}
}

func TestProduct_Validate_InternalWithoutPrice(t *testing.T) {
	p := product.Product{
		ID:       uuid.New(),
		TenantID: uuid.New(),
		Name:     "Producto Interno Sin Precio",
		IsFiscal: false,
		// InternalPrice intentionally omitted
	}
	if err := p.Validate(); err == nil {
		t.Fatal("expected error for internal product without price, got nil")
	} else if err != product.ErrInternalPriceMissing {
		t.Fatalf("expected ErrInternalPriceMissing, got %v", err)
	}
}

func TestProduct_Validate_InternalWithTaxRate(t *testing.T) {
	p := product.Product{
		ID:            uuid.New(),
		TenantID:      uuid.New(),
		Name:          "Producto Interno Con IVA",
		IsFiscal:      false,
		InternalPrice: ptr(50.0),
		TaxRate:       ptr(16.0), // must NOT be set on internal products
	}
	if err := p.Validate(); err == nil {
		t.Fatal("expected error for internal product with tax_rate, got nil")
	} else if err != product.ErrTaxRateOnInternal {
		t.Fatalf("expected ErrTaxRateOnInternal, got %v", err)
	}
}

func TestProduct_Validate_ValidFiscalProduct(t *testing.T) {
	p := product.Product{
		ID:          uuid.New(),
		TenantID:    uuid.New(),
		Name:        "Producto Fiscal Válido",
		IsFiscal:    true,
		FiscalPrice: ptr(100.0),
		TaxRate:     ptr(16.0),
	}
	if err := p.Validate(); err != nil {
		t.Fatalf("expected nil for valid fiscal product, got %v", err)
	}
}

func TestProduct_Validate_ValidInternalProduct(t *testing.T) {
	p := product.Product{
		ID:            uuid.New(),
		TenantID:      uuid.New(),
		Name:          "Producto Interno Válido",
		IsFiscal:      false,
		InternalPrice: ptr(50.0),
	}
	if err := p.Validate(); err != nil {
		t.Fatalf("expected nil for valid internal product, got %v", err)
	}
}
