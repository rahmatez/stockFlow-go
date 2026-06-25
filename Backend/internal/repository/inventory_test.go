package repository

import (
	"context"
	"strings"
	"testing"

	"github.com/oms-saas/oms-saas-go/internal/testutil"
)

func TestInventoryAdjustTransferAndOrderStock(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	repo := New(pool)
	ctx := context.Background()

	suffix := strings.ReplaceAll(t.Name(), "/", "_")
	user, tenant, err := repo.RegisterTenant(ctx, "Inv Test "+suffix, suffix+"@inv.local", "password123", "Test Owner")
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	ctx = testutil.TenantCtx(ctx, tenant.ID)

	wh2, err := repo.CreateWarehouse(ctx, tenant.ID, WarehouseInput{Name: "Secondary WH"})
	if err != nil {
		t.Fatalf("create warehouse: %v", err)
	}

	defaultWH, err := repo.GetDefaultWarehouse(ctx, tenant.ID)
	if err != nil {
		t.Fatalf("default warehouse: %v", err)
	}

	product, err := repo.CreateProduct(ctx, tenant.ID, ProductInput{
		SKU: "INV-001", Name: "Inv Product", SellPrice: 10000, CostPrice: 5000, IsActive: true,
	})
	if err != nil {
		t.Fatalf("create product: %v", err)
	}

	_, err = repo.AdjustInventory(ctx, tenant.ID, InventoryAdjustInput{
		WarehouseID: defaultWH.ID, ProductID: product.ID, MovementType: "IN", Quantity: 100,
	})
	if err != nil {
		t.Fatalf("adjust in: %v", err)
	}

	err = repo.TransferInventory(ctx, tenant.ID, InventoryTransferInput{
		FromWarehouseID: defaultWH.ID, ToWarehouseID: wh2.ID,
		ProductID: product.ID, Quantity: 30,
	})
	if err != nil {
		t.Fatalf("transfer: %v", err)
	}

	updated, err := repo.GetProduct(ctx, tenant.ID, product.ID)
	if err != nil {
		t.Fatalf("get product: %v", err)
	}
	if updated.StockOnHand != 100 {
		t.Fatalf("expected total stock 100, got %d", updated.StockOnHand)
	}

	order, err := repo.CreateOrder(ctx, tenant.ID, CreateOrderInput{
		Items:  []OrderItemInput{{ProductID: product.ID, Quantity: 10}},
		UserID: &user.ID,
	})
	if err != nil {
		t.Fatalf("create order: %v", err)
	}

	_, err = repo.UpdateOrderStatus(ctx, tenant.ID, order.ID, "confirmed", &user.ID)
	if err != nil {
		t.Fatalf("confirm order: %v", err)
	}

	updated, err = repo.GetProduct(ctx, tenant.ID, product.ID)
	if err != nil {
		t.Fatalf("get product after order: %v", err)
	}
	if updated.StockOnHand != 90 {
		t.Fatalf("expected stock 90 after order, got %d", updated.StockOnHand)
	}

	_, err = repo.UpdateOrderStatus(ctx, tenant.ID, order.ID, "cancelled", &user.ID)
	if err != nil {
		t.Fatalf("cancel order: %v", err)
	}

	updated, err = repo.GetProduct(ctx, tenant.ID, product.ID)
	if err != nil {
		t.Fatalf("get product after cancel: %v", err)
	}
	if updated.StockOnHand != 100 {
		t.Fatalf("expected stock 100 after cancel, got %d", updated.StockOnHand)
	}
}
