"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { PageHeader } from "@/components/layout/page-header";
import { StatCard } from "@/components/dashboard/stat-card";
import { EmptyState } from "@/components/dashboard/empty-state";
import { Skeleton } from "@/components/ui/skeleton";
import { formatCurrency, formatDate } from "@/lib/utils";
import {
  ShoppingCart, TrendUp, Clock, Warning, Tag, Users, ChartBar, Medal,
} from "@phosphor-icons/react";
import {
  BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer,
} from "recharts";
import { chartTheme, useTheme } from "@/components/theme-provider";

export default function DashboardPage() {
  const { theme } = useTheme();
  const colors = chartTheme[theme];

  const { data, isLoading } = useQuery({
    queryKey: ["dashboard"],
    queryFn: () => api.getDashboard(),
  });

  if (isLoading) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-10 w-48" />
        <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
          {Array.from({ length: 4 }).map((_, i) => (
            <Skeleton key={i} className="h-32" />
          ))}
        </div>
      </div>
    );
  }

  if (!data) return <p className="text-sm text-muted">Failed to load dashboard</p>;

  const {
    stats,
    trends,
    sales_chart = [],
    top_products = [],
    low_stock = [],
    recent_orders = [],
  } = data;

  const fmtTrend = (v?: number) => {
    if (v === undefined || v === null) return undefined;
    const sign = v >= 0 ? "+" : "";
    return { value: `${sign}${v.toFixed(1)}%`, positive: v >= 0 };
  };

  return (
    <div className="space-y-6">
      <PageHeader title="Dashboard" description="Ringkasan bisnis Anda hari ini." />

      <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <StatCard label="Total Customers" value={stats.total_customers} icon={Users} trend={fmtTrend(trends?.customers_change)} />
        <StatCard label="Orders Today" value={stats.orders_today} icon={ShoppingCart} trend={fmtTrend(trends?.orders_change)} />
        <StatCard label="Pending Orders" value={stats.pending_orders} icon={Clock} />
        <StatCard label="Revenue Today" value={formatCurrency(stats.revenue_today)} icon={TrendUp} trend={fmtTrend(trends?.revenue_change)} />
      </div>

      <div className="grid gap-6 xl:grid-cols-3">
        <Card className="xl:col-span-2">
          <CardHeader>
            <div className="flex items-center justify-between">
              <div>
                <CardTitle>Monthly Sales</CardTitle>
                <CardDescription>Revenue 7 hari terakhir</CardDescription>
              </div>
              <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary-light text-primary-brand">
                <ChartBar size={16} weight="fill" />
              </div>
            </div>
          </CardHeader>
          <CardContent>
            <ResponsiveContainer width="100%" height={280}>
              <BarChart data={sales_chart} barSize={32}>
                <CartesianGrid strokeDasharray="3 3" stroke={colors.grid} vertical={false} />
                <XAxis dataKey="date" tickFormatter={(v) => v.slice(5)} tick={{ fontSize: 12, fill: colors.tick }} axisLine={false} tickLine={false} />
                <YAxis tick={{ fontSize: 12, fill: colors.tick }} axisLine={false} tickLine={false} />
                <Tooltip
                  contentStyle={{ borderRadius: 8, border: `1px solid ${colors.tooltipBorder}`, background: colors.tooltipBg, boxShadow: "0 4px 12px rgba(0,0,0,0.12)" }}
                  formatter={(v) => [formatCurrency(Number(v ?? 0)), "Revenue"]}
                />
                <Bar dataKey="revenue" fill={colors.bar} radius={[4, 4, 0, 0]} />
              </BarChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Top Products</CardTitle>
            <CardDescription>Produk terlaris</CardDescription>
          </CardHeader>
          <CardContent>
            {top_products.length === 0 ? (
              <EmptyState icon={Tag} title="Belum ada penjualan" description="Data akan muncul setelah ada order." />
            ) : (
              <div className="space-y-3">
                {top_products.map((p, i) => (
                  <div key={p.product_id} className="flex items-center gap-3 rounded-lg bg-surface-muted px-3 py-2.5">
                    <span className="flex h-7 w-7 shrink-0 items-center justify-center rounded-lg bg-card text-xs font-bold text-primary-brand ring-1 ring-border">
                      {i + 1}
                    </span>
                    <div className="min-w-0 flex-1">
                      <p className="truncate text-sm font-medium text-foreground">{p.product_name}</p>
                      <p className="text-xs text-muted-light">{p.total_sold} terjual</p>
                    </div>
                    <p className="text-sm font-semibold text-foreground">{formatCurrency(p.revenue)}</p>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <div className="flex items-center gap-2">
              <Warning size={18} className="text-danger-text" weight="fill" />
              <CardTitle>Low Stock Alerts</CardTitle>
            </div>
          </CardHeader>
          <CardContent>
            {low_stock.length === 0 ? (
              <EmptyState icon={Tag} title="Stok aman" description="Semua produk di atas batas minimum." />
            ) : (
              <div className="space-y-2">
                {low_stock.map((p) => (
                  <div key={p.id} className="flex items-center justify-between rounded-lg border border-danger-border bg-danger-bg px-4 py-3">
                    <div>
                      <p className="text-sm font-medium text-foreground">{p.name}</p>
                      <p className="font-mono text-xs text-muted-light">{p.sku}</p>
                    </div>
                    <span className="rounded-md bg-danger-bg px-2 py-1 text-xs font-semibold text-danger-text ring-1 ring-danger-border">
                      {p.stock_on_hand} left
                    </span>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <div className="flex items-center gap-2">
              <Medal size={18} className="text-primary-brand" weight="fill" />
              <CardTitle>Recent Orders</CardTitle>
            </div>
          </CardHeader>
          <CardContent>
            {recent_orders.length === 0 ? (
              <EmptyState icon={ShoppingCart} title="Belum ada pesanan" description="Buat pesanan pertama Anda." />
            ) : (
              <div className="space-y-2">
                {recent_orders.map((o) => (
                  <div key={o.id} className="flex items-center justify-between rounded-lg bg-surface-muted px-4 py-3">
                    <div>
                      <p className="font-mono text-sm font-medium text-foreground">{o.order_number}</p>
                      <p className="text-xs text-muted-light">{formatDate(o.created_at)}</p>
                    </div>
                    <div className="text-right">
                      <Badge status={o.status} />
                      <p className="mt-1.5 text-sm font-semibold text-foreground">{formatCurrency(o.total)}</p>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
