import { cn } from "@/lib/utils";
import type { Icon } from "@phosphor-icons/react";

interface StatCardProps {
  label: string;
  value: string | number;
  icon: Icon;
  trend?: { value: string; positive: boolean };
}

export function StatCard({ label, value, icon: Icon, trend }: StatCardProps) {
  return (
    <div className="rounded-xl border border-border bg-card p-5 shadow-sm">
      <div className="flex items-center justify-between">
        <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-surface-muted ring-1 ring-border">
          <Icon size={24} weight="regular" className="text-muted" />
        </div>
        {trend && (
          <span
            className={cn(
              "flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-semibold",
              trend.positive ? "bg-success-bg text-success-text" : "bg-danger-bg text-danger-text"
            )}
          >
            {trend.value}
          </span>
        )}
      </div>
      <div className="mt-4">
        <p className="text-2xl font-bold tracking-tight text-foreground">{value}</p>
        <p className="mt-1 text-sm text-muted">{label}</p>
      </div>
    </div>
  );
}
