"use client";

import { use } from "react";
import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { PageHeader } from "@/components/layout/page-header";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { formatCurrency, formatDate } from "@/lib/utils";
import { ArrowLeft } from "@phosphor-icons/react";

export default function OrderDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params);

  const { data: order, isLoading } = useQuery({
    queryKey: ["order", id],
    queryFn: () => api.getOrder(id),
  });

  if (isLoading) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-10 w-64" />
        <Skeleton className="h-48" />
      </div>
    );
  }

  if (!order) return <p className="text-sm text-muted">Order not found</p>;

  return (
    <div className="space-y-6">
      <PageHeader
        title={order.order_number}
        description={`Created ${formatDate(order.created_at)}`}
        action={
          <Link href="/orders">
            <Button variant="outline"><ArrowLeft size={16} /> Back</Button>
          </Link>
        }
      />

      <div className="grid gap-6 lg:grid-cols-3">
        <Card className="lg:col-span-2">
          <CardHeader><CardTitle>Order Items</CardTitle></CardHeader>
          <CardContent>
            <div className="space-y-3">
              {(order.items ?? []).map((item) => (
                <div key={item.id} className="flex items-center justify-between rounded-lg bg-surface-muted px-4 py-3">
                  <div>
                    <p className="font-medium text-foreground">{item.product_name}</p>
                    <p className="font-mono text-xs text-muted">{item.product_sku}</p>
                  </div>
                  <div className="text-right">
                    <p className="text-sm text-muted">{item.quantity} × {formatCurrency(item.unit_price)}</p>
                    <p className="font-semibold">{formatCurrency(item.line_total)}</p>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader><CardTitle>Summary</CardTitle></CardHeader>
          <CardContent className="space-y-3">
            <div className="flex justify-between"><span className="text-muted">Status</span><Badge status={order.status} /></div>
            <div className="flex justify-between"><span className="text-muted">Customer</span><span>{order.customer_name || "—"}</span></div>
            <div className="flex justify-between"><span className="text-muted">Subtotal</span><span>{formatCurrency(order.subtotal)}</span></div>
            <div className="flex justify-between border-t pt-3 font-semibold"><span>Total</span><span>{formatCurrency(order.total)}</span></div>
            {order.notes && <p className="text-sm text-muted">Notes: {order.notes}</p>}
          </CardContent>
        </Card>
      </div>

      {(order.status_history ?? []).length > 0 && (
        <Card>
          <CardHeader><CardTitle>Status Timeline</CardTitle></CardHeader>
          <CardContent>
            <div className="relative space-y-0">
              {(order.status_history ?? []).map((entry, i) => (
                <div key={entry.id} className="flex gap-4 pb-6 last:pb-0">
                  <div className="flex flex-col items-center">
                    <div className="flex h-8 w-8 items-center justify-center rounded-full bg-primary-light text-xs font-bold text-primary-brand">
                      {i + 1}
                    </div>
                    {i < (order.status_history ?? []).length - 1 && (
                      <div className="mt-1 w-px flex-1 bg-border" />
                    )}
                  </div>
                  <div className="min-w-0 flex-1 pt-0.5">
                    <div className="flex flex-wrap items-center gap-2">
                      {entry.from_status && (
                        <>
                          <Badge status={entry.from_status} />
                          <span className="text-muted-light">→</span>
                        </>
                      )}
                      <Badge status={entry.to_status} />
                    </div>
                    <p className="mt-1 text-xs text-muted">
                      {formatDate(entry.created_at)}
                      {entry.changed_by_name && ` · ${entry.changed_by_name}`}
                    </p>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
