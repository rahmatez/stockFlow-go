CREATE TABLE plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    slug VARCHAR(50) NOT NULL UNIQUE,
    max_products INT NOT NULL DEFAULT 50,
    max_orders_per_month INT NOT NULL DEFAULT 100,
    max_users INT NOT NULL DEFAULT 3,
    price_cents INT NOT NULL DEFAULT 0,
    stripe_price_id VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE tenant_subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL UNIQUE REFERENCES tenants(id) ON DELETE CASCADE,
    plan_id UUID NOT NULL REFERENCES plans(id),
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    stripe_customer_id VARCHAR(255),
    stripe_subscription_id VARCHAR(255),
    current_period_start TIMESTAMPTZ,
    current_period_end TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO plans (name, slug, max_products, max_orders_per_month, max_users, price_cents) VALUES
    ('Free', 'free', 50, 100, 3, 0),
    ('Starter', 'starter', 500, 1000, 10, 2900),
    ('Pro', 'pro', 5000, 10000, 50, 9900);
