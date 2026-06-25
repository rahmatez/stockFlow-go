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

export default function ProfilePage() {
  const queryClient = useQueryClient();
  const { success, error: toastError } = useToast();
  const { data: user } = useQuery({ queryKey: ["user-me"], queryFn: () => api.getUserMe() });
  const [fullName, setFullName] = useState("");
  const [oldPassword, setOldPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");

  const updateProfile = useMutation({
    mutationFn: () => api.updateUserMe({ full_name: fullName || user?.full_name }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["user-me"] });
      success("Profil diperbarui");
    },
    onError: (err) => toastError(getErrorMessage(err)),
  });

  const updatePassword = useMutation({
    mutationFn: () => api.updateUserMe({ old_password: oldPassword, new_password: newPassword }),
    onSuccess: () => {
      setOldPassword("");
      setNewPassword("");
      success("Password diperbarui");
    },
    onError: (err) => toastError(getErrorMessage(err)),
  });

  return (
    <div className="space-y-6">
      <PageHeader title="Profile" description="Kelola informasi akun Anda." />
      <div className="grid gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader><CardTitle>Personal Info</CardTitle></CardHeader>
          <CardContent className="space-y-4">
            <div><Label>Email</Label><Input value={user?.email || ""} disabled /></div>
            <div>
              <Label>Full Name</Label>
              <Input
                defaultValue={user?.full_name}
                onChange={(e) => setFullName(e.target.value)}
                placeholder={user?.full_name}
              />
            </div>
            <Button onClick={() => updateProfile.mutate()} disabled={updateProfile.isPending}>Save</Button>
          </CardContent>
        </Card>
        <Card>
          <CardHeader><CardTitle>Change Password</CardTitle></CardHeader>
          <CardContent className="space-y-4">
            <div><Label>Current Password</Label><Input type="password" value={oldPassword} onChange={(e) => setOldPassword(e.target.value)} /></div>
            <div><Label>New Password</Label><Input type="password" value={newPassword} onChange={(e) => setNewPassword(e.target.value)} /></div>
            <Button onClick={() => updatePassword.mutate()} disabled={updatePassword.isPending || !oldPassword || !newPassword}>Update Password</Button>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
