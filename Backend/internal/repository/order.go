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

var validTransitions = map[string][]string{
	"draft":      {"confirmed", "cancelled"},
	"confirmed":  {"processing", "cancelled"},
	"processing": {"shipped", "cancelled"},
	"shipped":    {"delivered"},
	"delivered":  {},
	"cancelled":  {},
}

func canTransition(from, to string) bool {
	for _, s := range validTransitions[from] {
		if s == to {
			return true
		}
	}
	return false
}

type orderItemQty struct {
	productID uuid.UUID
	qty       int
}

func (r *Repository) orderItemsInTx(ctx context.Context, tx pgx.Tx, orderID, tenantID uuid.UUID) ([]orderItemQty, error) {
	rows, err := tx.Query(ctx, `SELECT product_id, quantity FROM order_items WHERE order_id = $1 AND tenant_id = $2`, orderID, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]orderItemQty, 0)
	for rows.Next() {
		var item orderItemQty
		if err := rows.Scan(&item.productID, &item.qty); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

type OrderItemInput struct {
	ProductID uuid.UUID
	Quantity  int
}

type CreateOrderInput struct {
	CustomerID *uuid.UUID
	Notes      *string
	Items      []OrderItemInput
	UserID     *uuid.UUID
}

func (r *Repository) ListOrders(ctx context.Context, tenantID uuid.UUID, p ListParams) ([]models.Order, int64, error) {
	p = p.Normalize()
	search := "%" + p.Search + "%"

	var total int64
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM orders o
		LEFT JOIN customers c ON c.id = o.customer_id
		WHERE o.tenant_id = $1 AND ($2 = '' OR o.order_number ILIKE $3 OR COALESCE(c.name,'') ILIKE $3)
	`, tenantID, p.Search, search).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT o.id, o.tenant_id, o.customer_id, o.order_number, o.status, o.subtotal, o.total, o.notes,
		       c.name, o.created_at, o.updated_at
		FROM orders o
		LEFT JOIN customers c ON c.id = o.customer_id
		WHERE o.tenant_id = $1 AND ($2 = '' OR o.order_number ILIKE $3 OR COALESCE(c.name,'') ILIKE $3)
		ORDER BY o.created_at DESC
		LIMIT $4 OFFSET $5
	`, tenantID, p.Search, search, p.Limit, p.Offset())
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]models.Order, 0)
	for rows.Next() {
		var o models.Order
		if err := rows.Scan(
			&o.ID, &o.TenantID, &o.CustomerID, &o.OrderNumber, &o.Status, &o.Subtotal, &o.Total, &o.Notes,
			&o.CustomerName, &o.CreatedAt, &o.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		items = append(items, o)
	}
	return items, total, rows.Err()
}

func (r *Repository) GetOrder(ctx context.Context, tenantID, id uuid.UUID) (*models.Order, error) {
	var o models.Order
	err := r.pool.QueryRow(ctx, `
		SELECT o.id, o.tenant_id, o.customer_id, o.order_number, o.status, o.subtotal, o.total, o.notes,
		       c.name, o.created_at, o.updated_at
		FROM orders o
		LEFT JOIN customers c ON c.id = o.customer_id
		WHERE o.id = $1 AND o.tenant_id = $2
	`, id, tenantID).Scan(
		&o.ID, &o.TenantID, &o.CustomerID, &o.OrderNumber, &o.Status, &o.Subtotal, &o.Total, &o.Notes,
		&o.CustomerName, &o.CreatedAt, &o.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, order_id, tenant_id, product_id, product_name, product_sku, quantity, unit_price, line_total
		FROM order_items WHERE order_id = $1 AND tenant_id = $2
	`, id, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item models.OrderItem
		if err := rows.Scan(&item.ID, &item.OrderID, &item.TenantID, &item.ProductID, &item.ProductName, &item.ProductSKU, &item.Quantity, &item.UnitPrice, &item.LineTotal); err != nil {
			return nil, err
		}
		o.Items = append(o.Items, item)
	}

	history, err := r.ListOrderStatusHistory(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	o.StatusHistory = history
	return &o, nil
}

func (r *Repository) ListOrderStatusHistory(ctx context.Context, tenantID, orderID uuid.UUID) ([]models.OrderStatusHistory, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT h.id, h.from_status, h.to_status, u.full_name, h.created_at
		FROM order_status_history h
		LEFT JOIN users u ON u.id = h.changed_by
		WHERE h.order_id = $1 AND h.tenant_id = $2
		ORDER BY h.created_at ASC
	`, orderID, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.OrderStatusHistory, 0)
	for rows.Next() {
		var h models.OrderStatusHistory
		if err := rows.Scan(&h.ID, &h.FromStatus, &h.ToStatus, &h.ChangedByName, &h.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, h)
	}
	return items, rows.Err()
}

func (r *Repository) CreateOrder(ctx context.Context, tenantID uuid.UUID, in CreateOrderInput) (*models.Order, error) {
	if len(in.Items) == 0 {
		return nil, fmt.Errorf("%w: order must have items", apperror.ErrValidation)
	}
	if err := r.CheckPlanLimit(ctx, tenantID, "order"); err != nil {
		return nil, err
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var tenantSlug string
	if err := tx.QueryRow(ctx, `SELECT slug FROM tenants WHERE id = $1`, tenantID).Scan(&tenantSlug); err != nil {
		return nil, err
	}

	var seq int64
	if err := tx.QueryRow(ctx, `
		UPDATE tenants SET order_seq = order_seq + 1 WHERE id = $1 RETURNING order_seq
	`, tenantID).Scan(&seq); err != nil {
		return nil, err
	}
	orderNumber := fmt.Sprintf("ORD-%s-%d", tenantSlug[:min(8, len(tenantSlug))], seq)

	var subtotal float64
	type lineData struct {
		product models.Product
		qty     int
	}
	var lines []lineData

	for _, item := range in.Items {
		var p models.Product
		err := tx.QueryRow(ctx, `
			SELECT id, tenant_id, sku, name, sell_price, stock_on_hand, is_active
			FROM products WHERE id = $1 AND tenant_id = $2
		`, item.ProductID, tenantID).Scan(&p.ID, &p.TenantID, &p.SKU, &p.Name, &p.SellPrice, &p.StockOnHand, &p.IsActive)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: product not found", apperror.ErrNotFound)
		}
		if err != nil {
			return nil, err
		}
		if !p.IsActive {
			return nil, fmt.Errorf("%w: product %s is inactive", apperror.ErrValidation, p.SKU)
		}
		if item.Quantity <= 0 {
			return nil, fmt.Errorf("%w: invalid quantity", apperror.ErrValidation)
		}
		subtotal += p.SellPrice * float64(item.Quantity)
		lines = append(lines, lineData{product: p, qty: item.Quantity})
	}

	var order models.Order
	err = tx.QueryRow(ctx, `
		INSERT INTO orders (tenant_id, customer_id, order_number, status, subtotal, total, notes, created_by)
		VALUES ($1, $2, $3, 'draft', $4, $4, $5, $6)
		RETURNING id, tenant_id, customer_id, order_number, status, subtotal, total, notes, created_at, updated_at
	`, tenantID, in.CustomerID, orderNumber, subtotal, in.Notes, in.UserID).Scan(
		&order.ID, &order.TenantID, &order.CustomerID, &order.OrderNumber, &order.Status,
		&order.Subtotal, &order.Total, &order.Notes, &order.CreatedAt, &order.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	for _, line := range lines {
		lineTotal := line.product.SellPrice * float64(line.qty)
		var item models.OrderItem
		err = tx.QueryRow(ctx, `
			INSERT INTO order_items (order_id, tenant_id, product_id, product_name, product_sku, quantity, unit_price, line_total)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			RETURNING id, order_id, tenant_id, product_id, product_name, product_sku, quantity, unit_price, line_total
		`, order.ID, tenantID, line.product.ID, line.product.Name, line.product.SKU, line.qty, line.product.SellPrice, lineTotal).Scan(
			&item.ID, &item.OrderID, &item.TenantID, &item.ProductID, &item.ProductName, &item.ProductSKU,
			&item.Quantity, &item.UnitPrice, &item.LineTotal,
		)
		if err != nil {
			return nil, err
		}
		order.Items = append(order.Items, item)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO order_status_history (order_id, tenant_id, from_status, to_status, changed_by)
		VALUES ($1, $2, NULL, 'draft', $3)
	`, order.ID, tenantID, in.UserID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *Repository) UpdateOrderStatus(ctx context.Context, tenantID, orderID uuid.UUID, newStatus string, userID *uuid.UUID) (*models.Order, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var currentStatus string
	err = tx.QueryRow(ctx, `
		SELECT status FROM orders WHERE id = $1 AND tenant_id = $2 FOR UPDATE
	`, orderID, tenantID).Scan(&currentStatus)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	if !canTransition(currentStatus, newStatus) {
		return nil, apperror.ErrInvalidStatus
	}

	var whID uuid.UUID
	_ = tx.QueryRow(ctx, `SELECT id FROM warehouses WHERE tenant_id = $1 AND is_default = true LIMIT 1`, tenantID).Scan(&whID)

	if newStatus == "confirmed" && currentStatus == "draft" {
		items, err := r.orderItemsInTx(ctx, tx, orderID, tenantID)
		if err != nil {
			return nil, err
		}
		for _, item := range items {
			if err := r.DeductStockForOrder(ctx, tx, tenantID, item.productID, item.qty, whID, orderID, userID); err != nil {
				return nil, err
			}
		}
	}

	if newStatus == "cancelled" && (currentStatus == "confirmed" || currentStatus == "processing") {
		items, err := r.orderItemsInTx(ctx, tx, orderID, tenantID)
		if err != nil {
			return nil, err
		}
		for _, item := range items {
			if err := r.RestoreStockForOrder(ctx, tx, tenantID, item.productID, item.qty, whID, orderID, userID); err != nil {
				return nil, err
			}
		}
	}

	_, err = tx.Exec(ctx, `UPDATE orders SET status = $3, updated_at = NOW() WHERE id = $1 AND tenant_id = $2`, orderID, tenantID, newStatus)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO order_status_history (order_id, tenant_id, from_status, to_status, changed_by)
		VALUES ($1, $2, $3, $4, $5)
	`, orderID, tenantID, currentStatus, newStatus, userID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.GetOrder(ctx, tenantID, orderID)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
