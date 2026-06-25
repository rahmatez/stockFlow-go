package testutil

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/oms-saas/oms-saas-go/internal/auth"
	"github.com/oms-saas/oms-saas-go/internal/database"
)

func SetupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		url = "postgres://oms_app:oms@localhost:5433/oms?sslmode=disable"
	}
	migrateURL := os.Getenv("MIGRATION_DATABASE_URL")
	if migrateURL == "" {
		migrateURL = "postgres://oms:oms@localhost:5433/oms?sslmode=disable"
	}

	ctx := context.Background()

	migrationsDir := filepath.Join("db", "migrations")
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		migrationsDir = filepath.Join("..", "..", "db", "migrations")
	}

	migratePool, err := database.NewPool(ctx, migrateURL)
	if err != nil {
		t.Fatalf("connect migration database: %v", err)
	}
	if err := database.RunMigrations(ctx, migratePool, migrationsDir); err != nil {
		migratePool.Close()
		t.Fatalf("run migrations: %v", err)
	}
	migratePool.Close()

	pool, err := database.NewPool(ctx, url)
	if err != nil {
		t.Fatalf("connect database: %v", err)
	}

	t.Cleanup(func() { pool.Close() })
	return pool
}

func TenantCtx(ctx context.Context, tenantID uuid.UUID) context.Context {
	return auth.WithTenantID(ctx, tenantID)
}
