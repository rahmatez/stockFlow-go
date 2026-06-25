"use client";

import { Logo } from "@/components/brand/logo";

export function AuthSidePanel() {
  return (
    <div className="hidden w-[45%] flex-col justify-between border-r border-border bg-surface-muted px-12 py-10 lg:flex xl:px-16 xl:py-12">
      <Logo size="md" />

      <div className="max-w-sm space-y-5">
        <h2 className="text-2xl font-semibold leading-snug tracking-tight text-foreground xl:text-3xl">
          Kelola stok dan pesanan dalam satu tempat.
        </h2>
        <p className="text-sm leading-relaxed text-muted">
          StockFlow membantu tim Anda mencatat produk, memantau inventori, dan memproses pesanan dengan lebih rapi.
        </p>
        <ul className="space-y-2.5 pt-2 text-sm text-muted">
          <li className="flex items-center gap-2">
            <span className="h-1.5 w-1.5 shrink-0 rounded-full bg-muted-light" />
            Manajemen produk & stok
          </li>
          <li className="flex items-center gap-2">
            <span className="h-1.5 w-1.5 shrink-0 rounded-full bg-muted-light" />
            Alur pesanan yang jelas
          </li>
          <li className="flex items-center gap-2">
            <span className="h-1.5 w-1.5 shrink-0 rounded-full bg-muted-light" />
            Laporan penjualan harian
          </li>
        </ul>
      </div>

      <p className="text-xs text-muted-light">© 2026 StockFlow</p>
    </div>
  );
}
