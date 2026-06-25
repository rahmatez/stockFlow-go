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

func (r *Repository) getWarehouseStock(ctx context.Context, tx pgx.Tx, tenantID, warehouseID, productID uuid.UUID) (int, error) {
	var qty int
	err := tx.QueryRow(ctx, `
		SELECT COALESCE(sl.quantity, 0)
		FROM products p
		LEFT JOIN stock_levels sl ON sl.warehouse_id = $2 AND sl.product_id = p.id AND sl.tenant_id = $1
		WHERE p.id = $3 AND p.tenant_id = $1
	`, tenantID, warehouseID, productID).Scan(&qty)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, apperror.ErrNotFound
	}
	return qty, err
}

func (r *Repository) upsertWarehouseStock(ctx context.Context, tx pgx.Tx, tenantID, warehouseID, productID uuid.UUID, quantity int) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO stock_levels (tenant_id, warehouse_id, product_id, quantity, updated_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (warehouse_id, product_id) DO UPDATE SET quantity = $4, updated_at = NOW()
	`, tenantID, warehouseID, productID, quantity)
	return err
}

func (r *Repository) syncProductStockFromWarehouses(ctx context.Context, tx pgx.Tx, tenantID, productID uuid.UUID) error {
	var total int
	err := tx.QueryRow(ctx, `
		SELECT COALESCE(SUM(quantity), 0) FROM stock_levels
		WHERE tenant_id = $1 AND product_id = $2
	`, tenantID, productID).Scan(&total)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
		UPDATE products SET stock_on_hand = $3, updated_at = NOW()
		WHERE id = $2 AND tenant_id = $1
	`, tenantID, productID, total)
	return err
}

func (r *Repository) ListWarehouses(ctx context.Context, tenantID uuid.UUID) ([]models.Warehouse, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, name, is_default, created_at FROM warehouses WHERE tenant_id = $1 ORDER BY name
	`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.Warehouse, 0)
	for rows.Next() {
		var w models.Warehouse
		if err := rows.Scan(&w.ID, &w.TenantID, &w.Name, &w.IsDefault, &w.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, w)
	}
	return items, rows.Err()
}

func (r *Repository) GetWarehouse(ctx context.Context, tenantID, id uuid.UUID) (*models.Warehouse, error) {
	var w models.Warehouse
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, name, is_default, created_at FROM warehouses
		WHERE id = $1 AND tenant_id = $2
	`, id, tenantID).Scan(&w.ID, &w.TenantID, &w.Name, &w.IsDefault, &w.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.ErrNotFound
	}
	return &w, err
}

func (r *Repository) GetDefaultWarehouse(ctx context.Context, tenantID uuid.UUID) (*models.Warehouse, error) {
	var w models.Warehouse
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, name, is_default, created_at FROM warehouses
		WHERE tenant_id = $1 AND is_default = true LIMIT 1
	`, tenantID).Scan(&w.ID, &w.TenantID, &w.Name, &w.IsDefault, &w.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.ErrNotFound
	}
	return &w, err
}

type WarehouseInput struct {
	Name string
}

func (r *Repository) CreateWarehouse(ctx context.Context, tenantID uuid.UUID, in WarehouseInput) (*models.Warehouse, error) {
	var w models.Warehouse
	err := r.pool.QueryRow(ctx, `
		INSERT INTO warehouses (tenant_id, name, is_default) VALUES ($1, $2, false)
		RETURNING id, tenant_id, name, is_default, created_at
	`, tenantID, in.Name).Scan(&w.ID, &w.TenantID, &w.Name, &w.IsDefault, &w.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *Repository) UpdateWarehouse(ctx context.Context, tenantID, id uuid.UUID, in WarehouseInput) (*models.Warehouse, error) {
	var w models.Warehouse
	err := r.pool.QueryRow(ctx, `
		UPDATE warehouses SET name = $3, updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2
		RETURNING id, tenant_id, name, is_default, created_at
	`, id, tenantID, in.Name).Scan(&w.ID, &w.TenantID, &w.Name, &w.IsDefault, &w.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.ErrNotFound
	}
	return &w, err
}

func (r *Repository) DeleteWarehouse(ctx context.Context, tenantID, id uuid.UUID) error {
	var isDefault bool
	err := r.pool.QueryRow(ctx, `
		SELECT is_default FROM warehouses WHERE id = $1 AND tenant_id = $2
	`, id, tenantID).Scan(&isDefault)
	if errors.Is(err, pgx.ErrNoRows) {
		return apperror.ErrNotFound
	}
	if err != nil {
		return err
	}
	if isDefault {
		return fmt.Errorf("%w: cannot delete default warehouse", apperror.ErrValidation)
	}

	var stockCount int64
	err = r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM stock_levels WHERE warehouse_id = $1 AND tenant_id = $2 AND quantity > 0
	`, id, tenantID).Scan(&stockCount)
	if err != nil {
		return err
	}
	if stockCount > 0 {
		return fmt.Errorf("%w: warehouse has stock", apperror.ErrValidation)
	}

	_, err = r.pool.Exec(ctx, `DELETE FROM warehouses WHERE id = $1 AND tenant_id = $2`, id, tenantID)
	return err
}

func (r *Repository) ListInventoryMovements(ctx context.Context, tenantID uuid.UUID, p ListParams) ([]models.InventoryMovement, int64, error) {
	p = p.Normalize()

	var total int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM inventory_movements WHERE tenant_id = $1`, tenantID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT im.id, im.tenant_id, im.warehouse_id, im.product_id, im.movement_type, im.quantity,
		       im.reference_type, im.reference_id, im.notes, p.name, p.sku, w.name, im.created_at
		FROM inventory_movements im
		JOIN products p ON p.id = im.product_id
		JOIN warehouses w ON w.id = im.warehouse_id
		WHERE im.tenant_id = $1
		ORDER BY im.created_at DESC
		LIMIT $2 OFFSET $3
	`, tenantID, p.Limit, p.Offset())
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]models.InventoryMovement, 0)
	for rows.Next() {
		var m models.InventoryMovement
		if err := rows.Scan(
			&m.ID, &m.TenantID, &m.WarehouseID, &m.ProductID, &m.MovementType, &m.Quantity,
			&m.ReferenceType, &m.ReferenceID, &m.Notes, &m.ProductName, &m.ProductSKU, &m.WarehouseName, &m.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		items = append(items, m)
	}
	return items, total, rows.Err()
}

type InventoryAdjustInput struct {
	WarehouseID  uuid.UUID
	ProductID    uuid.UUID
	MovementType string
	Quantity     int
	Notes        *string
	UserID       *uuid.UUID
}

func (r *Repository) AdjustInventory(ctx context.Context, tenantID uuid.UUID, in InventoryAdjustInput) (*models.InventoryMovement, error) {
	if in.Quantity <= 0 {
		return nil, fmt.Errorf("%w: quantity must be positive", apperror.ErrValidation)
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var dummy uuid.UUID
	if err := tx.QueryRow(ctx, `
		SELECT id FROM products WHERE id = $1 AND tenant_id = $2 FOR UPDATE
	`, in.ProductID, tenantID).Scan(&dummy); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.ErrNotFound
		}
		return nil, err
	}

	current, err := r.getWarehouseStock(ctx, tx, tenantID, in.WarehouseID, in.ProductID)
	if err != nil {
		return nil, err
	}

	delta := in.Quantity
	switch in.MovementType {
	case "IN", "ADJUSTMENT":
		// positive delta
	case "OUT":
		delta = -in.Quantity
	default:
		return nil, fmt.Errorf("%w: invalid movement type", apperror.ErrValidation)
	}

	newQty := current + delta
	if newQty < 0 {
		return nil, apperror.ErrInsufficientStock
	}

	if err := r.upsertWarehouseStock(ctx, tx, tenantID, in.WarehouseID, in.ProductID, newQty); err != nil {
		return nil, err
	}
	if err := r.syncProductStockFromWarehouses(ctx, tx, tenantID, in.ProductID); err != nil {
		return nil, err
	}

	var movement models.InventoryMovement
	err = tx.QueryRow(ctx, `
		INSERT INTO inventory_movements (tenant_id, warehouse_id, product_id, movement_type, quantity, notes, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, tenant_id, warehouse_id, product_id, movement_type, quantity, reference_type, reference_id, notes, created_at
	`, tenantID, in.WarehouseID, in.ProductID, in.MovementType, in.Quantity, in.Notes, in.UserID).Scan(
		&movement.ID, &movement.TenantID, &movement.WarehouseID, &movement.ProductID,
		&movement.MovementType, &movement.Quantity, &movement.ReferenceType, &movement.ReferenceID,
		&movement.Notes, &movement.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &movement, nil
}

type InventoryTransferInput struct {
	FromWarehouseID uuid.UUID
	ToWarehouseID   uuid.UUID
	ProductID       uuid.UUID
	Quantity        int
	Notes           *string
	UserID          *uuid.UUID
}

func (r *Repository) TransferInventory(ctx context.Context, tenantID uuid.UUID, in InventoryTransferInput) error {
	if in.Quantity <= 0 {
		return fmt.Errorf("%w: quantity must be positive", apperror.ErrValidation)
	}
	if in.FromWarehouseID == in.ToWarehouseID {
		return fmt.Errorf("%w: warehouses must differ", apperror.ErrValidation)
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var dummy uuid.UUID
	if err := tx.QueryRow(ctx, `
		SELECT id FROM products WHERE id = $1 AND tenant_id = $2 FOR UPDATE
	`, in.ProductID, tenantID).Scan(&dummy); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperror.ErrNotFound
		}
		return err
	}

	fromQty, err := r.getWarehouseStock(ctx, tx, tenantID, in.FromWarehouseID, in.ProductID)
	if err != nil {
		return err
	}
	if fromQty < in.Quantity {
		return apperror.ErrInsufficientStock
	}

	toQty, err := r.getWarehouseStock(ctx, tx, tenantID, in.ToWarehouseID, in.ProductID)
	if err != nil {
		return err
	}

	if err := r.upsertWarehouseStock(ctx, tx, tenantID, in.FromWarehouseID, in.ProductID, fromQty-in.Quantity); err != nil {
		return err
	}
	if err := r.upsertWarehouseStock(ctx, tx, tenantID, in.ToWarehouseID, in.ProductID, toQty+in.Quantity); err != nil {
		return err
	}
	if err := r.syncProductStockFromWarehouses(ctx, tx, tenantID, in.ProductID); err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO inventory_movements (tenant_id, warehouse_id, product_id, movement_type, quantity, notes, created_by, reference_type)
		VALUES ($1, $2, $3, 'TRANSFER', $4, $5, $6, 'transfer_out')
	`, tenantID, in.FromWarehouseID, in.ProductID, in.Quantity, in.Notes, in.UserID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO inventory_movements (tenant_id, warehouse_id, product_id, movement_type, quantity, notes, created_by, reference_type)
		VALUES ($1, $2, $3, 'TRANSFER', $4, $5, $6, 'transfer_in')
	`, tenantID, in.ToWarehouseID, in.ProductID, in.Quantity, in.Notes, in.UserID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *Repository) DeductStockForOrder(ctx context.Context, tx pgx.Tx, tenantID, productID uuid.UUID, qty int, warehouseID uuid.UUID, orderID uuid.UUID, userID *uuid.UUID) error {
	current, err := r.getWarehouseStock(ctx, tx, tenantID, warehouseID, productID)
	if err != nil {
		return err
	}
	if current < qty {
		return apperror.ErrInsufficientStock
	}

	if err := r.upsertWarehouseStock(ctx, tx, tenantID, warehouseID, productID, current-qty); err != nil {
		return err
	}
	if err := r.syncProductStockFromWarehouses(ctx, tx, tenantID, productID); err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO inventory_movements (tenant_id, warehouse_id, product_id, movement_type, quantity, reference_type, reference_id, created_by)
		VALUES ($1, $2, $3, 'OUT', $4, 'order', $5, $6)
	`, tenantID, warehouseID, productID, qty, orderID, userID)
	return err
}

func (r *Repository) RestoreStockForOrder(ctx context.Context, tx pgx.Tx, tenantID, productID uuid.UUID, qty int, warehouseID, orderID uuid.UUID, userID *uuid.UUID) error {
	current, err := r.getWarehouseStock(ctx, tx, tenantID, warehouseID, productID)
	if err != nil {
		return err
	}

	if err := r.upsertWarehouseStock(ctx, tx, tenantID, warehouseID, productID, current+qty); err != nil {
		return err
	}
	if err := r.syncProductStockFromWarehouses(ctx, tx, tenantID, productID); err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO inventory_movements (tenant_id, warehouse_id, product_id, movement_type, quantity, reference_type, reference_id, created_by)
		VALUES ($1, $2, $3, 'IN', $4, 'order_cancel', $5, $6)
	`, tenantID, warehouseID, productID, qty, orderID, userID)
	return err
}
