package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/oms-saas/oms-saas-go/internal/apperror"
	"github.com/oms-saas/oms-saas-go/internal/models"
	"golang.org/x/crypto/bcrypt"
)

func (r *Repository) GetUserByEmailAndTenant(ctx context.Context, email, tenantSlug string) (*models.User, error) {
	var u models.User
	err := r.pool.QueryRow(ctx, `
		SELECT u.id, u.tenant_id, u.email, u.password_hash, u.full_name, u.role, u.created_at, u.updated_at
		FROM users u
		JOIN tenants t ON t.id = u.tenant_id
		WHERE u.email = $1 AND t.slug = $2
	`, strings.ToLower(email), tenantSlug).Scan(
		&u.ID, &u.TenantID, &u.Email, &u.PasswordHash, &u.FullName, &u.Role, &u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.ErrNotFound
	}
	return &u, err
}

func (r *Repository) CountUsersByEmail(ctx context.Context, email string) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE email = $1`, strings.ToLower(email)).Scan(&count)
	return count, err
}

func (r *Repository) ListUsers(ctx context.Context, tenantID uuid.UUID) ([]models.User, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, email, password_hash, full_name, role, created_at, updated_at
		FROM users WHERE tenant_id = $1 ORDER BY created_at
	`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]models.User, 0)
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.TenantID, &u.Email, &u.PasswordHash, &u.FullName, &u.Role, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

type CreateUserInput struct {
	Email    string
	Password string
	FullName string
	Role     string
}

func (r *Repository) CreateUser(ctx context.Context, tenantID uuid.UUID, in CreateUserInput) (*models.User, error) {
	if err := r.CheckPlanLimit(ctx, tenantID, "user"); err != nil {
		return nil, err
	}

	role := in.Role
	if role == "" {
		role = "staff"
	}
	if role != "admin" && role != "staff" {
		return nil, fmt.Errorf("%w: invalid role", apperror.ErrValidation)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	var u models.User
	err = r.pool.QueryRow(ctx, `
		INSERT INTO users (tenant_id, email, password_hash, full_name, role)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, tenant_id, email, password_hash, full_name, role, created_at, updated_at
	`, tenantID, strings.ToLower(in.Email), string(hash), in.FullName, role).Scan(
		&u.ID, &u.TenantID, &u.Email, &u.PasswordHash, &u.FullName, &u.Role, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

type UpdateUserInput struct {
	FullName *string
	Role     *string
}

func (r *Repository) UpdateUser(ctx context.Context, tenantID, userID uuid.UUID, in UpdateUserInput) (*models.User, error) {
	u, err := r.GetUserByID(ctx, tenantID, userID)
	if err != nil {
		return nil, err
	}

	fullName := u.FullName
	if in.FullName != nil {
		fullName = *in.FullName
	}
	role := u.Role
	if in.Role != nil {
		if *in.Role != "admin" && *in.Role != "staff" && *in.Role != "owner" {
			return nil, fmt.Errorf("%w: invalid role", apperror.ErrValidation)
		}
		if u.Role == "owner" && *in.Role != "owner" {
			return nil, fmt.Errorf("%w: cannot change owner role", apperror.ErrValidation)
		}
		role = *in.Role
	}

	err = r.pool.QueryRow(ctx, `
		UPDATE users SET full_name = $3, role = $4, updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2
		RETURNING id, tenant_id, email, password_hash, full_name, role, created_at, updated_at
	`, userID, tenantID, fullName, role).Scan(
		&u.ID, &u.TenantID, &u.Email, &u.PasswordHash, &u.FullName, &u.Role, &u.CreatedAt, &u.UpdatedAt,
	)
	return u, err
}

func (r *Repository) DeleteUser(ctx context.Context, tenantID, userID uuid.UUID) error {
	var role string
	err := r.pool.QueryRow(ctx, `SELECT role FROM users WHERE id = $1 AND tenant_id = $2`, userID, tenantID).Scan(&role)
	if errors.Is(err, pgx.ErrNoRows) {
		return apperror.ErrNotFound
	}
	if err != nil {
		return err
	}
	if role == "owner" {
		return fmt.Errorf("%w: cannot delete owner", apperror.ErrValidation)
	}
	_, err = r.pool.Exec(ctx, `DELETE FROM users WHERE id = $1 AND tenant_id = $2`, userID, tenantID)
	return err
}

func (r *Repository) SetUserPassword(ctx context.Context, tenantID, userID uuid.UUID, newPassword string) error {
	if len(newPassword) < 8 {
		return fmt.Errorf("%w: password must be at least 8 characters", apperror.ErrValidation)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, `UPDATE users SET password_hash = $3, updated_at = NOW() WHERE id = $1 AND tenant_id = $2`, userID, tenantID, string(hash))
	return err
}

func (r *Repository) UpdateUserPassword(ctx context.Context, tenantID, userID uuid.UUID, oldPassword, newPassword string) error {
	u, err := r.GetUserByID(ctx, tenantID, userID)
	if err != nil {
		return err
	}
	if err := VerifyPassword(u.PasswordHash, oldPassword); err != nil {
		return apperror.ErrUnauthorized
	}
	return r.SetUserPassword(ctx, tenantID, userID, newPassword)
}

func (r *Repository) UpdateUserProfile(ctx context.Context, tenantID, userID uuid.UUID, fullName string) (*models.User, error) {
	var u models.User
	err := r.pool.QueryRow(ctx, `
		UPDATE users SET full_name = $3, updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2
		RETURNING id, tenant_id, email, password_hash, full_name, role, created_at, updated_at
	`, userID, tenantID, fullName).Scan(
		&u.ID, &u.TenantID, &u.Email, &u.PasswordHash, &u.FullName, &u.Role, &u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.ErrNotFound
	}
	return &u, err
}
