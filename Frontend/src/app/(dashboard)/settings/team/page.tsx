"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api, getErrorMessage } from "@/lib/api/client";
import { useToast } from "@/components/ui/toast";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select } from "@/components/ui/select";
import { FormPanel } from "@/components/ui/form-panel";
import { PageHeader } from "@/components/layout/page-header";
import { DataTable, TableRow, TableCell } from "@/components/layout/data-table";
import { Plus, Users, Trash } from "@phosphor-icons/react";

export default function TeamPage() {
  const [showForm, setShowForm] = useState(false);
  const [form, setForm] = useState({ email: "", password: "", full_name: "", role: "staff" });
  const queryClient = useQueryClient();
  const { success, error: toastError } = useToast();

  const { data: users, isLoading } = useQuery({ queryKey: ["users"], queryFn: () => api.listUsers() });

  const createMutation = useMutation({
    mutationFn: () => api.createUser(form),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ["users"] });
      setShowForm(false);
      setForm({ email: "", password: "", full_name: "", role: "staff" });
      if (data.email_sent) {
        success(`Undangan dikirim ke ${data.email}`);
      } else {
        success("Anggota tim ditambahkan");
      }
    },
    onError: (err) => toastError(getErrorMessage(err)),
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => api.deleteUser(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["users"] });
      success("Anggota dihapus");
    },
    onError: (err) => toastError(getErrorMessage(err)),
  });

  return (
    <div className="space-y-6">
      <PageHeader
        title="Team"
        description="Kelola anggota tim workspace Anda."
        action={<Button onClick={() => setShowForm(true)}><Plus size={16} weight="bold" /> Add Member</Button>}
      />

      {showForm && (
        <FormPanel title="New Team Member" onClose={() => setShowForm(false)}
          footer={<><Button variant="outline" onClick={() => setShowForm(false)}>Cancel</Button>
            <Button onClick={() => createMutation.mutate()} disabled={createMutation.isPending}>Save</Button></>}>
          <div className="grid gap-4 sm:grid-cols-2">
            <div><Label>Full Name</Label><Input value={form.full_name} onChange={(e) => setForm({ ...form, full_name: e.target.value })} /></div>
            <div><Label>Email</Label><Input type="email" value={form.email} onChange={(e) => setForm({ ...form, email: e.target.value })} /></div>
            <div>
              <Label>Password</Label>
              <Input type="password" value={form.password} onChange={(e) => setForm({ ...form, password: e.target.value })} />
              <p className="mt-1 text-xs text-muted">Password sementara akan dikirim via email jika SMTP dikonfigurasi.</p>
            </div>
            <div>
              <Label>Role</Label>
              <Select value={form.role} onChange={(e) => setForm({ ...form, role: e.target.value })}>
                <option value="staff">Staff</option>
                <option value="admin">Admin</option>
              </Select>
            </div>
          </div>
        </FormPanel>
      )}

      <DataTable
        columns={[
          { key: "name", label: "Name" }, { key: "email", label: "Email" },
          { key: "role", label: "Role" }, { key: "actions", label: "", align: "right" },
        ]}
        isLoading={isLoading}
        isEmpty={!isLoading && (users ?? []).length === 0}
        emptyIcon={Users}
        emptyTitle="No team members"
        emptyDescription="Add team members to collaborate."
      >
        {(users ?? []).map((u) => (
          <TableRow key={u.id}>
            <TableCell className="font-medium">{u.full_name}</TableCell>
            <TableCell className="text-muted">{u.email}</TableCell>
            <TableCell className="capitalize">{u.role}</TableCell>
            <TableCell align="right">
              {u.role !== "owner" && (
                <Button size="sm" variant="destructive" onClick={() => deleteMutation.mutate(u.id)}>
                  <Trash size={14} weight="bold" />
                </Button>
              )}
            </TableCell>
          </TableRow>
        ))}
      </DataTable>
    </div>
  );
}
