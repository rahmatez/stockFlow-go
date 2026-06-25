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
import { SearchBar } from "@/components/layout/search-bar";
import { Pagination } from "@/components/layout/pagination";
import { DataTable, TableRow, TableCell } from "@/components/layout/data-table";
import { formatCurrency } from "@/lib/utils";
import { Plus, Tag, Trash, PencilSimple } from "@phosphor-icons/react";
import type { Product } from "@/lib/api/types";

const emptyForm = { sku: "", name: "", sell_price: 0, cost_price: 0, low_stock_threshold: 10, category_id: "", description: "", is_active: true };

export default function ProductsPage() {
  const [search, setSearch] = useState("");
  const [page, setPage] = useState(1);
  const [showForm, setShowForm] = useState(false);
  const [editId, setEditId] = useState<string | null>(null);
  const [form, setForm] = useState(emptyForm);
  const [showCategories, setShowCategories] = useState(false);
  const [catName, setCatName] = useState("");
  const queryClient = useQueryClient();
  const { success, error: toastError } = useToast();

  const { data, isLoading } = useQuery({
    queryKey: ["products", search, page],
    queryFn: () => api.listProducts({ search, page }),
  });

  const { data: categories } = useQuery({
    queryKey: ["categories"],
    queryFn: () => api.listCategories(),
  });

  const saveMutation = useMutation({
    mutationFn: () => {
      const body = {
        ...form,
        category_id: form.category_id || undefined,
        description: form.description || undefined,
      };
      return editId ? api.updateProduct(editId, body) : api.createProduct(body);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["products"] });
      setShowForm(false);
      setEditId(null);
      setForm(emptyForm);
      success(editId ? "Produk diperbarui" : "Produk ditambahkan");
    },
    onError: (err) => toastError(getErrorMessage(err)),
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => api.deleteProduct(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["products"] });
      success("Produk dihapus");
    },
    onError: (err) => toastError(getErrorMessage(err)),
  });

  const createCatMutation = useMutation({
    mutationFn: () => api.createCategory({ name: catName }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["categories"] });
      setCatName("");
      success("Kategori ditambahkan");
    },
    onError: (err) => toastError(getErrorMessage(err)),
  });

  const deleteCatMutation = useMutation({
    mutationFn: (id: string) => api.deleteCategory(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["categories"] }),
    onError: (err) => toastError(getErrorMessage(err)),
  });

  const products = data?.data ?? [];
  const openEdit = (p: Product) => {
    setEditId(p.id);
    setForm({
      sku: p.sku, name: p.name, sell_price: p.sell_price, cost_price: p.cost_price,
      low_stock_threshold: p.low_stock_threshold, category_id: p.category_id || "",
      description: p.description || "", is_active: p.is_active,
    });
    setShowForm(true);
  };

  return (
    <div className="space-y-6">
      <PageHeader
        title="Products"
        description="Manage your product catalog, pricing, and stock levels."
        action={
          <div className="flex gap-2">
            <Button variant="outline" onClick={() => api.exportProducts()}>Export CSV</Button>
            <Button variant="outline" onClick={() => setShowCategories(!showCategories)}>Categories</Button>
            <Button onClick={() => { setShowForm(true); setEditId(null); setForm(emptyForm); }}>
              <Plus size={16} weight="bold" /> Add Product
            </Button>
          </div>
        }
      />

      {showCategories && (
        <FormPanel title="Categories" onClose={() => setShowCategories(false)}>
          <div className="flex gap-2">
            <Input value={catName} onChange={(e) => setCatName(e.target.value)} placeholder="Category name" />
            <Button onClick={() => createCatMutation.mutate()} disabled={!catName}>Add</Button>
          </div>
          <div className="mt-3 space-y-2">
            {(categories ?? []).map((c) => (
              <div key={c.id} className="flex items-center justify-between rounded-lg bg-surface-muted px-3 py-2">
                <span className="text-sm font-medium">{c.name}</span>
                <Button size="sm" variant="destructive" onClick={() => deleteCatMutation.mutate(c.id)}><Trash size={12} /></Button>
              </div>
            ))}
          </div>
        </FormPanel>
      )}

      {showForm && (
        <FormPanel
          title={editId ? "Edit Product" : "New Product"}
          onClose={() => setShowForm(false)}
          footer={
            <>
              <Button variant="outline" onClick={() => setShowForm(false)}>Cancel</Button>
              <Button onClick={() => saveMutation.mutate()} disabled={saveMutation.isPending}>
                {saveMutation.isPending ? "Saving..." : "Save Product"}
              </Button>
            </>
          }
        >
          <div className="grid gap-4 sm:grid-cols-2">
            <div><Label>SKU</Label><Input value={form.sku} onChange={(e) => setForm({ ...form, sku: e.target.value })} disabled={!!editId} /></div>
            <div><Label>Product Name</Label><Input value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} /></div>
            <div><Label>Category</Label>
              <Select value={form.category_id} onChange={(e) => setForm({ ...form, category_id: e.target.value })}>
                <option value="">No category</option>
                {(categories ?? []).map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}
              </Select>
            </div>
            <div><Label>Low Stock Threshold</Label><Input type="number" value={form.low_stock_threshold} onChange={(e) => setForm({ ...form, low_stock_threshold: Number(e.target.value) })} /></div>
            <div><Label>Sell Price (IDR)</Label><Input type="number" value={form.sell_price || ""} onChange={(e) => setForm({ ...form, sell_price: Number(e.target.value) })} /></div>
            <div><Label>Cost Price (IDR)</Label><Input type="number" value={form.cost_price || ""} onChange={(e) => setForm({ ...form, cost_price: Number(e.target.value) })} /></div>
            <div className="sm:col-span-2"><Label>Description</Label><Input value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} /></div>
          </div>
        </FormPanel>
      )}

      <SearchBar value={search} onChange={(v) => { setSearch(v); setPage(1); }} placeholder="Search by name or SKU..." />

      <DataTable
        columns={[
          { key: "sku", label: "SKU" }, { key: "name", label: "Product" },
          { key: "price", label: "Price", align: "right" }, { key: "stock", label: "Stock", align: "right" },
          { key: "status", label: "Status", align: "center" }, { key: "actions", label: "", align: "right" },
        ]}
        isLoading={isLoading}
        isEmpty={!isLoading && products.length === 0}
        emptyIcon={Tag}
        emptyTitle="No products yet"
        emptyDescription="Add your first product to start managing inventory."
      >
        {products.map((p) => (
          <TableRow key={p.id}>
            <TableCell className="font-mono text-xs text-muted">{p.sku}</TableCell>
            <TableCell>
              <p className="font-medium text-foreground">{p.name}</p>
              {p.category_name && <p className="text-xs text-muted-light">{p.category_name}</p>}
            </TableCell>
            <TableCell align="right" className="font-medium">{formatCurrency(p.sell_price)}</TableCell>
            <TableCell align="right">
              <span className={p.stock_on_hand <= p.low_stock_threshold ? "font-semibold text-rose-600" : ""}>{p.stock_on_hand}</span>
            </TableCell>
            <TableCell align="center"><Badge status={p.is_active ? "active" : "cancelled"} /></TableCell>
            <TableCell align="right">
              <div className="flex justify-end gap-1.5">
                <Button size="sm" variant="outline" onClick={() => openEdit(p)}><PencilSimple size={14} /></Button>
                <Button variant="destructive" size="sm" onClick={() => deleteMutation.mutate(p.id)}><Trash size={14} weight="bold" /></Button>
              </div>
            </TableCell>
          </TableRow>
        ))}
      </DataTable>

      {data?.meta && <Pagination page={page} total={data.meta.total} limit={data.meta.limit} onPageChange={setPage} />}
    </div>
  );
}
