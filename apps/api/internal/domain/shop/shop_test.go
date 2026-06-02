package shop_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/danzt/daas/api/internal/domain/shop"
)

// TestShopProduct_SensitiveFieldsAbsent verifies that marshalling a ShopProduct
// to JSON does NOT expose cost_price, supplier_id, fiscal_id, or created_by.
// These fields belong to internal admin views only.
func TestShopProduct_SensitiveFieldsAbsent(t *testing.T) {
	price := 99.99
	fiscalPrice := 105.00
	category := "Electronics"

	p := shop.ShopProduct{
		ID:          uuid.New(),
		TenantID:    uuid.New(),
		Name:        "Laptop Pro",
		Description: "High-end laptop",
		Price:       price,
		FiscalPrice: &fiscalPrice,
		Category:    &category,
		StockQty:    10,
		IsFiscal:    true,
	}

	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("json.Marshal(ShopProduct): %v", err)
	}

	jsonStr := string(data)

	sensitiveFields := []string{"cost_price", "supplier_id", "fiscal_id", "created_by"}
	for _, field := range sensitiveFields {
		if strings.Contains(jsonStr, field) {
			t.Errorf("ShopProduct JSON must NOT contain %q, but it does.\nJSON: %s", field, jsonStr)
		}
	}
}

// TestShopProduct_PublicFieldsPresent verifies that the required public fields
// are present in the JSON output.
func TestShopProduct_PublicFieldsPresent(t *testing.T) {
	p := shop.ShopProduct{
		ID:       uuid.New(),
		TenantID: uuid.New(),
		Name:     "Test Product",
		Price:    50.0,
		StockQty: 5,
		IsFiscal: false,
	}

	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("json.Marshal(ShopProduct): %v", err)
	}

	jsonStr := string(data)

	requiredFields := []string{"id", "tenant_id", "name", "price", "stock_qty", "is_fiscal"}
	for _, field := range requiredFields {
		if !strings.Contains(jsonStr, field) {
			t.Errorf("ShopProduct JSON must contain %q, but it does not.\nJSON: %s", field, jsonStr)
		}
	}
}

// TestShopProduct_OptionalFieldsOmitEmpty verifies that nil optional fields
// (fiscal_price, category) are omitted from JSON.
func TestShopProduct_OptionalFieldsOmitEmpty(t *testing.T) {
	p := shop.ShopProduct{
		ID:       uuid.New(),
		TenantID: uuid.New(),
		Name:     "Simple Product",
		Price:    25.0,
		StockQty: 3,
		IsFiscal: false,
		// FiscalPrice and Category intentionally nil
	}

	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("json.Marshal(ShopProduct): %v", err)
	}

	jsonStr := string(data)

	if strings.Contains(jsonStr, "fiscal_price") {
		t.Errorf("fiscal_price should be omitted when nil, but is present.\nJSON: %s", jsonStr)
	}
	if strings.Contains(jsonStr, `"category"`) {
		t.Errorf("category should be omitted when nil, but is present.\nJSON: %s", jsonStr)
	}
}
