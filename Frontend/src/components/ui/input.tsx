import { cn } from "@/lib/utils";
import { InputHTMLAttributes, forwardRef } from "react";

export const Input = forwardRef<HTMLInputElement, InputHTMLAttributes<HTMLInputElement>>(
  ({ className, ...props }, ref) => (
    <input
      ref={ref}
      className={cn(
        "flex h-10 w-full rounded-lg border border-border bg-input-bg px-3.5 py-2 text-sm text-foreground transition-colors placeholder:text-muted-light hover:border-border focus:border-primary-brand focus:outline-none focus:ring-2 focus:ring-primary-brand/15",
        className
      )}
      {...props}
    />
  )
);
Input.displayName = "Input";
