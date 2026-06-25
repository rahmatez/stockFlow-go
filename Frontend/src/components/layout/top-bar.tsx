"use client";

import { useRouter } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { MagnifyingGlass } from "@phosphor-icons/react";
import { api } from "@/lib/api/client";
import { useState } from "react";
import { NotificationPanel } from "@/components/layout/notification-panel";
import { ThemeToggle } from "@/components/layout/theme-toggle";

export function TopBar() {
  const router = useRouter();
  const [search, setSearch] = useState("");
  const [notifOpen, setNotifOpen] = useState(false);

  const { data: user } = useQuery({
    queryKey: ["user-me"],
    queryFn: () => api.getUserMe(),
  });

  const { data: notifData } = useQuery({
    queryKey: ["notifications"],
    queryFn: () => api.listNotifications(),
    refetchInterval: 60_000,
  });

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    if (!search.trim()) return;
    router.push(`/products?search=${encodeURIComponent(search.trim())}`);
  };

  const initials = user?.full_name?.charAt(0)?.toUpperCase() || "A";

  return (
    <header className="sticky top-0 z-10 flex h-[72px] items-center justify-between gap-4 border-b border-border bg-card px-6">
      <form onSubmit={handleSearch} className="relative hidden max-w-md flex-1 md:block">
        <MagnifyingGlass
          className="pointer-events-none absolute left-4 top-1/2 h-[18px] w-[18px] -translate-y-1/2 text-muted-light"
          weight="bold"
        />
        <input
          type="search"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="Search products..."
          className="h-11 w-full rounded-full border border-border bg-input-bg-muted pl-11 pr-4 text-sm text-foreground placeholder:text-muted-light focus:border-primary-brand focus:bg-input-bg focus:outline-none focus:ring-2 focus:ring-primary-brand/15"
        />
      </form>

      <div className="ml-auto flex items-center gap-3">
        <ThemeToggle />
        <NotificationPanel
          notifications={notifData?.notifications ?? []}
          unreadCount={notifData?.unread_count ?? 0}
          open={notifOpen}
          onToggle={() => setNotifOpen((v) => !v)}
          onClose={() => setNotifOpen(false)}
        />

        <div className="flex items-center gap-3 rounded-full border border-border bg-card py-1.5 pl-1.5 pr-4">
          <div className="flex h-9 w-9 items-center justify-center rounded-full bg-primary-light text-sm font-bold text-primary-brand">
            {initials}
          </div>
          <div className="hidden sm:block">
            <p className="text-sm font-semibold text-foreground">{user?.full_name || "Loading..."}</p>
            <p className="text-xs capitalize text-muted-light">{user?.role || ""}</p>
          </div>
        </div>
      </div>
    </header>
  );
}
