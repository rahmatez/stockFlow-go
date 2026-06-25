"use client";

import { Moon, Sun } from "@phosphor-icons/react";
import { useTheme } from "@/components/theme-provider";
import { cn } from "@/lib/utils";

interface ThemeToggleProps {
  className?: string;
}

export function ThemeToggle({ className }: ThemeToggleProps) {
  const { theme, toggleTheme } = useTheme();
  const isDark = theme === "dark";

  return (
    <button
      type="button"
      onClick={toggleTheme}
      aria-label={isDark ? "Switch to light mode" : "Switch to dark mode"}
      className={cn(
        "flex h-10 w-10 items-center justify-center rounded-full border border-border bg-card text-muted transition-colors hover:border-border hover:bg-surface-hover hover:text-foreground",
        className
      )}
    >
      {isDark ? <Sun size={18} weight="fill" /> : <Moon size={18} weight="fill" />}
    </button>
  );
}
