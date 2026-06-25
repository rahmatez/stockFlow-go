package repository

import (
	"context"
	"strings"
	"testing"

	"github.com/oms-saas/oms-saas-go/internal/testutil"
)

func TestCreateOrderUsesPerTenantSequence(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	repo := New(pool)
	ctx := context.Background()

	suffix := strings.ReplaceAll(t.Name(), "/", "_")
	user, tenant, err := repo.RegisterTenant(ctx, "Order Test "+suffix, suffix+"@test.local", "password123", "Test Owner")
	if err != nil {
		t.Fatalf("register tenant: %v", err)
	}
	ctx = testutil.TenantCtx(ctx, tenant.ID)

	product, err := repo.CreateProduct(ctx, tenant.ID, ProductInput{
		SKU: "SKU-TEST-1", Name: "Test Product", SellPrice: 10000, CostPrice: 5000, IsActive: true,
	})
	if err != nil {
		t.Fatalf("create product: %v", err)
	}

	order1, err := repo.CreateOrder(ctx, tenant.ID, CreateOrderInput{
		Items:  []OrderItemInput{{ProductID: product.ID, Quantity: 1}},
		UserID: &user.ID,
	})
	if err != nil {
		t.Fatalf("create order 1: %v", err)
	}

	order2, err := repo.CreateOrder(ctx, tenant.ID, CreateOrderInput{
		Items:  []OrderItemInput{{ProductID: product.ID, Quantity: 1}},
		UserID: &user.ID,
	})
	if err != nil {
		t.Fatalf("create order 2: %v", err)
	}

	if !strings.HasPrefix(order1.OrderNumber, "ORD-") {
		t.Fatalf("unexpected order number format: %s", order1.OrderNumber)
	}
	if order1.OrderNumber == order2.OrderNumber {
		t.Fatalf("order numbers should differ: %s", order1.OrderNumber)
	}
}

func TestOrderStatusHistory(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	repo := New(pool)
	ctx := context.Background()

	suffix := strings.ReplaceAll(t.Name(), "/", "_")
	user, tenant, err := repo.RegisterTenant(ctx, "History Test "+suffix, suffix+"@hist.local", "password123", "Test Owner")
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	ctx = testutil.TenantCtx(ctx, tenant.ID)

	product, err := repo.CreateProduct(ctx, tenant.ID, ProductInput{
		SKU: "SKU-HIST-1", Name: "Hist Product", SellPrice: 5000, CostPrice: 3000, IsActive: true,
	})
	if err != nil {
		t.Fatalf("create product: %v", err)
	}

	wh, err := repo.GetDefaultWarehouse(ctx, tenant.ID)
	if err != nil {
		t.Fatalf("default warehouse: %v", err)
	}
	_, err = repo.AdjustInventory(ctx, tenant.ID, InventoryAdjustInput{
		ProductID: product.ID, MovementType: "IN", Quantity: 50,
		WarehouseID: wh.ID,
	})
	if err != nil {
		t.Fatalf("adjust inventory: %v", err)
	}

	order, err := repo.CreateOrder(ctx, tenant.ID, CreateOrderInput{
		Items:  []OrderItemInput{{ProductID: product.ID, Quantity: 1}},
		UserID: &user.ID,
	})
	if err != nil {
		t.Fatalf("create order: %v", err)
	}

	_, err = repo.UpdateOrderStatus(ctx, tenant.ID, order.ID, "confirmed", &user.ID)
	if err != nil {
		t.Fatalf("confirm order: %v", err)
	}

	got, err := repo.GetOrder(ctx, tenant.ID, order.ID)
	if err != nil {
		t.Fatalf("get order: %v", err)
	}
	if len(got.StatusHistory) < 2 {
		t.Fatalf("expected at least 2 history entries, got %d", len(got.StatusHistory))
	}
	if got.StatusHistory[0].ToStatus != "draft" {
		t.Fatalf("first history should be draft, got %s", got.StatusHistory[0].ToStatus)
	}
	if got.StatusHistory[len(got.StatusHistory)-1].ToStatus != "confirmed" {
		t.Fatalf("last history should be confirmed, got %s", got.StatusHistory[len(got.StatusHistory)-1].ToStatus)
	}
}
