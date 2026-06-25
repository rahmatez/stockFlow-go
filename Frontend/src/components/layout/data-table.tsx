import { ReactNode } from "react";
import { Card } from "@/components/ui/card";
import { TableSkeleton } from "@/components/ui/skeleton";
import { EmptyState } from "@/components/dashboard/empty-state";
import type { Icon } from "@phosphor-icons/react";

interface Column {
  key: string;
  label: string;
  align?: "left" | "center" | "right";
  className?: string;
}

interface DataTableProps {
  columns: Column[];
  children: ReactNode;
  isLoading?: boolean;
  isEmpty?: boolean;
  emptyIcon?: Icon;
  emptyTitle?: string;
  emptyDescription?: string;
}

export function DataTable({
  columns,
  children,
  isLoading,
  isEmpty,
  emptyIcon,
  emptyTitle = "No data found",
  emptyDescription = "Get started by adding your first item.",
}: DataTableProps) {
  return (
    <Card className="overflow-hidden">
      <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-border-subtle bg-surface-muted/80">
              {columns.map((col) => (
                <th
                  key={col.key}
                  className={`px-5 py-3.5 text-xs font-semibold uppercase tracking-wider text-muted ${
                    col.align === "right" ? "text-right" : col.align === "center" ? "text-center" : "text-left"
                  } ${col.className ?? ""}`}
                >
                  {col.label}
                </th>
              ))}
            </tr>
          </thead>
          <tbody className="divide-y divide-border-subtle">
            {isLoading ? (
              <tr>
                <td colSpan={columns.length}>
                  <TableSkeleton rows={4} cols={columns.length} />
                </td>
              </tr>
            ) : isEmpty ? (
              <tr>
                <td colSpan={columns.length}>
                  {emptyIcon ? (
                    <EmptyState icon={emptyIcon} title={emptyTitle} description={emptyDescription} />
                  ) : (
                    <p className="py-12 text-center text-sm text-muted">{emptyTitle}</p>
                  )}
                </td>
              </tr>
            ) : (
              children
            )}
          </tbody>
        </table>
      </div>
    </Card>
  );
}

export function TableRow({ children, className }: { children: ReactNode; className?: string }) {
  return (
    <tr className={`group transition-colors hover:bg-surface-muted/60 ${className ?? ""}`}>
      {children}
    </tr>
  );
}

export function TableCell({
  children,
  align = "left",
  className,
}: {
  children: ReactNode;
  align?: "left" | "center" | "right";
  className?: string;
}) {
  const alignClass =
    align === "right" ? "text-right" : align === "center" ? "text-center" : "text-left";
  return (
    <td className={`px-5 py-4 text-foreground/90 ${alignClass} ${className ?? ""}`}>
      {children}
    </td>
  );
}
