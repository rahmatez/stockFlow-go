package repository

import (
	"context"
	"strings"
	"testing"

	"github.com/oms-saas/oms-saas-go/internal/auth"
	"github.com/oms-saas/oms-saas-go/internal/testutil"
)

func TestRLSTenantIsolation(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	repo := New(pool)
	ctx := context.Background()

	suffix := strings.ReplaceAll(t.Name(), "/", "_")

	_, tenantA, err := repo.RegisterTenant(ctx, "RLS A "+suffix, suffix+"-a@rls.local", "password123", "Owner A")
	if err != nil {
		t.Fatalf("register tenant A: %v", err)
	}
	_, tenantB, err := repo.RegisterTenant(ctx, "RLS B "+suffix, suffix+"-b@rls.local", "password123", "Owner B")
	if err != nil {
		t.Fatalf("register tenant B: %v", err)
	}

	ctxA := auth.WithTenantID(ctx, tenantA.ID)
	ctxB := auth.WithTenantID(ctx, tenantB.ID)

	productA, err := repo.CreateProduct(ctxA, tenantA.ID, ProductInput{
		SKU: "RLS-A-1", Name: "Product A", SellPrice: 1000, CostPrice: 500, IsActive: true,
	})
	if err != nil {
		t.Fatalf("create product A: %v", err)
	}
	productB, err := repo.CreateProduct(ctxB, tenantB.ID, ProductInput{
		SKU: "RLS-B-1", Name: "Product B", SellPrice: 2000, CostPrice: 1000, IsActive: true,
	})
	if err != nil {
		t.Fatalf("create product B: %v", err)
	}

	gotA, err := repo.GetProduct(ctxA, tenantA.ID, productA.ID)
	if err != nil {
		t.Fatalf("get product A: %v", err)
	}
	if gotA.SKU != "RLS-A-1" {
		t.Fatalf("expected product A, got %s", gotA.SKU)
	}

	_, err = repo.GetProduct(ctxA, tenantA.ID, productB.ID)
	if err == nil {
		t.Fatal("tenant A should not see tenant B product")
	}

	productsA, _, err := repo.ListProducts(ctxA, tenantA.ID, ListParams{Page: 1, Limit: 50})
	if err != nil {
		t.Fatalf("list products A: %v", err)
	}
	for _, p := range productsA {
		if p.ID == productB.ID {
			t.Fatal("tenant A list should not include tenant B product")
		}
	}

	withoutRLS, total, err := repo.ListProducts(ctx, tenantA.ID, ListParams{Page: 1, Limit: 50})
	if err != nil {
		t.Fatalf("list without tenant context: %v", err)
	}
	if total > 0 || len(withoutRLS) > 0 {
		t.Fatal("without tenant context in ctx, RLS should return empty")
	}
	_ = tenantB
}
