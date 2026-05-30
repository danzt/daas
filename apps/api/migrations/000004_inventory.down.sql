DROP TRIGGER IF EXISTS initialize_product_stock ON products;
DROP FUNCTION IF EXISTS init_product_stock();
DROP TABLE IF EXISTS inventory_movements;
DROP TABLE IF EXISTS product_stock;
