// Package product contains the domain entities and business rules for product
// management. This package has ZERO imports from Echo, pgx, or any external
// framework — it is pure Go.
package product

import "errors"

// Domain errors for the product module. These are pure sentinel errors —
// no HTTP status codes or framework types here. HTTP translation happens
// in the handler layer.

var (
	// ErrFiscalPriceMissing is returned when a fiscal product has no fiscal_price
	// or the price is zero/negative.
	ErrFiscalPriceMissing = errors.New("fiscal product must have a positive fiscal_price")

	// ErrInternalPriceMissing is returned when a non-fiscal product has no
	// internal_price or the price is zero/negative.
	ErrInternalPriceMissing = errors.New("internal product must have a positive internal_price")

	// ErrTaxRateMissing is returned when a fiscal product has no tax_rate.
	ErrTaxRateMissing = errors.New("fiscal product must have a tax_rate")

	// ErrTaxRateOnInternal is returned when a non-fiscal product has a tax_rate
	// set. Tax rates only apply to fiscal products.
	ErrTaxRateOnInternal = errors.New("internal product must not have a tax_rate")

	// ErrZeroPrice is returned when a price field is explicitly set to zero.
	ErrZeroPrice = errors.New("price must be greater than zero")

	// ErrProductNotFound is returned when a product lookup yields no results.
	ErrProductNotFound = errors.New("product not found")

	// ErrDuplicateSKU is returned when a product with the same SKU already
	// exists within the tenant.
	ErrDuplicateSKU = errors.New("a product with this SKU already exists")
)
