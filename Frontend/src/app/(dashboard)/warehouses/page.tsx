"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api, getErrorMessage } from "@/lib/api/client";
import { useToast } from "@/components/ui/toast";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { FormPanel } from "@/components/ui/form-panel";
import { PageHeader } from "@/components/layout/page-header";
import { DataTable, TableRow, TableCell } from "@/components/layout/data-table";
import { Plus, Warehouse as WarehouseIcon, Trash, PencilSimple } from "@phosphor-icons/react";

export default function WarehousesPage() {
  const [showForm, setShowForm] = useState(false);
  const [editId, setEditId] = useState<string | null>(null);
  const [name, setName] = useState("");
  const queryClient = useQueryClient();
  const { success, error: toastError } = useToast();

  const { data: warehouses, isLoading } = useQuery({
    queryKey: ["warehouses"],
    queryFn: () => api.listWarehouses(),
  });

  const saveMutation = useMutation({
    mutationFn: () =>
      editId ? api.updateWarehouse(editId, { name }) : api.createWarehouse({ name }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["warehouses"] });
      setShowForm(false);
      setEditId(null);
      setName("");
      success(editId ? "Gudang diperbarui" : "Gudang ditambahkan");
    },
    onError: (err) => toastError(getErrorMessage(err)),
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => api.deleteWarehouse(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["warehouses"] });
      success("Gudang dihapus");
    },
    onError: (err) => toastError(getErrorMessage(err)),
  });

  const items = warehouses ?? [];

  return (
    <div className="space-y-6">
      <PageHeader
        title="Warehouses"
        description="Kelola lokasi penyimpanan stok."
        action={
          <Button onClick={() => { setShowForm(true); setEditId(null); setName(""); }}>
            <Plus size={16} weight="bold" />
            Add Warehouse
          </Button>
        }
      />

      {showForm && (
        <FormPanel
          title={editId ? "Edit Warehouse" : "New Warehouse"}
          onClose={() => setShowForm(false)}
          footer={
            <>
              <Button variant="outline" onClick={() => setShowForm(false)}>Cancel</Button>
              <Button onClick={() => saveMutation.mutate()} disabled={!name || saveMutation.isPending}>
                {saveMutation.isPending ? "Saving..." : "Save"}
              </Button>
            </>
          }
        >
          <div>
            <Label>Name</Label>
            <Input value={name} onChange={(e) => setName(e.target.value)} placeholder="Warehouse B" />
          </div>
        </FormPanel>
      )}

      <DataTable
        columns={[
          { key: "name", label: "Name" },
          { key: "default", label: "Default", align: "center" },
          { key: "actions", label: "", align: "right" },
        ]}
        isLoading={isLoading}
        isEmpty={!isLoading && items.length === 0}
        emptyIcon={WarehouseIcon}
        emptyTitle="No warehouses"
        emptyDescription="Default warehouse is created on registration."
      >
        {items.map((w) => (
          <TableRow key={w.id}>
            <TableCell className="font-medium">{w.name}</TableCell>
            <TableCell align="center">
              {w.is_default ? <Badge status="active" /> : <span className="text-muted-light">—</span>}
            </TableCell>
            <TableCell align="right">
              <div className="flex justify-end gap-1.5">
                <Button size="sm" variant="outline" onClick={() => { setEditId(w.id); setName(w.name); setShowForm(true); }}>
                  <PencilSimple size={14} />
                </Button>
                {!w.is_default && (
                  <Button size="sm" variant="destructive" onClick={() => deleteMutation.mutate(w.id)}>
                    <Trash size={14} weight="bold" />
                  </Button>
                )}
              </div>
            </TableCell>
          </TableRow>
        ))}
      </DataTable>
    </div>
  );
}
