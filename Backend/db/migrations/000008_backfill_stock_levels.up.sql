INSERT INTO stock_levels (tenant_id, warehouse_id, product_id, quantity, updated_at)
SELECT p.tenant_id, w.id, p.id, p.stock_on_hand, NOW()
FROM products p
JOIN warehouses w ON w.tenant_id = p.tenant_id AND w.is_default = true
WHERE p.stock_on_hand > 0
ON CONFLICT (warehouse_id, product_id) DO UPDATE
SET quantity = EXCLUDED.quantity, updated_at = NOW();
