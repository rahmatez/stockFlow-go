"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import {
  EnvelopeSimple,
  Lock,
  ArrowRight,
  Eye,
  EyeSlash,
  WarningCircle,
  CircleNotch,
} from "@phosphor-icons/react";
import { api } from "@/lib/api/client";
import { setTokens } from "@/lib/auth";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Logo } from "@/components/brand/logo";
import { cn } from "@/lib/utils";

const schema = z.object({
  email: z.string().email("Masukkan email yang valid"),
  password: z.string().min(8, "Password minimal 8 karakter"),
  tenant_slug: z.string().optional(),
});

type FormData = z.infer<typeof schema>;

const DEMO_EMAIL = "test@oms.local";
const DEMO_PASSWORD = "password123";

export default function LoginPage() {
  const router = useRouter();
  const [error, setError] = useState("");
  const [showPassword, setShowPassword] = useState(false);

  const {
    register,
    handleSubmit,
    setValue,
    formState: { isSubmitting, errors },
  } = useForm<FormData>({
    resolver: zodResolver(schema),
    defaultValues: { email: "", password: "", tenant_slug: "" },
  });

  const onSubmit = async (data: FormData) => {
    setError("");
    try {
      const res = await api.login(data);
      setTokens(res.tokens.access_token, res.tokens.refresh_token);
      router.push("/dashboard");
    } catch (e) {
      setError(e instanceof Error ? e.message : "Gagal masuk. Periksa email dan password Anda.");
    }
  };

  const fillDemo = () => {
    setValue("email", DEMO_EMAIL, { shouldValidate: true });
    setValue("password", DEMO_PASSWORD, { shouldValidate: true });
    setError("");
  };

  return (
    <div className="animate-fade-in">
      <div className="rounded-xl border border-border bg-card p-7 sm:p-8">
        <div className="mb-8 lg:hidden">
          <Logo size="md" />
        </div>

        <div className="mb-8">
          <h1 className="text-2xl font-semibold tracking-tight text-foreground">
            Masuk
          </h1>
          <p className="mt-2 text-sm text-muted">
            Masukkan email dan password untuk melanjutkan.
          </p>
        </div>

        <form onSubmit={handleSubmit(onSubmit)} className="space-y-5">
          {error && (
            <div className="flex items-start gap-3 rounded-xl border border-danger-border bg-danger-bg px-4 py-3.5">
              <WarningCircle size={18} className="mt-0.5 shrink-0 text-danger-text" weight="fill" />
              <p className="text-sm leading-relaxed text-danger-text">{error}</p>
            </div>
          )}

          <div>
            <Label htmlFor="email">Email</Label>
            <div className="relative">
              <EnvelopeSimple
                className="pointer-events-none absolute left-3.5 top-1/2 h-[18px] w-[18px] -translate-y-1/2 text-muted-light"
                weight="duotone"
              />
              <Input
                id="email"
                type="email"
                autoComplete="email"
                className={cn("h-11 pl-10", errors.email && "border-danger-border focus:border-danger-text focus:ring-danger-text/15")}
                {...register("email")}
                placeholder="nama@perusahaan.com"
              />
            </div>
            {errors.email && (
              <p className="mt-1.5 flex items-center gap-1 text-xs text-danger-text">
                {errors.email.message}
              </p>
            )}
          </div>

          <div>
            <Label htmlFor="tenant_slug">Workspace Slug (opsional)</Label>
            <Input
              id="tenant_slug"
              className="h-11"
              {...register("tenant_slug")}
              placeholder="demo-store-abc12345"
            />
            <p className="mt-1 text-xs text-muted-light">Diperlukan jika email terdaftar di beberapa workspace.</p>
          </div>

          <div>
            <div className="mb-1.5 flex items-center justify-between">
              <Label htmlFor="password" className="mb-0">Password</Label>
              <Link href="/forgot-password" className="text-xs font-medium text-primary-brand hover:underline">
                Lupa password?
              </Link>
            </div>
            <div className="relative">
              <Lock
                className="pointer-events-none absolute left-3.5 top-1/2 h-[18px] w-[18px] -translate-y-1/2 text-muted-light"
                weight="duotone"
              />
              <Input
                id="password"
                type={showPassword ? "text" : "password"}
                autoComplete="current-password"
                className={cn(
                  "h-11 pl-10 pr-11",
                  errors.password && "border-danger-border focus:border-danger-text focus:ring-danger-text/15"
                )}
                {...register("password")}
                placeholder="Masukkan password Anda"
              />
              <button
                type="button"
                onClick={() => setShowPassword((v) => !v)}
                className="absolute right-3 top-1/2 -translate-y-1/2 rounded-md p-1 text-muted-light transition-colors hover:bg-surface-muted hover:text-muted"
                aria-label={showPassword ? "Sembunyikan password" : "Tampilkan password"}
              >
                {showPassword ? <EyeSlash size={18} /> : <Eye size={18} />}
              </button>
            </div>
            {errors.password && (
              <p className="mt-1.5 text-xs text-danger-text">{errors.password.message}</p>
            )}
          </div>

          <Button type="submit" className="h-12 w-full text-[15px]" size="lg" disabled={isSubmitting}>
            {isSubmitting ? (
              <>
                <CircleNotch size={18} className="animate-spin" />
                Sedang masuk...
              </>
            ) : (
              <>
                Masuk ke dashboard
                <ArrowRight size={18} weight="bold" />
              </>
            )}
          </Button>
        </form>

        <div className="mt-6 rounded-lg border border-border bg-surface-muted px-4 py-3">
          <p className="text-xs text-muted">
            Demo: <span className="font-mono text-foreground/80">{DEMO_EMAIL}</span> / <span className="font-mono text-foreground/80">{DEMO_PASSWORD}</span>
          </p>
          <button
            type="button"
            onClick={fillDemo}
            className="mt-1.5 text-xs font-medium text-muted underline-offset-2 hover:text-foreground hover:underline"
          >
            Isi otomatis
          </button>
        </div>
      </div>

      <p className="mt-6 text-center text-sm text-muted">
        Belum punya akun?{" "}
        <Link
          href="/register"
          className="font-semibold text-primary-brand underline-offset-4 transition-colors hover:text-primary-hover hover:underline"
        >
          Daftar gratis
        </Link>
      </p>
    </div>
  );
}
