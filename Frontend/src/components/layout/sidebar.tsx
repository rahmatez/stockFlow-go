"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  SquaresFour,
  Tag,
  Archive,
  Receipt,
  UsersThree,
  ChartLineUp,
  CreditCard,
  SignOut,
  Warehouse,
  Gear,
  User,
  Users,
} from "@phosphor-icons/react";
import type { Icon } from "@phosphor-icons/react";
import { cn } from "@/lib/utils";
import { api } from "@/lib/api/client";
import { clearTokens } from "@/lib/auth";
import { Logo } from "@/components/brand/logo";

const mainNav: { href: string; label: string; icon: Icon }[] = [
  { href: "/dashboard", label: "Dashboard", icon: SquaresFour },
  { href: "/products", label: "Products", icon: Tag },
  { href: "/inventory", label: "Inventory", icon: Archive },
  { href: "/orders", label: "Orders", icon: Receipt },
  { href: "/customers", label: "Customers", icon: UsersThree },
  { href: "/warehouses", label: "Warehouses", icon: Warehouse },
];

const secondaryNav: { href: string; label: string; icon: Icon }[] = [
  { href: "/reports", label: "Reports", icon: ChartLineUp },
  { href: "/settings", label: "Settings", icon: Gear },
  { href: "/settings/profile", label: "Profile", icon: User },
  { href: "/settings/team", label: "Team", icon: Users },
  { href: "/settings/billing", label: "Billing", icon: CreditCard },
];

export function Sidebar() {
  const pathname = usePathname();

  const handleLogout = async () => {
    try {
      await api.logout();
    } finally {
      clearTokens();
      window.location.href = "/login";
    }
  };

  const NavItem = ({ href, label, icon: Icon }: { href: string; label: string; icon: Icon }) => {
    const active = pathname.startsWith(href);
    return (
      <Link
        href={href}
        className={cn(
          "group flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition-colors",
          active
            ? "bg-primary-light text-primary-brand"
            : "text-muted hover:bg-surface-hover hover:text-foreground"
        )}
      >
        <Icon
          size={20}
          weight={active ? "fill" : "regular"}
          className={active ? "text-primary-brand" : "text-muted-light group-hover:text-muted"}
        />
        {label}
      </Link>
    );
  };

  return (
    <aside className="sticky top-0 flex h-screen w-[290px] shrink-0 flex-col border-r border-border bg-card">
      <div className="px-6 py-7">
        <Logo />
      </div>

      <nav className="flex-1 space-y-6 overflow-y-auto px-4">
        <div>
          <p className="mb-3 px-3 text-xs font-semibold uppercase tracking-wider text-muted-light">
            Menu
          </p>
          <div className="space-y-1">
            {mainNav.map((item) => (
              <NavItem key={item.href} {...item} />
            ))}
          </div>
        </div>
        <div>
          <p className="mb-3 px-3 text-xs font-semibold uppercase tracking-wider text-muted-light">
            Others
          </p>
          <div className="space-y-1">
            {secondaryNav.map((item) => (
              <NavItem key={item.href} {...item} />
            ))}
          </div>
        </div>
      </nav>

      <div className="border-t border-border p-4">
        <button
          onClick={handleLogout}
          className="flex w-full items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium text-muted transition-colors hover:bg-surface-hover hover:text-foreground"
        >
          <SignOut size={20} />
          Sign out
        </button>
      </div>
    </aside>
  );
}
