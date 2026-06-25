"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api, getErrorMessage } from "@/lib/api/client";
import { useToast } from "@/components/ui/toast";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { FormPanel } from "@/components/ui/form-panel";
import { PageHeader } from "@/components/layout/page-header";
import { SearchBar } from "@/components/layout/search-bar";
import { Pagination } from "@/components/layout/pagination";
import { DataTable, TableRow, TableCell } from "@/components/layout/data-table";
import { Plus, UsersThree, Trash, PencilSimple, EnvelopeSimple, Phone } from "@phosphor-icons/react";
import type { Customer } from "@/lib/api/types";

export default function CustomersPage() {
  const [search, setSearch] = useState("");
  const [page, setPage] = useState(1);
  const [showForm, setShowForm] = useState(false);
  const [editId, setEditId] = useState<string | null>(null);
  const [form, setForm] = useState({ name: "", email: "", phone: "" });
  const queryClient = useQueryClient();
  const { success, error: toastError } = useToast();

  const { data, isLoading } = useQuery({
    queryKey: ["customers", search, page],
    queryFn: () => api.listCustomers({ search, page }),
  });

  const saveMutation = useMutation({
    mutationFn: () => {
      const body = { name: form.name, email: form.email || undefined, phone: form.phone || undefined };
      return editId ? api.updateCustomer(editId, body) : api.createCustomer(body);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["customers"] });
      setShowForm(false);
      setEditId(null);
      setForm({ name: "", email: "", phone: "" });
      success(editId ? "Pelanggan diperbarui" : "Pelanggan ditambahkan");
    },
    onError: (err) => toastError(getErrorMessage(err)),
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => api.deleteCustomer(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["customers"] });
      success("Pelanggan dihapus");
    },
    onError: (err) => toastError(getErrorMessage(err)),
  });

  const openEdit = (c: Customer) => {
    setEditId(c.id);
    setForm({ name: c.name, email: c.email || "", phone: c.phone || "" });
    setShowForm(true);
  };

  const customers = data?.data ?? [];

  return (
    <div className="space-y-6">
      <PageHeader
        title="Customers"
        description="Keep track of your customers and their contact information."
        action={
          <Button onClick={() => { setShowForm(true); setEditId(null); setForm({ name: "", email: "", phone: "" }); }}>
            <Plus size={16} weight="bold" /> Add Customer
          </Button>
        }
      />

      {showForm && (
        <FormPanel
          title={editId ? "Edit Customer" : "New Customer"}
          onClose={() => setShowForm(false)}
          footer={
            <>
              <Button variant="outline" onClick={() => setShowForm(false)}>Cancel</Button>
              <Button onClick={() => saveMutation.mutate()} disabled={saveMutation.isPending}>
                {saveMutation.isPending ? "Saving..." : "Save Customer"}
              </Button>
            </>
          }
        >
          <div className="grid gap-4 sm:grid-cols-3">
            <div><Label>Full Name</Label><Input value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} /></div>
            <div><Label>Email</Label><Input type="email" value={form.email} onChange={(e) => setForm({ ...form, email: e.target.value })} /></div>
            <div><Label>Phone</Label><Input value={form.phone} onChange={(e) => setForm({ ...form, phone: e.target.value })} /></div>
          </div>
        </FormPanel>
      )}

      <SearchBar value={search} onChange={(v) => { setSearch(v); setPage(1); }} placeholder="Search customers..." />

      <DataTable
        columns={[
          { key: "name", label: "Name" }, { key: "email", label: "Email" },
          { key: "phone", label: "Phone" }, { key: "actions", label: "", align: "right" },
        ]}
        isLoading={isLoading}
        isEmpty={!isLoading && customers.length === 0}
        emptyIcon={UsersThree}
        emptyTitle="No customers yet"
        emptyDescription="Add your first customer to get started."
      >
        {customers.map((c) => (
          <TableRow key={c.id}>
            <TableCell className="font-medium text-foreground">{c.name}</TableCell>
            <TableCell className="text-muted">
              {c.email ? <span className="flex items-center gap-1.5"><EnvelopeSimple size={14} />{c.email}</span> : "—"}
            </TableCell>
            <TableCell className="text-muted">
              {c.phone ? <span className="flex items-center gap-1.5"><Phone size={14} />{c.phone}</span> : "—"}
            </TableCell>
            <TableCell align="right">
              <div className="flex justify-end gap-1.5">
                <Button size="sm" variant="outline" onClick={() => openEdit(c)}><PencilSimple size={14} /></Button>
                <Button variant="destructive" size="sm" onClick={() => deleteMutation.mutate(c.id)}><Trash size={14} weight="bold" /></Button>
              </div>
            </TableCell>
          </TableRow>
        ))}
      </DataTable>

      {data?.meta && <Pagination page={page} total={data.meta.total} limit={data.meta.limit} onPageChange={setPage} />}
    </div>
  );
}
