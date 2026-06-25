package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/oms-saas/oms-saas-go/internal/apperror"
	"github.com/oms-saas/oms-saas-go/internal/models"
)

func (r *Repository) SavePasswordResetToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM password_reset_tokens WHERE user_id = $1`, userID)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, `
		INSERT INTO password_reset_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)
	`, userID, tokenHash, expiresAt)
	return err
}

func (r *Repository) GetUserByResetToken(ctx context.Context, tokenHash string) (*models.User, error) {
	var u models.User
	err := r.pool.QueryRow(ctx, `
		SELECT u.id, u.tenant_id, u.email, u.password_hash, u.full_name, u.role, u.created_at, u.updated_at
		FROM password_reset_tokens prt
		JOIN users u ON u.id = prt.user_id
		WHERE prt.token_hash = $1 AND prt.expires_at > NOW()
	`, tokenHash).Scan(
		&u.ID, &u.TenantID, &u.Email, &u.PasswordHash, &u.FullName, &u.Role, &u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.ErrNotFound
	}
	return &u, err
}

func (r *Repository) DeletePasswordResetToken(ctx context.Context, tokenHash string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM password_reset_tokens WHERE token_hash = $1`, tokenHash)
	return err
}
