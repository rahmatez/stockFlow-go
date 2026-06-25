import {
  getAccessToken,
  getRefreshToken,
  setTokens,
  clearTokens,
} from "@/lib/auth";
import type {
  ApiResponse,
  AuthResponse,
  BillingData,
  Category,
  Customer,
  DashboardData,
  InventoryMovement,
  Order,
  Product,
  Tokens,
  Warehouse,
  SalesDataPoint,
} from "./types";

const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1";

class ApiError extends Error {
  code: string;
  status: number;

  constructor(message: string, code: string, status: number) {
    super(message);
    this.code = code;
    this.status = status;
  }
}

async function refreshAccessToken(): Promise<string | null> {
  const refresh = getRefreshToken();
  if (!refresh) return null;

  const res = await fetch(`${API_URL}/auth/refresh`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ refresh_token: refresh }),
  });

  if (!res.ok) {
    clearTokens();
    return null;
  }

  const json: ApiResponse<Tokens> = await res.json();
  if (json.data) {
    setTokens(json.data.access_token, json.data.refresh_token);
    return json.data.access_token;
  }
  return null;
}

async function authFetch(path: string, options: RequestInit = {}, retry = true): Promise<Response> {
  const token = getAccessToken();
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(options.headers as Record<string, string>),
  };
  if (token) headers.Authorization = `Bearer ${token}`;

  const res = await fetch(`${API_URL}${path}`, { ...options, headers });

  if (res.status === 401 && retry) {
    const newToken = await refreshAccessToken();
    if (newToken) return authFetch(path, options, false);
    clearTokens();
    if (typeof window !== "undefined") window.location.href = "/login";
    throw new ApiError("Unauthorized", "UNAUTHORIZED", 401);
  }

  return res;
}

async function request<T>(path: string, options: RequestInit = {}, retry = true): Promise<T> {
  const res = await authFetch(path, options, retry);
  const json: ApiResponse<T> = await res.json();
  if (!json.success) {
    throw new ApiError(
      json.error?.message || "Request failed",
      json.error?.code || "ERROR",
      res.status
    );
  }
  return json.data as T;
}

async function requestWithMeta<T>(
  path: string,
  options: RequestInit = {}
): Promise<{ data: T; meta: { page: number; limit: number; total: number } }> {
  const res = await authFetch(path, options);
  const json = await res.json();
  if (!json.success) {
    throw new ApiError(json.error?.message, json.error?.code, res.status);
  }
  return {
    data: Array.isArray(json.data) ? json.data : json.data ?? [],
    meta: json.meta ?? { page: 1, limit: 20, total: 0 },
  };
}

export function getErrorMessage(err: unknown): string {
  if (err instanceof ApiError) {
    const messages: Record<string, string> = {
      INSUFFICIENT_STOCK: "Stok tidak mencukupi untuk operasi ini.",
      LIMIT_EXCEEDED: "Batas paket tercapai. Silakan upgrade plan Anda.",
      INVALID_STATUS: "Transisi status tidak valid.",
      VALIDATION_ERROR: err.message,
      CONFLICT: err.message,
      NOT_FOUND: "Data tidak ditemukan.",
    };
    return messages[err.code] || err.message;
  }
  if (err instanceof Error) return err.message;
  return "Terjadi kesalahan. Coba lagi.";
}

export const api = {
  register: (body: {
    tenant_name: string;
    email: string;
    password: string;
    full_name: string;
  }) => request<AuthResponse>("/auth/register", { method: "POST", body: JSON.stringify(body) }),

  login: (body: { email: string; password: string; tenant_slug?: string }) =>
    request<AuthResponse>("/auth/login", { method: "POST", body: JSON.stringify(body) }),

  logout: () =>
    request<{ message: string }>("/auth/logout", {
      method: "POST",
      body: JSON.stringify({ refresh_token: getRefreshToken() }),
    }),

  getTenant: () => request<{ tenant: unknown; subscription: unknown }>("/tenant/me"),

  updateTenant: (body: { name: string }) =>
    request<unknown>("/tenant/me", { method: "PATCH", body: JSON.stringify(body) }),

  getDashboard: () => request<DashboardData>("/reports/dashboard"),

  getSalesReport: (days = 30) => request<SalesDataPoint[]>(`/reports/sales?days=${days}`),

  getInventoryReport: () => request<Product[]>("/reports/inventory"),

  listProducts: (params?: { page?: number; search?: string }) => {
    const q = new URLSearchParams();
    if (params?.page) q.set("page", String(params.page));
    if (params?.search) q.set("search", params.search);
    return requestWithMeta<Product[]>(`/products?${q}`);
  },

  createProduct: (body: Partial<Product>) =>
    request<Product>("/products", { method: "POST", body: JSON.stringify(body) }),

  updateProduct: (id: string, body: Partial<Product>) =>
    request<Product>(`/products/${id}`, { method: "PUT", body: JSON.stringify(body) }),

  deleteProduct: (id: string) =>
    request<{ message: string }>(`/products/${id}`, { method: "DELETE" }),

  listCategories: () => request<Category[]>("/categories"),

  createCategory: (body: { name: string; description?: string }) =>
    request<Category>("/categories", { method: "POST", body: JSON.stringify(body) }),

  updateCategory: (id: string, body: { name: string; description?: string }) =>
    request<Category>(`/categories/${id}`, { method: "PUT", body: JSON.stringify(body) }),

  deleteCategory: (id: string) =>
    request<{ message: string }>(`/categories/${id}`, { method: "DELETE" }),

  listCustomers: (params?: { page?: number; search?: string }) => {
    const q = new URLSearchParams();
    if (params?.page) q.set("page", String(params.page));
    if (params?.search) q.set("search", params.search);
    return requestWithMeta<Customer[]>(`/customers?${q}`);
  },

  createCustomer: (body: { name: string; email?: string; phone?: string }) =>
    request<Customer>("/customers", { method: "POST", body: JSON.stringify(body) }),

  updateCustomer: (id: string, body: { name: string; email?: string; phone?: string }) =>
    request<Customer>(`/customers/${id}`, { method: "PUT", body: JSON.stringify(body) }),

  deleteCustomer: (id: string) =>
    request<{ message: string }>(`/customers/${id}`, { method: "DELETE" }),

  listInventory: (page = 1) =>
    requestWithMeta<InventoryMovement[]>(`/inventory?page=${page}`),

  listWarehouses: () => request<Warehouse[]>("/inventory/warehouses"),

  adjustInventory: (body: {
    product_id: string;
    movement_type: string;
    quantity: number;
    warehouse_id?: string;
    notes?: string;
  }) =>
    request<InventoryMovement>("/inventory/adjust", {
      method: "POST",
      body: JSON.stringify(body),
    }),

  transferInventory: (body: {
    from_warehouse_id: string;
    to_warehouse_id: string;
    product_id: string;
    quantity: number;
    notes?: string;
  }) =>
    request<{ message: string }>("/inventory/transfer", {
      method: "POST",
      body: JSON.stringify(body),
    }),

  listOrders: (params?: { page?: number; search?: string }) => {
    const q = new URLSearchParams();
    if (params?.page) q.set("page", String(params.page));
    if (params?.search) q.set("search", params.search);
    return requestWithMeta<Order[]>(`/orders?${q}`);
  },

  getOrder: (id: string) => request<Order>(`/orders/${id}`),

  createOrder: (body: {
    customer_id?: string;
    notes?: string;
    items: { product_id: string; quantity: number }[];
  }) => request<Order>("/orders", { method: "POST", body: JSON.stringify(body) }),

  updateOrderStatus: (id: string, status: string) =>
    request<Order>(`/orders/${id}/status`, {
      method: "PATCH",
      body: JSON.stringify({ status }),
    }),

  getUserMe: () => request<{ id: string; email: string; full_name: string; role: string }>("/users/me"),

  updateUserMe: (body: { full_name?: string; old_password?: string; new_password?: string }) =>
    request<unknown>("/users/me", { method: "PATCH", body: JSON.stringify(body) }),

  listUsers: () => request<{ id: string; email: string; full_name: string; role: string; created_at: string }[]>("/users"),

  createUser: (body: { email: string; password: string; full_name: string; role?: string }) =>
    request<{ id: string; email: string; full_name: string; role: string; email_sent?: boolean }>("/users", { method: "POST", body: JSON.stringify(body) }),

  updateUser: (id: string, body: { full_name?: string; role?: string }) =>
    request<unknown>(`/users/${id}`, { method: "PATCH", body: JSON.stringify(body) }),

  deleteUser: (id: string) =>
    request<{ message: string }>(`/users/${id}`, { method: "DELETE" }),

  createWarehouse: (body: { name: string }) =>
    request<Warehouse>("/warehouses", { method: "POST", body: JSON.stringify(body) }),

  updateWarehouse: (id: string, body: { name: string }) =>
    request<Warehouse>(`/warehouses/${id}`, { method: "PUT", body: JSON.stringify(body) }),

  deleteWarehouse: (id: string) =>
    request<{ message: string }>(`/warehouses/${id}`, { method: "DELETE" }),

  listNotifications: () =>
    request<{ notifications: { id: string; title: string; message: string; is_read: boolean; created_at: string }[]; unread_count: number }>("/notifications"),

  markNotificationRead: (id: string) =>
    request<{ message: string }>(`/notifications/${id}/read`, { method: "PATCH" }),

  forgotPassword: (email: string) =>
    request<{ message: string }>("/auth/forgot-password", {
      method: "POST",
      body: JSON.stringify({ email }),
    }),

  resetPassword: (token: string, new_password: string) =>
    request<{ message: string }>("/auth/reset-password", {
      method: "POST",
      body: JSON.stringify({ token, new_password }),
    }),

  exportProducts: () => downloadCSV("/products/export", "products.csv"),
  exportSales: (days = 30) => downloadCSV(`/reports/sales/export?days=${days}`, `sales-${days}days.csv`),
  exportInventory: () => downloadCSV("/reports/inventory/export", "inventory-low-stock.csv"),

  getBilling: () => request<BillingData>("/billing/plan"),

  createBillingPortal: () =>
    request<{ portal_url: string }>("/billing/portal", {
      method: "POST",
      body: JSON.stringify({ return_url: `${window.location.origin}/settings/billing` }),
    }),

  createCheckout: (planSlug: string) =>
    request<{ checkout_url: string }>("/billing/checkout", {
      method: "POST",
      body: JSON.stringify({
        plan_slug: planSlug,
        success_url: `${window.location.origin}/settings/billing?success=1`,
        cancel_url: `${window.location.origin}/settings/billing?cancel=1`,
      }),
    }),
};

async function downloadCSV(path: string, filename: string) {
  const token = getAccessToken();
  const res = await fetch(`${API_URL}${path}`, {
    headers: token ? { Authorization: `Bearer ${token}` } : {},
  });
  if (!res.ok) throw new ApiError("Export failed", "EXPORT_ERROR", res.status);
  const blob = await res.blob();
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = filename;
  a.click();
  URL.revokeObjectURL(url);
}

export { ApiError };
