"use client";

import { useEffect, useRef } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Bell } from "@phosphor-icons/react";
import { api } from "@/lib/api/client";

interface Notification {
  id: string;
  title: string;
  message: string;
  is_read: boolean;
  created_at: string;
}

interface NotificationPanelProps {
  notifications: Notification[];
  unreadCount: number;
  open: boolean;
  onToggle: () => void;
  onClose: () => void;
}

function formatRelativeTime(value: string) {
  const diff = Date.now() - new Date(value).getTime();
  const mins = Math.floor(diff / 60_000);
  if (mins < 1) return "Baru saja";
  if (mins < 60) return `${mins} menit lalu`;
  const hours = Math.floor(mins / 60);
  if (hours < 24) return `${hours} jam lalu`;
  const days = Math.floor(hours / 24);
  return `${days} hari lalu`;
}

export function NotificationPanel({
  notifications,
  unreadCount,
  open,
  onToggle,
  onClose,
}: NotificationPanelProps) {
  const panelRef = useRef<HTMLDivElement>(null);
  const queryClient = useQueryClient();

  const markReadMutation = useMutation({
    mutationFn: (id: string) => api.markNotificationRead(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["notifications"] });
    },
  });

  useEffect(() => {
    if (!open) return;
    const handleClick = (e: MouseEvent) => {
      if (panelRef.current && !panelRef.current.contains(e.target as Node)) {
        onClose();
      }
    };
    document.addEventListener("mousedown", handleClick);
    return () => document.removeEventListener("mousedown", handleClick);
  }, [open, onClose]);

  return (
    <div ref={panelRef} className="relative">
      <button
        type="button"
        onClick={onToggle}
        className="relative flex h-10 w-10 items-center justify-center rounded-full border border-border bg-card text-muted transition-colors hover:border-border hover:bg-surface-hover hover:text-foreground"
        aria-label="Notifications"
        aria-expanded={open}
      >
        <Bell size={18} />
        {unreadCount > 0 && (
          <span className="absolute right-2 top-2 flex h-4 w-4 items-center justify-center rounded-full bg-primary-brand text-[10px] font-bold text-white ring-2 ring-card">
            {unreadCount > 9 ? "9+" : unreadCount}
          </span>
        )}
      </button>

      {open && (
        <div className="absolute right-0 top-12 z-50 w-80 overflow-hidden rounded-xl border border-border bg-card shadow-lg">
          <div className="border-b border-border px-4 py-3">
            <p className="text-sm font-semibold text-foreground">Notifikasi</p>
          </div>
          <div className="max-h-80 overflow-y-auto">
            {notifications.length === 0 ? (
              <p className="px-4 py-8 text-center text-sm text-muted-light">Tidak ada notifikasi</p>
            ) : (
              notifications.map((n) => (
                <button
                  key={n.id}
                  type="button"
                  onClick={() => {
                    if (!n.is_read) markReadMutation.mutate(n.id);
                  }}
                  className={`w-full border-b border-border-subtle px-4 py-3 text-left transition-colors last:border-0 hover:bg-surface-hover ${
                    !n.is_read ? "bg-primary-light/40" : ""
                  }`}
                >
                  <p className="text-sm font-medium text-foreground">{n.title}</p>
                  <p className="mt-0.5 line-clamp-2 text-xs text-muted">{n.message}</p>
                  <p className="mt-1 text-[10px] text-muted-light">{formatRelativeTime(n.created_at)}</p>
                </button>
              ))
            )}
          </div>
        </div>
      )}
    </div>
  );
}
