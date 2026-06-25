package repository

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/oms-saas/oms-saas-go/internal/apperror"
	"github.com/oms-saas/oms-saas-go/internal/models"
	"golang.org/x/crypto/bcrypt"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

var slugRe = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = slugRe.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

func (r *Repository) RegisterTenant(ctx context.Context, tenantName, email, password, fullName string) (*models.User, *models.Tenant, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback(ctx)

	slug := slugify(tenantName)
	if slug == "" {
		slug = "tenant"
	}

	var tenant models.Tenant
	err = tx.QueryRow(ctx, `
		INSERT INTO tenants (name, slug) VALUES ($1, $2 || '-' || substr(gen_random_uuid()::text, 1, 8))
		RETURNING id, name, slug, status, created_at, updated_at
	`, tenantName, slug).Scan(&tenant.ID, &tenant.Name, &tenant.Slug, &tenant.Status, &tenant.CreatedAt, &tenant.UpdatedAt)
	if err != nil {
		return nil, nil, fmt.Errorf("create tenant: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, err
	}

	var user models.User
	err = tx.QueryRow(ctx, `
		INSERT INTO users (tenant_id, email, password_hash, full_name, role)
		VALUES ($1, $2, $3, $4, 'owner')
		RETURNING id, tenant_id, email, password_hash, full_name, role, created_at, updated_at
	`, tenant.ID, strings.ToLower(email), string(hash), fullName).Scan(
		&user.ID, &user.TenantID, &user.Email, &user.PasswordHash, &user.FullName, &user.Role, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("create user: %w", err)
	}

	var planID uuid.UUID
	err = tx.QueryRow(ctx, `SELECT id FROM plans WHERE slug = 'free' LIMIT 1`).Scan(&planID)
	if err != nil {
		return nil, nil, fmt.Errorf("get free plan: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO tenant_subscriptions (tenant_id, plan_id, status)
		VALUES ($1, $2, 'active')
	`, tenant.ID, planID)
	if err != nil {
		return nil, nil, fmt.Errorf("create subscription: %w", err)
	}

	if _, err = tx.Exec(ctx, `SELECT set_config('app.tenant_id', $1, true)`, tenant.ID.String()); err != nil {
		return nil, nil, fmt.Errorf("set tenant context: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO warehouses (tenant_id, name, is_default) VALUES ($1, 'Main Warehouse', true)
	`, tenant.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("create warehouse: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, nil, err
	}
	return &user, &tenant, nil
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var u models.User
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, email, password_hash, full_name, role, created_at, updated_at
		FROM users WHERE email = $1 LIMIT 1
	`, strings.ToLower(email)).Scan(
		&u.ID, &u.TenantID, &u.Email, &u.PasswordHash, &u.FullName, &u.Role, &u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.ErrNotFound
	}
	return &u, err
}

func (r *Repository) GetUserByID(ctx context.Context, tenantID, userID uuid.UUID) (*models.User, error) {
	var u models.User
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, email, password_hash, full_name, role, created_at, updated_at
		FROM users WHERE id = $1 AND tenant_id = $2
	`, userID, tenantID).Scan(
		&u.ID, &u.TenantID, &u.Email, &u.PasswordHash, &u.FullName, &u.Role, &u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.ErrNotFound
	}
	return &u, err
}

func (r *Repository) SaveRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)
	`, userID, tokenHash, expiresAt)
	return err
}

func (r *Repository) GetRefreshToken(ctx context.Context, tokenHash string) (*models.User, error) {
	var u models.User
	err := r.pool.QueryRow(ctx, `
		SELECT u.id, u.tenant_id, u.email, u.password_hash, u.full_name, u.role, u.created_at, u.updated_at
		FROM refresh_tokens rt
		JOIN users u ON u.id = rt.user_id
		WHERE rt.token_hash = $1 AND rt.expires_at > NOW()
	`, tokenHash).Scan(
		&u.ID, &u.TenantID, &u.Email, &u.PasswordHash, &u.FullName, &u.Role, &u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.ErrNotFound
	}
	return &u, err
}

func (r *Repository) DeleteRefreshToken(ctx context.Context, tokenHash string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM refresh_tokens WHERE token_hash = $1`, tokenHash)
	return err
}

func (r *Repository) GetTenant(ctx context.Context, tenantID uuid.UUID) (*models.Tenant, error) {
	var t models.Tenant
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, slug, status, created_at, updated_at FROM tenants WHERE id = $1
	`, tenantID).Scan(&t.ID, &t.Name, &t.Slug, &t.Status, &t.CreatedAt, &t.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.ErrNotFound
	}
	return &t, err
}

func (r *Repository) UpdateTenant(ctx context.Context, tenantID uuid.UUID, name string) (*models.Tenant, error) {
	var t models.Tenant
	err := r.pool.QueryRow(ctx, `
		UPDATE tenants SET name = $2, updated_at = NOW() WHERE id = $1
		RETURNING id, name, slug, status, created_at, updated_at
	`, tenantID, name).Scan(&t.ID, &t.Name, &t.Slug, &t.Status, &t.CreatedAt, &t.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.ErrNotFound
	}
	return &t, err
}

func (r *Repository) GetTenantSubscription(ctx context.Context, tenantID uuid.UUID) (*models.TenantSubscription, error) {
	var sub models.TenantSubscription
	var plan models.Plan
	err := r.pool.QueryRow(ctx, `
		SELECT ts.id, ts.tenant_id, ts.plan_id, ts.status, ts.stripe_customer_id, ts.stripe_subscription_id,
		       p.id, p.name, p.slug, p.max_products, p.max_orders_per_month, p.max_users, p.price_cents, p.stripe_price_id
		FROM tenant_subscriptions ts
		JOIN plans p ON p.id = ts.plan_id
		WHERE ts.tenant_id = $1
	`, tenantID).Scan(
		&sub.ID, &sub.TenantID, &sub.PlanID, &sub.Status, &sub.StripeCustomerID, &sub.StripeSubscriptionID,
		&plan.ID, &plan.Name, &plan.Slug, &plan.MaxProducts, &plan.MaxOrdersPerMonth, &plan.MaxUsers, &plan.PriceCents, &plan.StripePriceID,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.ErrNotFound
	}
	sub.Plan = &plan
	return &sub, err
}

func (r *Repository) CountProducts(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM products WHERE tenant_id = $1`, tenantID).Scan(&count)
	return count, err
}

func (r *Repository) CountOrdersThisMonth(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM orders
		WHERE tenant_id = $1 AND created_at >= date_trunc('month', NOW())
	`, tenantID).Scan(&count)
	return count, err
}

func (r *Repository) CountUsers(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE tenant_id = $1`, tenantID).Scan(&count)
	return count, err
}

func (r *Repository) CheckPlanLimit(ctx context.Context, tenantID uuid.UUID, resource string) error {
	sub, err := r.GetTenantSubscription(ctx, tenantID)
	if err != nil {
		return err
	}
	if sub.Plan == nil {
		return nil
	}
	switch resource {
	case "product":
		count, err := r.CountProducts(ctx, tenantID)
		if err != nil {
			return err
		}
		if count >= int64(sub.Plan.MaxProducts) {
			return apperror.ErrLimitExceeded
		}
	case "order":
		count, err := r.CountOrdersThisMonth(ctx, tenantID)
		if err != nil {
			return err
		}
		if count >= int64(sub.Plan.MaxOrdersPerMonth) {
			return apperror.ErrLimitExceeded
		}
	case "user":
		count, err := r.CountUsers(ctx, tenantID)
		if err != nil {
			return err
		}
		if count >= int64(sub.Plan.MaxUsers) {
			return apperror.ErrLimitExceeded
		}
	}
	return nil
}

func (r *Repository) ListPlans(ctx context.Context) ([]models.Plan, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, slug, max_products, max_orders_per_month, max_users, price_cents, stripe_price_id
		FROM plans ORDER BY price_cents
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	plans := make([]models.Plan, 0)
	for rows.Next() {
		var p models.Plan
		if err := rows.Scan(&p.ID, &p.Name, &p.Slug, &p.MaxProducts, &p.MaxOrdersPerMonth, &p.MaxUsers, &p.PriceCents, &p.StripePriceID); err != nil {
			return nil, err
		}
		plans = append(plans, p)
	}
	return plans, rows.Err()
}

func (r *Repository) UpdateTenantSubscriptionStatus(ctx context.Context, stripeSubID, status string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE tenant_subscriptions SET status = $2, updated_at = NOW()
		WHERE stripe_subscription_id = $1
	`, stripeSubID, status)
	return err
}

func (r *Repository) DowngradeTenantToFree(ctx context.Context, stripeSubID string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE tenant_subscriptions ts
		SET plan_id = p.id, status = 'active', stripe_subscription_id = NULL, updated_at = NOW()
		FROM plans p
		WHERE ts.stripe_subscription_id = $1 AND p.slug = 'free'
	`, stripeSubID)
	return err
}

func (r *Repository) GetTenantIDByStripeCustomer(ctx context.Context, customerID string) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `
		SELECT tenant_id FROM tenant_subscriptions WHERE stripe_customer_id = $1
	`, customerID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, apperror.ErrNotFound
	}
	return id, err
}

func (r *Repository) UpdateTenantSubscriptionPlan(ctx context.Context, tenantID uuid.UUID, planSlug, stripeCustomerID, stripeSubID string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE tenant_subscriptions ts
		SET plan_id = p.id,
		    stripe_customer_id = COALESCE(NULLIF($3, ''), ts.stripe_customer_id),
		    stripe_subscription_id = COALESCE(NULLIF($4, ''), ts.stripe_subscription_id),
		    status = 'active',
		    updated_at = NOW()
		FROM plans p
		WHERE ts.tenant_id = $1 AND p.slug = $2
	`, tenantID, planSlug, stripeCustomerID, stripeSubID)
	return err
}

func VerifyPassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
