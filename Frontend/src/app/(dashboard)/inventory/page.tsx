"use client";

import { useState } from "react";
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
import { DataTable, TableRow, TableCell } from "@/components/layout/data-table";
import { formatDate } from "@/lib/utils";
import { Plus, Archive, ArrowsLeftRight } from "@phosphor-icons/react";

export default function InventoryPage() {
  const [showForm, setShowForm] = useState(false);
  const [showTransfer, setShowTransfer] = useState(false);
  const [form, setForm] = useState({ product_id: "", warehouse_id: "", movement_type: "IN", quantity: 0, notes: "" });
  const [transfer, setTransfer] = useState({ product_id: "", from_warehouse_id: "", to_warehouse_id: "", quantity: 0, notes: "" });
  const queryClient = useQueryClient();
  const { success, error: toastError } = useToast();

  const { data: movements, isLoading } = useQuery({ queryKey: ["inventory"], queryFn: () => api.listInventory() });
  const { data: products } = useQuery({ queryKey: ["products"], queryFn: () => api.listProducts() });
  const { data: warehouses } = useQuery({ queryKey: ["warehouses"], queryFn: () => api.listWarehouses() });

  const adjustMutation = useMutation({
    mutationFn: () => api.adjustInventory({
      product_id: form.product_id,
      movement_type: form.movement_type,
      quantity: form.quantity,
      warehouse_id: form.warehouse_id || undefined,
      notes: form.notes || undefined,
    }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["inventory"] });
      queryClient.invalidateQueries({ queryKey: ["products"] });
      setShowForm(false);
      success("Stok berhasil disesuaikan");
    },
    onError: (err) => toastError(getErrorMessage(err)),
  });

  const transferMutation = useMutation({
    mutationFn: () => api.transferInventory({
      product_id: transfer.product_id,
      from_warehouse_id: transfer.from_warehouse_id,
      to_warehouse_id: transfer.to_warehouse_id,
      quantity: transfer.quantity,
      notes: transfer.notes || undefined,
    }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["inventory"] });
      setShowTransfer(false);
      success("Transfer stok berhasil");
    },
    onError: (err) => toastError(getErrorMessage(err)),
  });

  const items = movements?.data ?? [];
  const whList = warehouses ?? [];

  return (
    <div className="space-y-6">
      <PageHeader
        title="Inventory"
        description="Track stock movements, adjustments, and warehouse activity."
        action={
          <div className="flex gap-2">
            <Button variant="outline" onClick={() => setShowTransfer(!showTransfer)}>
              <ArrowsLeftRight size={16} /> Transfer
            </Button>
            <Button onClick={() => setShowForm(!showForm)}>
              <Plus size={16} weight="bold" /> {showForm ? "Cancel" : "Adjust Stock"}
            </Button>
          </div>
        }
      />

      {showForm && (
        <FormPanel title="Stock Adjustment" onClose={() => setShowForm(false)}
          footer={<><Button variant="outline" onClick={() => setShowForm(false)}>Cancel</Button>
            <Button onClick={() => adjustMutation.mutate()} disabled={adjustMutation.isPending || !form.product_id}>Apply</Button></>}>
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="sm:col-span-2">
              <Label>Product</Label>
              <Select value={form.product_id} onChange={(e) => setForm({ ...form, product_id: e.target.value })}>
                <option value="">Select product</option>
                {(products?.data ?? []).map((p) => <option key={p.id} value={p.id}>{p.name} ({p.stock_on_hand} in stock)</option>)}
              </Select>
            </div>
            <div>
              <Label>Warehouse</Label>
              <Select value={form.warehouse_id} onChange={(e) => setForm({ ...form, warehouse_id: e.target.value })}>
                <option value="">Default warehouse</option>
                {whList.map((w) => <option key={w.id} value={w.id}>{w.name}</option>)}
              </Select>
            </div>
            <div>
              <Label>Type</Label>
              <Select value={form.movement_type} onChange={(e) => setForm({ ...form, movement_type: e.target.value })}>
                <option value="IN">Stock In</option>
                <option value="OUT">Stock Out</option>
                <option value="ADJUSTMENT">Adjustment</option>
              </Select>
            </div>
            <div><Label>Quantity</Label><Input type="number" min={1} value={form.quantity || ""} onChange={(e) => setForm({ ...form, quantity: Number(e.target.value) })} /></div>
            <div><Label>Notes</Label><Input value={form.notes} onChange={(e) => setForm({ ...form, notes: e.target.value })} /></div>
          </div>
        </FormPanel>
      )}

      {showTransfer && (
        <FormPanel title="Transfer Stock" onClose={() => setShowTransfer(false)}
          footer={<><Button variant="outline" onClick={() => setShowTransfer(false)}>Cancel</Button>
            <Button onClick={() => transferMutation.mutate()} disabled={transferMutation.isPending}>Transfer</Button></>}>
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="sm:col-span-2">
              <Label>Product</Label>
              <Select value={transfer.product_id} onChange={(e) => setTransfer({ ...transfer, product_id: e.target.value })}>
                <option value="">Select product</option>
                {(products?.data ?? []).map((p) => <option key={p.id} value={p.id}>{p.name}</option>)}
              </Select>
            </div>
            <div>
              <Label>From</Label>
              <Select value={transfer.from_warehouse_id} onChange={(e) => setTransfer({ ...transfer, from_warehouse_id: e.target.value })}>
                <option value="">Select warehouse</option>
                {whList.map((w) => <option key={w.id} value={w.id}>{w.name}</option>)}
              </Select>
            </div>
            <div>
              <Label>To</Label>
              <Select value={transfer.to_warehouse_id} onChange={(e) => setTransfer({ ...transfer, to_warehouse_id: e.target.value })}>
                <option value="">Select warehouse</option>
                {whList.map((w) => <option key={w.id} value={w.id}>{w.name}</option>)}
              </Select>
            </div>
            <div><Label>Quantity</Label><Input type="number" min={1} value={transfer.quantity || ""} onChange={(e) => setTransfer({ ...transfer, quantity: Number(e.target.value) })} /></div>
          </div>
        </FormPanel>
      )}

      <DataTable
        columns={[
          { key: "date", label: "Date" }, { key: "product", label: "Product" },
          { key: "warehouse", label: "Warehouse" }, { key: "type", label: "Type", align: "center" },
          { key: "qty", label: "Qty", align: "right" },
        ]}
        isLoading={isLoading}
        isEmpty={!isLoading && items.length === 0}
        emptyIcon={Archive}
        emptyTitle="No movements yet"
        emptyDescription="Adjust stock to see movement history."
      >
        {items.map((m) => (
          <TableRow key={m.id}>
            <TableCell className="text-muted">{formatDate(m.created_at)}</TableCell>
            <TableCell><p className="font-medium">{m.product_name}</p><p className="font-mono text-xs text-muted-light">{m.product_sku}</p></TableCell>
            <TableCell className="text-muted">{m.warehouse_name}</TableCell>
            <TableCell align="center"><Badge status={m.movement_type === "OUT" ? "cancelled" : "active"} /></TableCell>
            <TableCell align="right" className="font-semibold">{m.quantity}</TableCell>
          </TableRow>
        ))}
      </DataTable>
    </div>
  );
}
