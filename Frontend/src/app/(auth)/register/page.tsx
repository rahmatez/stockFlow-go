"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import {
  Buildings,
  User,
  EnvelopeSimple,
  Lock,
  ArrowRight,
  WarningCircle,
  CircleNotch,
} from "@phosphor-icons/react";
import type { Icon } from "@phosphor-icons/react";
import { api } from "@/lib/api/client";
import { setTokens } from "@/lib/auth";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Logo } from "@/components/brand/logo";
import { cn } from "@/lib/utils";

const schema = z.object({
  tenant_name: z.string().min(2, "Nama perusahaan wajib diisi"),
  full_name: z.string().min(2, "Nama lengkap wajib diisi"),
  email: z.string().email("Masukkan email yang valid"),
  password: z.string().min(8, "Password minimal 8 karakter"),
});

type FormData = z.infer<typeof schema>;

export default function RegisterPage() {
  const router = useRouter();
  const [error, setError] = useState("");
  const { register, handleSubmit, formState: { isSubmitting, errors } } = useForm<FormData>({
    resolver: zodResolver(schema),
  });

  const onSubmit = async (data: FormData) => {
    setError("");
    try {
      const res = await api.register(data);
      setTokens(res.tokens.access_token, res.tokens.refresh_token);
      router.push("/dashboard");
    } catch (e) {
      setError(e instanceof Error ? e.message : "Gagal mendaftar. Coba lagi.");
    }
  };

  const Field = ({
    name, label, type = "text", icon: Icon, placeholder,
  }: {
    name: keyof FormData; label: string; type?: string;
    icon: Icon; placeholder: string;
  }) => (
    <div>
      <Label htmlFor={name}>{label}</Label>
      <div className="relative">
        <Icon
          className="pointer-events-none absolute left-3.5 top-1/2 h-[18px] w-[18px] -translate-y-1/2 text-muted-light"
          weight="duotone"
        />
        <Input
          id={name}
          type={type}
          className={cn("h-11 pl-10", errors[name] && "border-danger-border focus:border-danger-text focus:ring-danger-text/15")}
          {...register(name)}
          placeholder={placeholder}
        />
      </div>
      {errors[name] && <p className="mt-1.5 text-xs text-danger-text">{errors[name]?.message}</p>}
    </div>
  );

  return (
    <div className="animate-fade-in">
      <div className="rounded-xl border border-border bg-card p-7 sm:p-8">
        <div className="mb-8 lg:hidden">
          <Logo size="md" />
        </div>

        <div className="mb-8">
          <h1 className="text-2xl font-semibold tracking-tight text-foreground">
            Daftar
          </h1>
          <p className="mt-2 text-sm text-muted">
            Buat akun untuk mulai mengelola bisnis Anda.
          </p>
        </div>

        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          {error && (
            <div className="flex items-start gap-3 rounded-xl border border-danger-border bg-danger-bg px-4 py-3.5">
              <WarningCircle size={18} className="mt-0.5 shrink-0 text-danger-text" weight="fill" />
              <p className="text-sm leading-relaxed text-danger-text">{error}</p>
            </div>
          )}

          <Field name="tenant_name" label="Nama perusahaan" icon={Buildings} placeholder="PT Maju Jaya" />
          <Field name="full_name" label="Nama lengkap" icon={User} placeholder="Budi Santoso" />
          <Field name="email" label="Email kerja" type="email" icon={EnvelopeSimple} placeholder="nama@perusahaan.com" />
          <Field name="password" label="Password" type="password" icon={Lock} placeholder="Min. 8 karakter" />

          <Button type="submit" className="mt-2 h-12 w-full text-[15px]" size="lg" disabled={isSubmitting}>
            {isSubmitting ? (
              <>
                <CircleNotch size={18} className="animate-spin" />
                Membuat workspace...
              </>
            ) : (
              <>
                Mulai sekarang
                <ArrowRight size={18} weight="bold" />
              </>
            )}
          </Button>
        </form>
      </div>

      <p className="mt-6 text-center text-sm text-muted">
        Sudah punya akun?{" "}
        <Link
          href="/login"
          className="font-semibold text-primary-brand underline-offset-4 transition-colors hover:text-primary-hover hover:underline"
        >
          Masuk
        </Link>
      </p>
    </div>
  );
}
