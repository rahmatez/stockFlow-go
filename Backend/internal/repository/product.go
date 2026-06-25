package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/oms-saas/oms-saas-go/internal/apperror"
	"github.com/oms-saas/oms-saas-go/internal/models"
)

type ListParams struct {
	Page   int
	Limit  int
	Search string
}

func (p ListParams) Offset() int {
	if p.Page < 1 {
		p.Page = 1
	}
	return (p.Page - 1) * p.Limit
}

func (p ListParams) Normalize() ListParams {
	if p.Limit <= 0 {
		p.Limit = 20
	}
	if p.Page <= 0 {
		p.Page = 1
	}
	return p
}

func (r *Repository) ListCategories(ctx context.Context, tenantID uuid.UUID) ([]models.Category, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, name, description, created_at, updated_at
		FROM categories WHERE tenant_id = $1 ORDER BY name
	`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.Category, 0)
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.TenantID, &c.Name, &c.Description, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, c)
	}
	return items, rows.Err()
}

func (r *Repository) CreateCategory(ctx context.Context, tenantID uuid.UUID, name string, description *string) (*models.Category, error) {
	var c models.Category
	err := r.pool.QueryRow(ctx, `
		INSERT INTO categories (tenant_id, name, description) VALUES ($1, $2, $3)
		RETURNING id, tenant_id, name, description, created_at, updated_at
	`, tenantID, name, description).Scan(&c.ID, &c.TenantID, &c.Name, &c.Description, &c.CreatedAt, &c.UpdatedAt)
	return &c, err
}

func (r *Repository) UpdateCategory(ctx context.Context, tenantID, id uuid.UUID, name string, description *string) (*models.Category, error) {
	var c models.Category
	err := r.pool.QueryRow(ctx, `
		UPDATE categories SET name = $3, description = $4, updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2
		RETURNING id, tenant_id, name, description, created_at, updated_at
	`, id, tenantID, name, description).Scan(&c.ID, &c.TenantID, &c.Name, &c.Description, &c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.ErrNotFound
	}
	return &c, err
}

func (r *Repository) DeleteCategory(ctx context.Context, tenantID, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM categories WHERE id = $1 AND tenant_id = $2`, id, tenantID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return apperror.ErrNotFound
	}
	return nil
}

func (r *Repository) ListProducts(ctx context.Context, tenantID uuid.UUID, p ListParams) ([]models.Product, int64, error) {
	p = p.Normalize()
	search := "%" + p.Search + "%"

	var total int64
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM products
		WHERE tenant_id = $1 AND ($2 = '' OR name ILIKE $3 OR sku ILIKE $3)
	`, tenantID, p.Search, search).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT p.id, p.tenant_id, p.category_id, p.sku, p.name, p.description,
		       p.cost_price, p.sell_price, p.stock_on_hand, p.low_stock_threshold, p.is_active,
		       c.name, p.created_at, p.updated_at
		FROM products p
		LEFT JOIN categories c ON c.id = p.category_id
		WHERE p.tenant_id = $1 AND ($2 = '' OR p.name ILIKE $3 OR p.sku ILIKE $3)
		ORDER BY p.created_at DESC
		LIMIT $4 OFFSET $5
	`, tenantID, p.Search, search, p.Limit, p.Offset())
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]models.Product, 0)
	for rows.Next() {
		var pr models.Product
		if err := rows.Scan(
			&pr.ID, &pr.TenantID, &pr.CategoryID, &pr.SKU, &pr.Name, &pr.Description,
			&pr.CostPrice, &pr.SellPrice, &pr.StockOnHand, &pr.LowStockThreshold, &pr.IsActive,
			&pr.CategoryName, &pr.CreatedAt, &pr.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		items = append(items, pr)
	}
	return items, total, rows.Err()
}

func (r *Repository) GetProduct(ctx context.Context, tenantID, id uuid.UUID) (*models.Product, error) {
	var pr models.Product
	err := r.pool.QueryRow(ctx, `
		SELECT p.id, p.tenant_id, p.category_id, p.sku, p.name, p.description,
		       p.cost_price, p.sell_price, p.stock_on_hand, p.low_stock_threshold, p.is_active,
		       c.name, p.created_at, p.updated_at
		FROM products p
		LEFT JOIN categories c ON c.id = p.category_id
		WHERE p.id = $1 AND p.tenant_id = $2
	`, id, tenantID).Scan(
		&pr.ID, &pr.TenantID, &pr.CategoryID, &pr.SKU, &pr.Name, &pr.Description,
		&pr.CostPrice, &pr.SellPrice, &pr.StockOnHand, &pr.LowStockThreshold, &pr.IsActive,
		&pr.CategoryName, &pr.CreatedAt, &pr.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.ErrNotFound
	}
	return &pr, err
}

type ProductInput struct {
	CategoryID        *uuid.UUID
	SKU               string
	Name              string
	Description       *string
	CostPrice         float64
	SellPrice         float64
	LowStockThreshold int
	IsActive          bool
}

func (r *Repository) CreateProduct(ctx context.Context, tenantID uuid.UUID, in ProductInput) (*models.Product, error) {
	if err := r.CheckPlanLimit(ctx, tenantID, "product"); err != nil {
		return nil, err
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var pr models.Product
	err = tx.QueryRow(ctx, `
		INSERT INTO products (tenant_id, category_id, sku, name, description, cost_price, sell_price, low_stock_threshold, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, tenant_id, category_id, sku, name, description, cost_price, sell_price, stock_on_hand, low_stock_threshold, is_active, created_at, updated_at
	`, tenantID, in.CategoryID, in.SKU, in.Name, in.Description, in.CostPrice, in.SellPrice, in.LowStockThreshold, in.IsActive).Scan(
		&pr.ID, &pr.TenantID, &pr.CategoryID, &pr.SKU, &pr.Name, &pr.Description,
		&pr.CostPrice, &pr.SellPrice, &pr.StockOnHand, &pr.LowStockThreshold, &pr.IsActive, &pr.CreatedAt, &pr.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create product: %w", err)
	}

	var whID uuid.UUID
	err = tx.QueryRow(ctx, `SELECT id FROM warehouses WHERE tenant_id = $1 AND is_default = true LIMIT 1`, tenantID).Scan(&whID)
	if err == nil {
		_, _ = tx.Exec(ctx, `
			INSERT INTO stock_levels (tenant_id, warehouse_id, product_id, quantity) VALUES ($1, $2, $3, 0)
			ON CONFLICT (warehouse_id, product_id) DO NOTHING
		`, tenantID, whID, pr.ID)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &pr, nil
}

func (r *Repository) UpdateProduct(ctx context.Context, tenantID, id uuid.UUID, in ProductInput) (*models.Product, error) {
	var pr models.Product
	err := r.pool.QueryRow(ctx, `
		UPDATE products SET category_id = $3, sku = $4, name = $5, description = $6,
		    cost_price = $7, sell_price = $8, low_stock_threshold = $9, is_active = $10, updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2
		RETURNING id, tenant_id, category_id, sku, name, description, cost_price, sell_price, stock_on_hand, low_stock_threshold, is_active, created_at, updated_at
	`, id, tenantID, in.CategoryID, in.SKU, in.Name, in.Description, in.CostPrice, in.SellPrice, in.LowStockThreshold, in.IsActive).Scan(
		&pr.ID, &pr.TenantID, &pr.CategoryID, &pr.SKU, &pr.Name, &pr.Description,
		&pr.CostPrice, &pr.SellPrice, &pr.StockOnHand, &pr.LowStockThreshold, &pr.IsActive, &pr.CreatedAt, &pr.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.ErrNotFound
	}
	return &pr, err
}

func (r *Repository) DeleteProduct(ctx context.Context, tenantID, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM products WHERE id = $1 AND tenant_id = $2`, id, tenantID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return apperror.ErrNotFound
	}
	return nil
}
