package repository

import (
	"context"
	"errors"
	"strconv"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/oms-saas/oms-saas-go/internal/apperror"
	"github.com/oms-saas/oms-saas-go/internal/models"
)

func (r *Repository) CreateNotification(ctx context.Context, tenantID uuid.UUID, notifType, title, message string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO notifications (tenant_id, type, title, message) VALUES ($1, $2, $3, $4)
	`, tenantID, notifType, title, message)
	return err
}

func (r *Repository) ListNotifications(ctx context.Context, tenantID uuid.UUID, limit int) ([]models.Notification, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	var unread int64
	_ = r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM notifications WHERE tenant_id = $1 AND is_read = false`, tenantID).Scan(&unread)

	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, type, title, message, is_read, created_at
		FROM notifications WHERE tenant_id = $1 ORDER BY created_at DESC LIMIT $2
	`, tenantID, limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]models.Notification, 0)
	for rows.Next() {
		var n models.Notification
		if err := rows.Scan(&n.ID, &n.TenantID, &n.Type, &n.Title, &n.Message, &n.IsRead, &n.CreatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, n)
	}
	return items, unread, rows.Err()
}

func (r *Repository) MarkNotificationRead(ctx context.Context, tenantID, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE notifications SET is_read = true WHERE id = $1 AND tenant_id = $2
	`, id, tenantID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return apperror.ErrNotFound
	}
	return nil
}

func (r *Repository) CheckAndCreateLowStockNotifications(ctx context.Context, tenantID uuid.UUID) error {
	rows, err := r.pool.Query(ctx, `
		SELECT name, stock_on_hand FROM products
		WHERE tenant_id = $1 AND is_active = true AND stock_on_hand <= low_stock_threshold
		LIMIT 5
	`, tenantID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var stock int
		if err := rows.Scan(&name, &stock); err != nil {
			return err
		}
		var exists bool
		_ = r.pool.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM notifications
				WHERE tenant_id = $1 AND type = 'low_stock' AND message LIKE '%' || $2 || '%'
				AND created_at > NOW() - interval '24 hours'
			)
		`, tenantID, name).Scan(&exists)
		if !exists {
			_ = r.CreateNotification(ctx, tenantID, "low_stock", "Low Stock Alert",
				name+" has only "+strconv.Itoa(stock)+" units left")
		}
	}
	return nil
}

func (r *Repository) GetNotification(ctx context.Context, tenantID, id uuid.UUID) (*models.Notification, error) {
	var n models.Notification
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, type, title, message, is_read, created_at
		FROM notifications WHERE id = $1 AND tenant_id = $2
	`, id, tenantID).Scan(&n.ID, &n.TenantID, &n.Type, &n.Title, &n.Message, &n.IsRead, &n.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.ErrNotFound
	}
	return &n, err
}

func (r *Repository) CheckAndCreateLimitWarningNotifications(ctx context.Context, tenantID uuid.UUID) error {
	sub, err := r.GetTenantSubscription(ctx, tenantID)
	if err != nil || sub.Plan == nil {
		return err
	}

	productCount, _ := r.CountProducts(ctx, tenantID)
	orderCount, _ := r.CountOrdersThisMonth(ctx, tenantID)
	userCount, _ := r.CountUsers(ctx, tenantID)

	checks := []struct {
		label   string
		current int64
		max     int
	}{
		{"products", productCount, sub.Plan.MaxProducts},
		{"orders this month", orderCount, sub.Plan.MaxOrdersPerMonth},
		{"users", userCount, sub.Plan.MaxUsers},
	}

	for _, c := range checks {
		if c.max <= 0 {
			continue
		}
		pct := float64(c.current) / float64(c.max) * 100
		if pct < 80 {
			continue
		}
		var exists bool
		_ = r.pool.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM notifications
				WHERE tenant_id = $1 AND type = 'limit_warning' AND message LIKE '%' || $2 || '%'
				AND created_at > NOW() - interval '24 hours'
			)
		`, tenantID, c.label).Scan(&exists)
		if exists {
			continue
		}
		msg := c.label + " usage at " + strconv.Itoa(int(pct)) + "% (" +
			strconv.FormatInt(c.current, 10) + "/" + strconv.Itoa(c.max) + ")"
		_ = r.CreateNotification(ctx, tenantID, "limit_warning", "Plan Limit Warning", msg)
	}
	return nil
}
