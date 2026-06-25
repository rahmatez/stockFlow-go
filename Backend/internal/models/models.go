package models

import (
	"time"

	"github.com/google/uuid"
)

type Tenant struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type User struct {
	ID           uuid.UUID `json:"id"`
	TenantID     uuid.UUID `json:"tenant_id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	FullName     string    `json:"full_name"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Plan struct {
	ID                uuid.UUID `json:"id"`
	Name              string    `json:"name"`
	Slug              string    `json:"slug"`
	MaxProducts       int       `json:"max_products"`
	MaxOrdersPerMonth int       `json:"max_orders_per_month"`
	MaxUsers          int       `json:"max_users"`
	PriceCents        int       `json:"price_cents"`
	StripePriceID     *string   `json:"stripe_price_id,omitempty"`
}

type TenantSubscription struct {
	ID                   uuid.UUID  `json:"id"`
	TenantID             uuid.UUID  `json:"tenant_id"`
	PlanID               uuid.UUID  `json:"plan_id"`
	Status               string     `json:"status"`
	StripeCustomerID     *string    `json:"stripe_customer_id,omitempty"`
	StripeSubscriptionID *string    `json:"stripe_subscription_id,omitempty"`
	Plan                 *Plan      `json:"plan,omitempty"`
}

type Category struct {
	ID          uuid.UUID `json:"id"`
	TenantID    uuid.UUID `json:"tenant_id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Product struct {
	ID                uuid.UUID  `json:"id"`
	TenantID          uuid.UUID  `json:"tenant_id"`
	CategoryID        *uuid.UUID `json:"category_id,omitempty"`
	SKU               string     `json:"sku"`
	Name              string     `json:"name"`
	Description       *string    `json:"description,omitempty"`
	CostPrice         float64    `json:"cost_price"`
	SellPrice         float64    `json:"sell_price"`
	StockOnHand       int        `json:"stock_on_hand"`
	LowStockThreshold int        `json:"low_stock_threshold"`
	IsActive          bool       `json:"is_active"`
	CategoryName      *string    `json:"category_name,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type Warehouse struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenant_id"`
	Name      string    `json:"name"`
	IsDefault bool      `json:"is_default"`
	CreatedAt time.Time `json:"created_at"`
}

type InventoryMovement struct {
	ID            uuid.UUID  `json:"id"`
	TenantID      uuid.UUID  `json:"tenant_id"`
	WarehouseID   uuid.UUID  `json:"warehouse_id"`
	ProductID     uuid.UUID  `json:"product_id"`
	MovementType  string     `json:"movement_type"`
	Quantity      int        `json:"quantity"`
	ReferenceType *string    `json:"reference_type,omitempty"`
	ReferenceID   *uuid.UUID `json:"reference_id,omitempty"`
	Notes         *string    `json:"notes,omitempty"`
	ProductName   string     `json:"product_name,omitempty"`
	ProductSKU    string     `json:"product_sku,omitempty"`
	WarehouseName string     `json:"warehouse_name,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

type Customer struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenant_id"`
	Name      string    `json:"name"`
	Email     *string   `json:"email,omitempty"`
	Phone     *string   `json:"phone,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type OrderStatusHistory struct {
	ID            uuid.UUID  `json:"id"`
	FromStatus    *string    `json:"from_status,omitempty"`
	ToStatus      string     `json:"to_status"`
	ChangedByName *string    `json:"changed_by_name,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

type Order struct {
	ID            uuid.UUID            `json:"id"`
	TenantID      uuid.UUID            `json:"tenant_id"`
	CustomerID    *uuid.UUID           `json:"customer_id,omitempty"`
	OrderNumber   string               `json:"order_number"`
	Status        string               `json:"status"`
	Subtotal      float64              `json:"subtotal"`
	Total         float64              `json:"total"`
	Notes         *string              `json:"notes,omitempty"`
	CustomerName  *string              `json:"customer_name,omitempty"`
	Items         []OrderItem          `json:"items,omitempty"`
	StatusHistory []OrderStatusHistory `json:"status_history,omitempty"`
	CreatedAt     time.Time            `json:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at"`
}

type OrderItem struct {
	ID          uuid.UUID `json:"id"`
	OrderID     uuid.UUID `json:"order_id"`
	TenantID    uuid.UUID `json:"tenant_id"`
	ProductID   uuid.UUID `json:"product_id"`
	ProductName string    `json:"product_name"`
	ProductSKU  string    `json:"product_sku"`
	Quantity    int       `json:"quantity"`
	UnitPrice   float64   `json:"unit_price"`
	LineTotal   float64   `json:"line_total"`
}

type DashboardStats struct {
	OrdersToday    int64   `json:"orders_today"`
	RevenueToday   float64 `json:"revenue_today"`
	PendingOrders  int64   `json:"pending_orders"`
	LowStockCount  int64   `json:"low_stock_count"`
	TotalProducts  int64   `json:"total_products"`
	TotalCustomers int64   `json:"total_customers"`
}

type DashboardTrends struct {
	OrdersChange    float64 `json:"orders_change"`
	RevenueChange   float64 `json:"revenue_change"`
	CustomersChange float64 `json:"customers_change"`
}

type Notification struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenant_id"`
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}

type SalesDataPoint struct {
	Date    string  `json:"date"`
	Revenue float64 `json:"revenue"`
	Orders  int64   `json:"orders"`
}

type TopProduct struct {
	ProductID   uuid.UUID `json:"product_id"`
	ProductName string    `json:"product_name"`
	TotalSold   int64     `json:"total_sold"`
	Revenue     float64   `json:"revenue"`
}
