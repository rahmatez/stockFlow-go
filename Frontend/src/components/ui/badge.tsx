import { cn } from "@/lib/utils";

const statusStyles: Record<string, string> = {
  draft: "bg-surface-muted text-muted ring-border",
  confirmed: "bg-primary-light text-primary-brand ring-primary-brand/20",
  processing: "bg-amber-500/10 text-amber-600 ring-amber-500/20 dark:text-amber-400",
  shipped: "bg-violet-500/10 text-violet-600 ring-violet-500/20 dark:text-violet-400",
  delivered: "bg-success-bg text-success-text ring-success-text/20",
  cancelled: "bg-danger-bg text-danger-text ring-danger-border",
  active: "bg-success-bg text-success-text ring-success-text/20",
  in: "bg-success-bg text-success-text ring-success-text/20",
  out: "bg-danger-bg text-danger-text ring-danger-border",
  adjustment: "bg-amber-500/10 text-amber-600 ring-amber-500/20 dark:text-amber-400",
  transfer: "bg-violet-500/10 text-violet-600 ring-violet-500/20 dark:text-violet-400",
};

export function Badge({ status, className }: { status: string; className?: string }) {
  const key = status.toLowerCase();
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-medium capitalize ring-1 ring-inset",
        statusStyles[key] || "bg-surface-muted text-muted ring-border",
        className
      )}
    >
      <span className="h-1.5 w-1.5 rounded-full bg-current opacity-60" />
      {status}
    </span>
  );
}
