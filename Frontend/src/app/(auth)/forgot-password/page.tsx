"use client";

import { useState } from "react";
import Link from "next/link";
import { api, getErrorMessage } from "@/lib/api/client";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Logo } from "@/components/brand/logo";

export default function ForgotPasswordPage() {
  const [email, setEmail] = useState("");
  const [token, setToken] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");
  const [step, setStep] = useState<"request" | "reset">("request");

  const handleRequest = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    try {
      const res = await api.forgotPassword(email);
      setMessage(res.message);
      setStep("reset");
    } catch (err) {
      setError(getErrorMessage(err));
    }
  };

  const handleReset = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    try {
      const res = await api.resetPassword(token, newPassword);
      setMessage(res.message);
    } catch (err) {
      setError(getErrorMessage(err));
    }
  };

  return (
    <div className="animate-fade-in">
      <div className="rounded-xl border border-border bg-card p-7 sm:p-8">
        <div className="mb-8 lg:hidden"><Logo size="md" /></div>
        <h1 className="text-2xl font-semibold text-foreground">Reset Password</h1>
        <p className="mt-2 text-sm text-muted">
          {step === "request" ? "Masukkan email untuk menerima token reset." : "Masukkan token dan password baru."}
        </p>

        {error && <p className="mt-4 text-sm text-rose-600">{error}</p>}
        {message && <p className="mt-4 text-sm text-emerald-600">{message}</p>}

        {step === "request" ? (
          <form onSubmit={handleRequest} className="mt-6 space-y-4">
            <div><Label>Email</Label><Input type="email" value={email} onChange={(e) => setEmail(e.target.value)} required /></div>
            <Button type="submit" className="w-full">Kirim Token Reset</Button>
          </form>
        ) : (
          <form onSubmit={handleReset} className="mt-6 space-y-4">
            <div><Label>Reset Token</Label><Input value={token} onChange={(e) => setToken(e.target.value)} required /></div>
            <div><Label>Password Baru</Label><Input type="password" value={newPassword} onChange={(e) => setNewPassword(e.target.value)} required minLength={8} /></div>
            <Button type="submit" className="w-full">Reset Password</Button>
          </form>
        )}

        <p className="mt-6 text-center text-sm text-muted">
          <Link href="/login" className="font-medium text-primary-brand hover:underline">Kembali ke login</Link>
        </p>
      </div>
    </div>
  );
}
