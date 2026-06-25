"use client";

import { MagnifyingGlass } from "@phosphor-icons/react";
import { Input } from "@/components/ui/input";

interface SearchBarProps {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  className?: string;
}

export function SearchBar({ value, onChange, placeholder = "Search...", className }: SearchBarProps) {
  return (
    <div className={`relative ${className ?? "max-w-sm"}`}>
      <MagnifyingGlass className="pointer-events-none absolute left-3.5 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-light" weight="bold" />
      <Input
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
        className="pl-10"
      />
    </div>
  );
}
