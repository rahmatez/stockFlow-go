import type { Icon } from "@phosphor-icons/react";

interface EmptyStateProps {
  icon: Icon;
  title: string;
  description: string;
}

export function EmptyState({ icon: Icon, title, description }: EmptyStateProps) {
  return (
    <div className="flex flex-col items-center justify-center py-14 text-center">
      <div className="mb-4 flex h-14 w-14 items-center justify-center rounded-2xl bg-surface-muted text-muted-light ring-1 ring-border">
        <Icon size={28} weight="duotone" />
      </div>
      <p className="text-sm font-semibold text-foreground">{title}</p>
      <p className="mt-1.5 max-w-xs text-sm leading-relaxed text-muted">{description}</p>
    </div>
  );
}
