package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/oms-saas/oms-saas-go/internal/apperror"
	"github.com/oms-saas/oms-saas-go/internal/models"
)

func (r *Repository) ListCustomers(ctx context.Context, tenantID uuid.UUID, p ListParams) ([]models.Customer, int64, error) {
	p = p.Normalize()
	search := "%" + p.Search + "%"

	var total int64
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM customers
		WHERE tenant_id = $1 AND ($2 = '' OR name ILIKE $3 OR COALESCE(email,'') ILIKE $3)
	`, tenantID, p.Search, search).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, name, email, phone, created_at, updated_at
		FROM customers
		WHERE tenant_id = $1 AND ($2 = '' OR name ILIKE $3 OR COALESCE(email,'') ILIKE $3)
		ORDER BY created_at DESC
		LIMIT $4 OFFSET $5
	`, tenantID, p.Search, search, p.Limit, p.Offset())
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]models.Customer, 0)
	for rows.Next() {
		var c models.Customer
		if err := rows.Scan(&c.ID, &c.TenantID, &c.Name, &c.Email, &c.Phone, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, c)
	}
	return items, total, rows.Err()
}

func (r *Repository) GetCustomer(ctx context.Context, tenantID, id uuid.UUID) (*models.Customer, error) {
	var c models.Customer
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, name, email, phone, created_at, updated_at
		FROM customers WHERE id = $1 AND tenant_id = $2
	`, id, tenantID).Scan(&c.ID, &c.TenantID, &c.Name, &c.Email, &c.Phone, &c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.ErrNotFound
	}
	return &c, err
}

type CustomerInput struct {
	Name  string
	Email *string
	Phone *string
}

func (r *Repository) CreateCustomer(ctx context.Context, tenantID uuid.UUID, in CustomerInput) (*models.Customer, error) {
	var c models.Customer
	err := r.pool.QueryRow(ctx, `
		INSERT INTO customers (tenant_id, name, email, phone) VALUES ($1, $2, $3, $4)
		RETURNING id, tenant_id, name, email, phone, created_at, updated_at
	`, tenantID, in.Name, in.Email, in.Phone).Scan(
		&c.ID, &c.TenantID, &c.Name, &c.Email, &c.Phone, &c.CreatedAt, &c.UpdatedAt,
	)
	return &c, err
}

func (r *Repository) UpdateCustomer(ctx context.Context, tenantID, id uuid.UUID, in CustomerInput) (*models.Customer, error) {
	var c models.Customer
	err := r.pool.QueryRow(ctx, `
		UPDATE customers SET name = $3, email = $4, phone = $5, updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2
		RETURNING id, tenant_id, name, email, phone, created_at, updated_at
	`, id, tenantID, in.Name, in.Email, in.Phone).Scan(
		&c.ID, &c.TenantID, &c.Name, &c.Email, &c.Phone, &c.CreatedAt, &c.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.ErrNotFound
	}
	return &c, err
}

func (r *Repository) DeleteCustomer(ctx context.Context, tenantID, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM customers WHERE id = $1 AND tenant_id = $2`, id, tenantID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return apperror.ErrNotFound
	}
	return nil
}
