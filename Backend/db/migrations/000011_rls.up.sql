-- Enable RLS on tenant-scoped tables (defense in depth; app still filters by tenant_id)
ALTER TABLE products ENABLE ROW LEVEL SECURITY;
ALTER TABLE categories ENABLE ROW LEVEL SECURITY;
ALTER TABLE customers ENABLE ROW LEVEL SECURITY;
ALTER TABLE orders ENABLE ROW LEVEL SECURITY;
ALTER TABLE order_items ENABLE ROW LEVEL SECURITY;
ALTER TABLE warehouses ENABLE ROW LEVEL SECURITY;
ALTER TABLE stock_levels ENABLE ROW LEVEL SECURITY;
ALTER TABLE inventory_movements ENABLE ROW LEVEL SECURITY;
ALTER TABLE notifications ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_products ON products
    USING (tenant_id::text = current_setting('app.tenant_id', true));
CREATE POLICY tenant_isolation_categories ON categories
    USING (tenant_id::text = current_setting('app.tenant_id', true));
CREATE POLICY tenant_isolation_customers ON customers
    USING (tenant_id::text = current_setting('app.tenant_id', true));
CREATE POLICY tenant_isolation_orders ON orders
    USING (tenant_id::text = current_setting('app.tenant_id', true));
CREATE POLICY tenant_isolation_order_items ON order_items
    USING (tenant_id::text = current_setting('app.tenant_id', true));
CREATE POLICY tenant_isolation_warehouses ON warehouses
    USING (tenant_id::text = current_setting('app.tenant_id', true));
CREATE POLICY tenant_isolation_stock_levels ON stock_levels
    USING (tenant_id::text = current_setting('app.tenant_id', true));
CREATE POLICY tenant_isolation_inventory_movements ON inventory_movements
    USING (tenant_id::text = current_setting('app.tenant_id', true));
CREATE POLICY tenant_isolation_notifications ON notifications
    USING (tenant_id::text = current_setting('app.tenant_id', true));
