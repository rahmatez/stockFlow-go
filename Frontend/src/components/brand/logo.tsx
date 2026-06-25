import { cn } from "@/lib/utils";

interface LogoProps {
  size?: "sm" | "md" | "lg";
  showText?: boolean;
  className?: string;
}

const sizes = {
  sm: { box: "h-8 w-8", icon: 16, title: "text-sm", sub: "text-[10px]" },
  md: { box: "h-9 w-9", icon: 18, title: "text-lg", sub: "text-[11px]" },
  lg: { box: "h-11 w-11", icon: 22, title: "text-xl", sub: "text-xs" },
};

export function Logo({ size = "md", showText = true, className }: LogoProps) {
  const s = sizes[size];

  return (
    <div className={cn("flex items-center gap-2.5", className)}>
      <div
        className={cn(
          "flex shrink-0 items-center justify-center rounded-lg bg-primary-brand shadow-sm shadow-primary-brand/30",
          s.box
        )}
      >
        <svg viewBox="0 0 32 32" fill="none" width={s.icon} height={s.icon} aria-hidden>
          <rect x="5" y="15" width="9" height="9" rx="1.5" fill="#fff" fillOpacity="0.95" />
          <rect x="18" y="9" width="9" height="9" rx="1.5" fill="#fff" fillOpacity="0.75" />
          <rect x="18" y="21" width="9" height="5" rx="1" fill="#a5b4fc" />
        </svg>
      </div>
      {showText && (
        <span className={cn("font-bold tracking-tight text-foreground", s.title)}>
          StockFlow
        </span>
      )}
    </div>
  );
}
