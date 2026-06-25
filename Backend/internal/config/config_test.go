package config

import "testing"

func TestLoadDefaultDatabaseURL(t *testing.T) {
	cfg := Load()
	if cfg.DatabaseURL != "postgres://oms_app:oms@localhost:5433/oms?sslmode=disable" {
		t.Fatalf("expected default app database URL, got %s", cfg.DatabaseURL)
	}
	if cfg.MigrationDatabaseURL != "postgres://oms:oms@localhost:5433/oms?sslmode=disable" {
		t.Fatalf("expected default migration database URL, got %s", cfg.MigrationDatabaseURL)
	}
}
