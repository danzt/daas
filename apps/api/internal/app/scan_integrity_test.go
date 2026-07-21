package app

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Regresión del bug de prod "number of field descriptions must equal number of
// destinations (11 vs 15)": si una SELECT y su scan divergen en columnas, estas
// llamadas explotan en runtime aunque compile. Corre solo con TEST_DATABASE_URL
// seteada (una DB con el schema migrado y al menos un supplier/PO).
func TestSupplierPOScanIntegrity(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	var tenantID, poID, supID uuid.UUID
	if err := pool.QueryRow(ctx,
		`SELECT tenant_id, id, supplier_id FROM purchase_orders LIMIT 1`,
	).Scan(&tenantID, &poID, &supID); err != nil {
		t.Skipf("no purchase orders in test DB: %v", err)
	}

	svc := NewSupplierService(pool)
	if _, err := svc.GetPO(ctx, tenantID, poID); err != nil {
		t.Errorf("GetPO: %v", err)
	}
	if _, err := svc.ListPOs(ctx, tenantID, nil, nil); err != nil {
		t.Errorf("ListPOs: %v", err)
	}
	if _, err := svc.GetSupplier(ctx, tenantID, supID); err != nil {
		t.Errorf("GetSupplier: %v", err)
	}
	if _, err := svc.ListSuppliers(ctx, tenantID, false); err != nil {
		t.Errorf("ListSuppliers: %v", err)
	}
	if _, err := svc.ListCatalog(ctx, supID); err != nil {
		t.Errorf("ListCatalog: %v", err)
	}
}
