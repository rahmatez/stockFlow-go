"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api, getErrorMessage } from "@/lib/api/client";
import { useToast } from "@/components/ui/toast";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { PageHeader } from "@/components/layout/page-header";

export default function SettingsPage() {
  const queryClient = useQueryClient();
  const { success, error: toastError } = useToast();
  const { data } = useQuery({ queryKey: ["tenant"], queryFn: () => api.getTenant() });
  const [name, setName] = useState("");

  const tenant = data?.tenant as { name?: string; slug?: string } | undefined;

  const updateMutation = useMutation({
    mutationFn: () => api.updateTenant({ name: name || tenant?.name || "" }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["tenant"] });
      success("Nama bisnis diperbarui");
    },
    onError: (err) => toastError(getErrorMessage(err)),
  });

  return (
    <div className="space-y-6">
      <PageHeader title="Settings" description="Konfigurasi workspace bisnis Anda." />
      <Card>
        <CardHeader><CardTitle>Business Info</CardTitle></CardHeader>
        <CardContent className="space-y-4 max-w-md">
          <div><Label>Workspace Slug</Label><Input value={tenant?.slug || ""} disabled /></div>
          <div>
            <Label>Business Name</Label>
            <Input
              defaultValue={tenant?.name}
              onChange={(e) => setName(e.target.value)}
              placeholder={tenant?.name}
            />
          </div>
          <Button onClick={() => updateMutation.mutate()} disabled={updateMutation.isPending}>Save</Button>
        </CardContent>
      </Card>
    </div>
  );
}
