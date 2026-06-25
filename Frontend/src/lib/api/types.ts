export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  meta?: { page: number; limit: number; total: number };
  error?: { code: string; message: string };
}

export interface User {
  id: string;
  email: string;
  full_name: string;
  role: string;
}

export interface Tenant {
  id: string;
  name: string;
  slug: string;
  status: string;
}

export interface Tokens {
  access_token: string;
  refresh_token: string;
  expires_in: number;
}

export interface AuthResponse {
  user: User;
  tenant: Tenant;
  tokens: Tokens;
}

export interface Product {
  id: string;
  sku: string;
  name: string;
  description?: string;
  cost_price: number;
  sell_price: number;
  stock_on_hand: number;
  low_stock_threshold: number;
  is_active: boolean;
  category_name?: string;
  category_id?: string;
}

export interface Category {
  id: string;
  name: string;
  description?: string;
}

export interface Customer {
  id: string;
  name: string;
  email?: string;
  phone?: string;
}

export interface OrderStatusHistory {
  id: string;
  from_status?: string;
  to_status: string;
  changed_by_name?: string;
  created_at: string;
}

export interface Order {
  id: string;
  order_number: string;
  status: string;
  subtotal: number;
  total: number;
  customer_name?: string;
  customer_id?: string;
  notes?: string;
  items?: OrderItem[];
  status_history?: OrderStatusHistory[];
  created_at: string;
}

export interface OrderItem {
  id: string;
  product_id: string;
  product_name: string;
  product_sku: string;
  quantity: number;
  unit_price: number;
  line_total: number;
}

export interface InventoryMovement {
  id: string;
  movement_type: string;
  quantity: number;
  product_name: string;
  product_sku: string;
  warehouse_name: string;
  notes?: string;
  created_at: string;
}

export interface DashboardData {
  stats: {
    orders_today: number;
    revenue_today: number;
    pending_orders: number;
    low_stock_count: number;
    total_products: number;
    total_customers: number;
  };
  trends?: {
    orders_change: number;
    revenue_change: number;
    customers_change: number;
  };
  sales_chart: { date: string; revenue: number; orders: number }[];
  top_products: { product_id: string; product_name: string; total_sold: number; revenue: number }[];
  low_stock: Product[];
  recent_orders: Order[];
}

export interface Plan {
  id: string;
  name: string;
  slug: string;
  max_products: number;
  max_orders_per_month: number;
  max_users: number;
  price_cents: number;
}

export interface Warehouse {
  id: string;
  name: string;
  is_default: boolean;
  created_at: string;
}

export interface SalesDataPoint {
  date: string;
  revenue: number;
  orders: number;
}

export interface BillingData {
  subscription: {
    status: string;
    plan?: Plan;
  };
  plans: Plan[];
  usage: {
    products: number;
    orders_month: number;
    users: number;
  };
}
