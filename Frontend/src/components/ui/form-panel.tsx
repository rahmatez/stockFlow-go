import { ReactNode } from "react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { X } from "@phosphor-icons/react";
import { Button } from "@/components/ui/button";

interface FormPanelProps {
  title: string;
  description?: string;
  onClose?: () => void;
  children: ReactNode;
  footer?: ReactNode;
}

export function FormPanel({ title, description, onClose, children, footer }: FormPanelProps) {
  return (
    <Card className="animate-fade-in border-ring-accent ring-1 ring-primary-light">
      <CardHeader className="flex flex-row items-start justify-between">
        <div>
          <CardTitle>{title}</CardTitle>
          {description && <CardDescription>{description}</CardDescription>}
        </div>
        {onClose && (
          <Button variant="ghost" size="icon" onClick={onClose} className="-mr-2 -mt-1">
            <X size={16} weight="bold" />
          </Button>
        )}
      </CardHeader>
      <CardContent className="space-y-4">{children}</CardContent>
      {footer && (
        <div className="flex justify-end gap-2 border-t border-border-subtle px-6 py-4">
          {footer}
        </div>
      )}
    </Card>
  );
}
