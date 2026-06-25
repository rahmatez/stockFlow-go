import { cn } from "@/lib/utils";
import { SelectHTMLAttributes, forwardRef } from "react";

export const Select = forwardRef<HTMLSelectElement, SelectHTMLAttributes<HTMLSelectElement>>(
  ({ className, children, ...props }, ref) => (
    <select
      ref={ref}
      className={cn(
        "flex h-10 w-full appearance-none rounded-lg border border-border bg-input-bg px-3.5 py-2 text-sm text-foreground transition-colors hover:border-border focus:border-primary-brand focus:outline-none focus:ring-2 focus:ring-primary-brand/15",
        className
      )}
      {...props}
    >
      {children}
    </select>
  )
);
Select.displayName = "Select";
