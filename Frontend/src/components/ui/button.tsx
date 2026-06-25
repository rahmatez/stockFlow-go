import { cn } from "@/lib/utils";
import { ButtonHTMLAttributes, forwardRef } from "react";

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: "default" | "outline" | "ghost" | "destructive" | "secondary";
  size?: "sm" | "md" | "lg" | "icon";
}

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant = "default", size = "md", ...props }, ref) => {
    const variants = {
      default:
        "bg-primary-brand text-white shadow-sm hover:bg-primary-hover active:scale-[0.98]",
      secondary:
        "bg-background text-muted hover:bg-surface-hover active:scale-[0.98]",
      outline:
        "border border-border bg-card text-muted hover:border-border hover:bg-surface-hover",
      ghost: "text-muted hover:bg-surface-muted hover:text-foreground",
      destructive:
        "bg-danger-bg text-danger-text border border-danger-border hover:opacity-90",
    };
    const sizes = {
      sm: "h-8 gap-1.5 rounded-lg px-3 text-xs font-medium",
      md: "h-10 gap-2 rounded-lg px-4 text-sm font-medium",
      lg: "h-11 gap-2 rounded-lg px-6 text-sm font-semibold",
      icon: "h-9 w-9 rounded-lg",
    };
    return (
      <button
        ref={ref}
        className={cn(
          "inline-flex items-center justify-center transition-all duration-200 disabled:pointer-events-none disabled:opacity-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary-brand/30",
          variants[variant],
          sizes[size],
          className
        )}
        {...props}
      />
    );
  }
);
Button.displayName = "Button";
