"use client";

import { createContext, useCallback, useContext, useState } from "react";
import { X, CheckCircle, WarningCircle, Info } from "@phosphor-icons/react";
import { cn } from "@/lib/utils";

type ToastType = "success" | "error" | "info";

interface Toast {
  id: string;
  message: string;
  type: ToastType;
}

interface ToastContextValue {
  toast: (message: string, type?: ToastType) => void;
  success: (message: string) => void;
  error: (message: string) => void;
}

const ToastContext = createContext<ToastContextValue | null>(null);

const icons = {
  success: CheckCircle,
  error: WarningCircle,
  info: Info,
};

const styles = {
  success: "border-emerald-500/30 bg-success-bg text-success-text",
  error: "border-danger-border bg-danger-bg text-danger-text",
  info: "border-primary-brand/30 bg-primary-light text-primary-brand",
};

export function ToastProvider({ children }: { children: React.ReactNode }) {
  const [toasts, setToasts] = useState<Toast[]>([]);

  const dismiss = useCallback((id: string) => {
    setToasts((prev) => prev.filter((t) => t.id !== id));
  }, []);

  const addToast = useCallback((message: string, type: ToastType = "info") => {
    const id = crypto.randomUUID();
    setToasts((prev) => [...prev, { id, message, type }]);
    setTimeout(() => dismiss(id), 5000);
  }, [dismiss]);

  const value: ToastContextValue = {
    toast: addToast,
    success: (msg) => addToast(msg, "success"),
    error: (msg) => addToast(msg, "error"),
  };

  return (
    <ToastContext.Provider value={value}>
      {children}
      <div className="pointer-events-none fixed bottom-4 right-4 z-50 flex flex-col gap-2">
        {toasts.map((t) => {
          const Icon = icons[t.type];
          return (
            <div
              key={t.id}
              className={cn(
                "pointer-events-auto flex items-center gap-3 rounded-lg border px-4 py-3 text-sm shadow-lg",
                styles[t.type]
              )}
            >
              <Icon size={18} weight="fill" className="shrink-0" />
              <span className="flex-1">{t.message}</span>
              <button onClick={() => dismiss(t.id)} className="shrink-0 opacity-60 hover:opacity-100">
                <X size={14} weight="bold" />
              </button>
            </div>
          );
        })}
      </div>
    </ToastContext.Provider>
  );
}

export function useToast() {
  const ctx = useContext(ToastContext);
  if (!ctx) throw new Error("useToast must be used within ToastProvider");
  return ctx;
}
