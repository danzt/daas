ALTER TABLE purchase_orders
  DROP COLUMN IF EXISTS supplier_invoice_number,
  DROP COLUMN IF EXISTS supplier_invoice_date,
  DROP COLUMN IF EXISTS supplier_invoice_tax_base,
  DROP COLUMN IF EXISTS supplier_invoice_tax_amount;
