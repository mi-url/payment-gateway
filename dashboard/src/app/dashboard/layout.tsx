import { Sidebar } from "@/components/Sidebar";
import { Toaster } from "@/components/ui/sonner";
import { SidebarProvider, SidebarTrigger } from "@/components/ui/sidebar";

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <SidebarProvider>
      <Sidebar />
      <main className="flex-1 w-full flex flex-col min-h-screen">
        <div className="flex h-14 items-center gap-2 border-b border-border/50 bg-background/95 px-4 backdrop-blur supports-[backdrop-filter]:bg-background/60 lg:hidden">
          <SidebarTrigger />
        </div>
        <div className="flex-1 p-6 md:p-8">
          {children}
        </div>
      </main>
      <Toaster position="top-right" theme="dark" />
    </SidebarProvider>
  );
}
