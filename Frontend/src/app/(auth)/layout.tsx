import { AuthSidePanel } from "@/components/auth/auth-side-panel";
import { ThemeToggle } from "@/components/layout/theme-toggle";

export default function AuthLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="relative flex min-h-screen bg-background">
      <div className="absolute right-5 top-5 z-10 sm:right-8 sm:top-8">
        <ThemeToggle />
      </div>
      <AuthSidePanel />

      <div className="flex flex-1 items-center justify-center p-5 sm:p-8 lg:p-12">
        <div className="w-full max-w-[440px]">
          {children}
        </div>
      </div>
    </div>
  );
}
