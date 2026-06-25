"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { PageHeader } from "@/components/layout/page-header";
import { EmptyState } from "@/components/dashboard/empty-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Button } from "@/components/ui/button";
import { formatCurrency } from "@/lib/utils";
import { ChartLineUp, Medal, Warning } from "@phosphor-icons/react";
import { XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Area, AreaChart } from "recharts";
import { chartTheme, useTheme } from "@/components/theme-provider";

export default function ReportsPage() {
  const { theme } = useTheme();
  const colors = chartTheme[theme];
  const [days, setDays] = useState(30);

  const { data: sales, isLoading: salesLoading } = useQuery({
    queryKey: ["sales-report", days],
    queryFn: () => api.getSalesReport(days),
  });

  const { data: lowStock, isLoading: invLoading } = useQuery({
    queryKey: ["inventory-report"],
    queryFn: () => api.getInventoryReport(),
  });

  const { data: dashboard } = useQuery({
    queryKey: ["dashboard"],
    queryFn: () => api.getDashboard(),
  });

  const topProducts = dashboard?.top_products ?? [];

  return (
    <div className="space-y-8">
      <PageHeader title="Reports" description="Analyze sales performance and inventory health." />

      <div className="flex flex-wrap items-center gap-2">
        {[7, 30, 90].map((d) => (
          <Button key={d} variant={days === d ? "default" : "outline"} size="sm" onClick={() => setDays(d)}>
            {d} days
          </Button>
        ))}
        <Button variant="outline" size="sm" onClick={() => api.exportSales(days)}>Export Sales CSV</Button>
        <Button variant="outline" size="sm" onClick={() => api.exportInventory()}>Export Inventory CSV</Button>
      </div>

      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <ChartLineUp size={18} className="text-primary-brand" />
            <CardTitle>Sales Report</CardTitle>
          </div>
          <CardDescription>Revenue over the last {days} days</CardDescription>
        </CardHeader>
        <CardContent>
          {salesLoading ? <Skeleton className="h-64" /> : (
            <ResponsiveContainer width="100%" height={300}>
              <AreaChart data={sales ?? []}>
                <CartesianGrid strokeDasharray="3 3" stroke={colors.grid} />
                <XAxis dataKey="date" tickFormatter={(v) => v.slice(5)} tick={{ fontSize: 12, fill: colors.tick }} />
                <YAxis tick={{ fontSize: 12, fill: colors.tick }} />
                <Tooltip
                  contentStyle={{ borderRadius: 8, border: `1px solid ${colors.tooltipBorder}`, background: colors.tooltipBg }}
                  formatter={(v) => [formatCurrency(Number(v ?? 0)), "Revenue"]}
                />
                <Area type="monotone" dataKey="revenue" stroke={colors.bar} fill={colors.bar} fillOpacity={0.15} />
              </AreaChart>
            </ResponsiveContainer>
          )}
        </CardContent>
      </Card>

      <div className="grid gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader><CardTitle>Top Products</CardTitle></CardHeader>
          <CardContent>
            {topProducts.length === 0 ? (
              <EmptyState icon={Medal} title="No sales data" description="Top products will appear after orders." />
            ) : (
              <div className="space-y-3">
                {topProducts.map((p, i) => (
                  <div key={p.product_id} className="flex items-center justify-between rounded-lg bg-surface-muted px-3 py-2">
                    <span className="text-sm font-medium">#{i + 1} {p.product_name}</span>
                    <span className="text-sm font-semibold">{formatCurrency(p.revenue)}</span>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <div className="flex items-center gap-2">
              <Warning size={18} className="text-rose-500" />
              <CardTitle>Low Stock ({(lowStock ?? []).length})</CardTitle>
            </div>
          </CardHeader>
          <CardContent>
            {invLoading ? <Skeleton className="h-40" /> : (lowStock ?? []).length === 0 ? (
              <EmptyState icon={Warning} title="Stock levels OK" description="All products above threshold." />
            ) : (
              <div className="space-y-2">
                {(lowStock ?? []).map((p) => (
                  <div key={p.id} className="flex justify-between rounded-lg border border-rose-100 bg-rose-50 px-3 py-2 text-sm">
                    <span>{p.name}</span>
                    <span className="font-semibold text-rose-600">{p.stock_on_hand} left</span>
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
