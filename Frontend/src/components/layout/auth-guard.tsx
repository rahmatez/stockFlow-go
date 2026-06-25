"use client";

import { isAuthenticated } from "@/lib/auth";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { Logo } from "@/components/brand/logo";

export function AuthGuard({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const [ready, setReady] = useState(false);

  useEffect(() => {
    if (!isAuthenticated()) {
      router.replace("/login");
    } else {
      setReady(true);
    }
  }, [router]);

  if (!ready) {
    return (
      <div className="flex h-screen flex-col items-center justify-center gap-5 bg-background">
        <Logo size="lg" />
        <div className="h-1 w-32 overflow-hidden rounded-full bg-border">
          <div className="h-full w-1/2 animate-pulse rounded-full bg-primary-brand" />
        </div>
      </div>
    );
  }

  return <>{children}</>;
}
