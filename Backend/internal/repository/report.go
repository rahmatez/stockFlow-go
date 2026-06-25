package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/oms-saas/oms-saas-go/internal/models"
)

func (r *Repository) GetDashboardStats(ctx context.Context, tenantID uuid.UUID) (*models.DashboardStats, error) {
	var s models.DashboardStats
	err := r.pool.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM orders WHERE tenant_id = $1 AND created_at >= CURRENT_DATE),
			(SELECT COALESCE(SUM(total), 0) FROM orders WHERE tenant_id = $1 AND created_at >= CURRENT_DATE AND status NOT IN ('draft','cancelled')),
			(SELECT COUNT(*) FROM orders WHERE tenant_id = $1 AND status IN ('confirmed','processing')),
			(SELECT COUNT(*) FROM products WHERE tenant_id = $1 AND stock_on_hand <= low_stock_threshold AND is_active = true),
			(SELECT COUNT(*) FROM products WHERE tenant_id = $1),
			(SELECT COUNT(*) FROM customers WHERE tenant_id = $1)
	`, tenantID).Scan(&s.OrdersToday, &s.RevenueToday, &s.PendingOrders, &s.LowStockCount, &s.TotalProducts, &s.TotalCustomers)
	return &s, err
}

func (r *Repository) GetSalesReport(ctx context.Context, tenantID uuid.UUID, days int) ([]models.SalesDataPoint, error) {
	if days <= 0 {
		days = 7
	}
	rows, err := r.pool.Query(ctx, `
		SELECT to_char(d.day, 'YYYY-MM-DD') AS date,
		       COALESCE(SUM(o.total), 0) AS revenue,
		       COUNT(o.id) AS orders
		FROM generate_series(CURRENT_DATE - ($2::int - 1), CURRENT_DATE, '1 day'::interval) AS d(day)
		LEFT JOIN orders o ON o.tenant_id = $1
			AND o.created_at >= d.day
			AND o.created_at < d.day + interval '1 day'
			AND o.status NOT IN ('draft','cancelled')
		GROUP BY d.day
		ORDER BY d.day
	`, tenantID, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	points := make([]models.SalesDataPoint, 0)
	for rows.Next() {
		var p models.SalesDataPoint
		if err := rows.Scan(&p.Date, &p.Revenue, &p.Orders); err != nil {
			return nil, err
		}
		points = append(points, p)
	}
	return points, rows.Err()
}

func (r *Repository) GetTopProducts(ctx context.Context, tenantID uuid.UUID, limit int) ([]models.TopProduct, error) {
	if limit <= 0 {
		limit = 5
	}
	rows, err := r.pool.Query(ctx, `
		SELECT oi.product_id, oi.product_name, SUM(oi.quantity) AS total_sold, SUM(oi.line_total) AS revenue
		FROM order_items oi
		JOIN orders o ON o.id = oi.order_id
		WHERE oi.tenant_id = $1 AND o.status NOT IN ('draft','cancelled')
		GROUP BY oi.product_id, oi.product_name
		ORDER BY total_sold DESC
		LIMIT $2
	`, tenantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.TopProduct, 0)
	for rows.Next() {
		var t models.TopProduct
		if err := rows.Scan(&t.ProductID, &t.ProductName, &t.TotalSold, &t.Revenue); err != nil {
			return nil, err
		}
		items = append(items, t)
	}
	return items, rows.Err()
}

func (r *Repository) GetLowStockProducts(ctx context.Context, tenantID uuid.UUID, limit int) ([]models.Product, error) {
	if limit <= 0 {
		limit = 10
	}
	rows, err := r.pool.Query(ctx, `
		SELECT p.id, p.tenant_id, p.category_id, p.sku, p.name, p.description,
		       p.cost_price, p.sell_price, p.stock_on_hand, p.low_stock_threshold, p.is_active,
		       NULL, p.created_at, p.updated_at
		FROM products p
		WHERE p.tenant_id = $1 AND p.is_active = true AND p.stock_on_hand <= p.low_stock_threshold
		ORDER BY p.stock_on_hand ASC
		LIMIT $2
	`, tenantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.Product, 0)
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(
			&p.ID, &p.TenantID, &p.CategoryID, &p.SKU, &p.Name, &p.Description,
			&p.CostPrice, &p.SellPrice, &p.StockOnHand, &p.LowStockThreshold, &p.IsActive,
			&p.CategoryName, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, p)
	}
	return items, rows.Err()
}

func pctChange(today, yesterday float64) float64 {
	if yesterday == 0 {
		if today > 0 {
			return 100
		}
		return 0
	}
	return ((today - yesterday) / yesterday) * 100
}

func (r *Repository) GetDashboardTrends(ctx context.Context, tenantID uuid.UUID) (*models.DashboardTrends, error) {
	var ordersToday, ordersYesterday int64
	var revenueToday, revenueYesterday float64
	var customersToday, customersYesterday int64

	err := r.pool.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM orders WHERE tenant_id = $1 AND created_at >= CURRENT_DATE),
			(SELECT COUNT(*) FROM orders WHERE tenant_id = $1 AND created_at >= CURRENT_DATE - 1 AND created_at < CURRENT_DATE),
			(SELECT COALESCE(SUM(total),0) FROM orders WHERE tenant_id = $1 AND created_at >= CURRENT_DATE AND status NOT IN ('draft','cancelled')),
			(SELECT COALESCE(SUM(total),0) FROM orders WHERE tenant_id = $1 AND created_at >= CURRENT_DATE - 1 AND created_at < CURRENT_DATE AND status NOT IN ('draft','cancelled')),
			(SELECT COUNT(*) FROM customers WHERE tenant_id = $1 AND created_at >= CURRENT_DATE),
			(SELECT COUNT(*) FROM customers WHERE tenant_id = $1 AND created_at >= CURRENT_DATE - 1 AND created_at < CURRENT_DATE)
	`, tenantID).Scan(&ordersToday, &ordersYesterday, &revenueToday, &revenueYesterday, &customersToday, &customersYesterday)
	if err != nil {
		return nil, err
	}

	return &models.DashboardTrends{
		OrdersChange:    pctChange(float64(ordersToday), float64(ordersYesterday)),
		RevenueChange:   pctChange(revenueToday, revenueYesterday),
		CustomersChange: pctChange(float64(customersToday), float64(customersYesterday)),
	}, nil
}

func (r *Repository) GetRecentOrders(ctx context.Context, tenantID uuid.UUID, limit int) ([]models.Order, error) {
	if limit <= 0 {
		limit = 5
	}
	rows, err := r.pool.Query(ctx, `
		SELECT o.id, o.tenant_id, o.customer_id, o.order_number, o.status, o.subtotal, o.total, o.notes,
		       c.name, o.created_at, o.updated_at
		FROM orders o
		LEFT JOIN customers c ON c.id = o.customer_id
		WHERE o.tenant_id = $1
		ORDER BY o.created_at DESC
		LIMIT $2
	`, tenantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.Order, 0)
	for rows.Next() {
		var o models.Order
		if err := rows.Scan(
			&o.ID, &o.TenantID, &o.CustomerID, &o.OrderNumber, &o.Status, &o.Subtotal, &o.Total, &o.Notes,
			&o.CustomerName, &o.CreatedAt, &o.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, o)
	}
	return items, rows.Err()
}
