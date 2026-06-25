import { AuthGuard } from "@/components/layout/auth-guard";
import { Sidebar } from "@/components/layout/sidebar";
import { TopBar } from "@/components/layout/top-bar";

export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  return (
    <AuthGuard>
      <div className="flex min-h-screen bg-background">
        <Sidebar />
        <div className="flex min-w-0 flex-1 flex-col">
          <TopBar />
          <main className="flex-1 overflow-auto p-4 md:p-6">
            <div className="mx-auto max-w-[1400px] animate-fade-in">{children}</div>
          </main>
        </div>
      </div>
    </AuthGuard>
  );
}
