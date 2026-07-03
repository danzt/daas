-- Fase 1 — Portal de Proveedores: registrar la factura de compra en la OC.
-- Permite declarar la compra ante el fisco: número de factura del proveedor,
-- fecha, base imponible e IVA (crédito fiscal). Todos nullable — la factura se
-- registra cuando llega, no es obligatoria para crear la OC.
ALTER TABLE purchase_orders
  ADD COLUMN supplier_invoice_number     varchar,
  ADD COLUMN supplier_invoice_date       date,
  ADD COLUMN supplier_invoice_tax_base   numeric,
  ADD COLUMN supplier_invoice_tax_amount numeric;
