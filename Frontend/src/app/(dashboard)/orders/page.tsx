"use client";

import { useState } from "react";
import Link from "next/link";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api, getErrorMessage } from "@/lib/api/client";
import { useToast } from "@/components/ui/toast";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select } from "@/components/ui/select";
import { Badge } from "@/components/ui/badge";
import { FormPanel } from "@/components/ui/form-panel";
import { PageHeader } from "@/components/layout/page-header";
import { Pagination } from "@/components/layout/pagination";
import { DataTable, TableRow, TableCell } from "@/components/layout/data-table";
import { formatCurrency, formatDate } from "@/lib/utils";
import { Plus, Receipt, ArrowRight, X, Tag, Eye } from "@phosphor-icons/react";

const nextStatus: Record<string, string> = {
  draft: "confirmed", confirmed: "processing", processing: "shipped", shipped: "delivered",
};

export default function OrdersPage() {
  const [showForm, setShowForm] = useState(false);
  const [page, setPage] = useState(1);
  const [selectedProduct, setSelectedProduct] = useState("");
  const [selectedCustomer, setSelectedCustomer] = useState("");
  const [quantity, setQuantity] = useState(1);
  const [orderItems, setOrderItems] = useState<{ product_id: string; quantity: number; name: string }[]>([]);
  const queryClient = useQueryClient();
  const { success, error: toastError } = useToast();

  const { data, isLoading } = useQuery({
    queryKey: ["orders", page],
    queryFn: () => api.listOrders({ page }),
  });

  const { data: products } = useQuery({ queryKey: ["products"], queryFn: () => api.listProducts() });
  const { data: customers } = useQuery({ queryKey: ["customers"], queryFn: () => api.listCustomers() });

  const createMutation = useMutation({
    mutationFn: () => api.createOrder({
      customer_id: selectedCustomer || undefined,
      items: orderItems.map((i) => ({ product_id: i.product_id, quantity: i.quantity })),
    }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["orders"] });
      setShowForm(false);
      setOrderItems([]);
      setSelectedCustomer("");
      success("Pesanan berhasil dibuat");
    },
    onError: (err) => toastError(getErrorMessage(err)),
  });

  const statusMutation = useMutation({
    mutationFn: ({ id, status }: { id: string; status: string }) => api.updateOrderStatus(id, status),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["orders"] });
      queryClient.invalidateQueries({ queryKey: ["products"] });
      queryClient.invalidateQueries({ queryKey: ["inventory"] });
      queryClient.invalidateQueries({ queryKey: ["dashboard"] });
      success("Status pesanan diperbarui");
    },
    onError: (err) => toastError(getErrorMessage(err)),
  });

  const addItem = () => {
    const product = (products?.data ?? []).find((p) => p.id === selectedProduct);
    if (!product) return;
    setOrderItems([...orderItems, { product_id: product.id, quantity, name: product.name }]);
    setSelectedProduct("");
    setQuantity(1);
  };

  const orders = data?.data ?? [];

  return (
    <div className="space-y-6">
      <PageHeader
        title="Orders"
        description="Create and manage customer orders through fulfillment."
        action={
          <Button onClick={() => setShowForm(!showForm)}>
            <Plus size={16} weight="bold" />
            {showForm ? "Cancel" : "Create Order"}
          </Button>
        }
      />

      {showForm && (
        <FormPanel
          title="New Order"
          description="Add products and create a draft order."
          onClose={() => setShowForm(false)}
          footer={
            <>
              <Button variant="outline" onClick={() => setShowForm(false)}>Cancel</Button>
              <Button onClick={() => createMutation.mutate()} disabled={orderItems.length === 0 || createMutation.isPending}>
                {createMutation.isPending ? "Creating..." : "Create Order"}
              </Button>
            </>
          }
        >
          <div className="mb-4">
            <Label>Customer (optional)</Label>
            <Select value={selectedCustomer} onChange={(e) => setSelectedCustomer(e.target.value)}>
              <option value="">No customer</option>
              {(customers?.data ?? []).map((c) => (
                <option key={c.id} value={c.id}>{c.name}</option>
              ))}
            </Select>
          </div>
          <div className="flex flex-col gap-3 sm:flex-row">
            <div className="flex-1">
              <Label>Product</Label>
              <Select value={selectedProduct} onChange={(e) => setSelectedProduct(e.target.value)}>
                <option value="">Select product</option>
                {(products?.data ?? []).map((p) => (
                  <option key={p.id} value={p.id}>{p.name} — {formatCurrency(p.sell_price)}</option>
                ))}
              </Select>
            </div>
            <div className="w-24">
              <Label>Qty</Label>
              <Input type="number" min={1} value={quantity} onChange={(e) => setQuantity(Number(e.target.value))} />
            </div>
            <div className="flex items-end"><Button onClick={addItem} disabled={!selectedProduct}>Add</Button></div>
          </div>
          {orderItems.length > 0 && (
            <div className="mt-4 rounded-xl border border-border-subtle bg-surface-muted/80 p-4">
              <div className="space-y-2">
                {orderItems.map((item, i) => (
                  <div key={i} className="flex items-center justify-between text-sm">
                    <span className="flex items-center gap-2 text-foreground/90"><Tag size={14} />{item.name}</span>
                    <span className="font-medium">× {item.quantity}</span>
                  </div>
                ))}
              </div>
            </div>
          )}
        </FormPanel>
      )}

      <DataTable
        columns={[
          { key: "order", label: "Order" }, { key: "customer", label: "Customer" },
          { key: "date", label: "Date" }, { key: "status", label: "Status", align: "center" },
          { key: "total", label: "Total", align: "right" }, { key: "actions", label: "", align: "right" },
        ]}
        isLoading={isLoading}
        isEmpty={!isLoading && orders.length === 0}
        emptyIcon={Receipt}
        emptyTitle="No orders yet"
        emptyDescription="Create your first order to start selling."
      >
        {orders.map((o) => (
          <TableRow key={o.id}>
            <TableCell>
              <Link href={`/orders/${o.id}`} className="font-mono text-sm font-medium text-primary-brand hover:underline">
                {o.order_number}
              </Link>
            </TableCell>
            <TableCell className="text-muted">{o.customer_name || "—"}</TableCell>
            <TableCell className="text-muted">{formatDate(o.created_at)}</TableCell>
            <TableCell align="center"><Badge status={o.status} /></TableCell>
            <TableCell align="right" className="font-semibold">{formatCurrency(o.total)}</TableCell>
            <TableCell align="right">
              <div className="flex justify-end gap-1.5">
                <Link href={`/orders/${o.id}`}><Button size="sm" variant="outline"><Eye size={14} /></Button></Link>
                {nextStatus[o.status] && (
                  <Button size="sm" onClick={() => statusMutation.mutate({ id: o.id, status: nextStatus[o.status] })}>
                    <ArrowRight size={14} weight="bold" />{nextStatus[o.status]}
                  </Button>
                )}
                {o.status !== "cancelled" && o.status !== "delivered" && (
                  <Button variant="destructive" size="sm" onClick={() => statusMutation.mutate({ id: o.id, status: "cancelled" })}>
                    <X size={14} weight="bold" />
                  </Button>
                )}
              </div>
            </TableCell>
          </TableRow>
        ))}
      </DataTable>

      {data?.meta && <Pagination page={page} total={data.meta.total} limit={data.meta.limit} onPageChange={setPage} />}
    </div>
  );
}
