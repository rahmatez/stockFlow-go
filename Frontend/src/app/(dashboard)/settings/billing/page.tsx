"use client";

import { useEffect } from "react";
import { useSearchParams } from "next/navigation";
import { useQuery, useMutation } from "@tanstack/react-query";
import { api, getErrorMessage } from "@/lib/api/client";
import { useToast } from "@/components/ui/toast";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { PageHeader } from "@/components/layout/page-header";
import { Skeleton } from "@/components/ui/skeleton";
import { Check, Lightning, Crown, Leaf } from "@phosphor-icons/react";
import type { Icon } from "@phosphor-icons/react";

const planIcons: Record<string, Icon> = {
  free: Leaf,
  starter: Lightning,
  pro: Crown,
};

export default function BillingPage() {
  const searchParams = useSearchParams();
  const { success, error: toastError } = useToast();

  useEffect(() => {
    if (searchParams.get("success") === "1") success("Langganan berhasil diperbarui!");
    if (searchParams.get("cancel") === "1") toastError("Checkout dibatalkan.");
  }, [searchParams, success, toastError]);

  const { data, isLoading } = useQuery({
    queryKey: ["billing"],
    queryFn: () => api.getBilling(),
  });

  const checkoutMutation = useMutation({
    mutationFn: (planSlug: string) => api.createCheckout(planSlug),
    onSuccess: (res) => { if (res.checkout_url) window.location.href = res.checkout_url; },
    onError: (err) => toastError(getErrorMessage(err)),
  });

  const portalMutation = useMutation({
    mutationFn: () => api.createBillingPortal(),
    onSuccess: (res) => { if (res.portal_url) window.location.href = res.portal_url; },
    onError: (err) => toastError(getErrorMessage(err)),
  });

  if (isLoading) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-10 w-48" />
        <Skeleton className="h-40" />
        <div className="grid gap-6 sm:grid-cols-3">
          {Array.from({ length: 3 }).map((_, i) => (
            <Skeleton key={i} className="h-72" />
          ))}
        </div>
      </div>
    );
  }

  if (!data) return <p className="text-sm text-muted">Failed to load billing</p>;

  const currentPlan = data.subscription.plan;
  const subStatus = data.subscription.status;

  return (
    <div className="space-y-8">
      <PageHeader
        title="Billing & Plans"
        description="Kelola langganan dan pantau batas penggunaan."
        action={
          currentPlan?.slug !== "free" && (
            <Button variant="outline" onClick={() => portalMutation.mutate()} disabled={portalMutation.isPending}>
              Manage Subscription
            </Button>
          )
        }
      />

      <Card className="overflow-hidden">
        <div className="bg-linear-to-r from-primary-brand to-primary-hover px-6 py-5 text-white">
          <p className="text-sm font-medium text-primary-light">Paket Aktif</p>
          <p className="text-2xl font-bold">{currentPlan?.name || "Free"}</p>
          {subStatus && subStatus !== "active" && (
            <p className="mt-1 text-sm text-amber-200 capitalize">Status: {subStatus}</p>
          )}
        </div>
        <CardContent className="grid gap-4 pt-6 sm:grid-cols-3">
          {[
            { label: "Produk", used: data.usage.products, max: currentPlan?.max_products || 50 },
            { label: "Pesanan / bulan", used: data.usage.orders_month, max: currentPlan?.max_orders_per_month || 100 },
            { label: "Anggota tim", used: data.usage.users, max: currentPlan?.max_users || 3 },
          ].map((item) => {
            const pct = Math.min((item.used / item.max) * 100, 100);
            return (
              <div key={item.label} className="rounded-xl bg-surface-muted p-4">
                <div className="flex items-center justify-between text-sm">
                  <span className="text-muted">{item.label}</span>
                  <span className="font-semibold text-foreground">{item.used} / {item.max}</span>
                </div>
                <div className="mt-2 h-2 overflow-hidden rounded-full bg-border">
                  <div
                    className={`h-full rounded-full transition-all ${pct > 80 ? "bg-rose-500" : "bg-primary-brand"}`}
                    style={{ width: `${pct}%` }}
                  />
                </div>
              </div>
            );
          })}
        </CardContent>
      </Card>

      <div className="grid gap-6 sm:grid-cols-3">
        {(data.plans ?? []).map((plan) => {
          const Icon = planIcons[plan.slug] || Leaf;
          const isCurrent = currentPlan?.slug === plan.slug;
          return (
            <Card
              key={plan.id}
              className={`relative overflow-hidden transition-all hover:shadow-md ${
                isCurrent ? "ring-2 ring-primary-brand shadow-md shadow-primary-light/50" : ""
              }`}
            >
              {isCurrent && (
                <div className="absolute right-4 top-4 rounded-full bg-primary-light px-2.5 py-0.5 text-[10px] font-bold uppercase tracking-wider text-primary-brand">
                  Aktif
                </div>
              )}
              <CardHeader>
                <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-primary-light text-primary-brand">
                  <Icon size={22} weight="duotone" />
                </div>
                <CardTitle className="mt-3">{plan.name}</CardTitle>
                <p className="text-3xl font-bold text-foreground">
                  {plan.price_cents === 0 ? (
                    <span>Gratis</span>
                  ) : (
                    <>
                      <span className="text-lg font-normal text-muted">Rp </span>
                      {(plan.price_cents / 100).toLocaleString("id-ID")}
                      <span className="text-sm font-normal text-muted">/bln</span>
                    </>
                  )}
                </p>
              </CardHeader>
              <CardContent>
                <ul className="space-y-2.5">
                  {[
                    `${plan.max_products.toLocaleString()} produk`,
                    `${plan.max_orders_per_month.toLocaleString()} pesanan/bulan`,
                    `${plan.max_users} anggota tim`,
                  ].map((feat) => (
                    <li key={feat} className="flex items-center gap-2 text-sm text-muted">
                      <Check size={14} weight="bold" className="shrink-0 text-emerald-500" />
                      {feat}
                    </li>
                  ))}
                </ul>
                {plan.slug !== "free" && !isCurrent && (
                  <Button
                    className="mt-6 w-full"
                    onClick={() => checkoutMutation.mutate(plan.slug)}
                    disabled={checkoutMutation.isPending}
                  >
                    Upgrade ke {plan.name}
                  </Button>
                )}
              </CardContent>
            </Card>
          );
        })}
      </div>
    </div>
  );
}
