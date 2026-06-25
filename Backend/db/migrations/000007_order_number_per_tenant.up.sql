ALTER TABLE tenants ADD COLUMN IF NOT EXISTS order_seq BIGINT NOT NULL DEFAULT 1000;

-- Migrate existing sequence value to tenants that have orders
UPDATE tenants t
SET order_seq = GREATEST(t.order_seq, (
    SELECT COALESCE(MAX(
        NULLIF(regexp_replace(o.order_number, '.*-', ''), '')::bigint
    ), 1000)
    FROM orders o WHERE o.tenant_id = t.id
));

DROP SEQUENCE IF EXISTS order_number_seq;
