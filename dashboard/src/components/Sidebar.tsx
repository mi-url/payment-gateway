"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  LayoutDashboard,
  ArrowLeftRight,
  Settings,
  KeyRound,
  Building2,
  LogOut,
  User,
  ChevronsUpDown,
  Terminal,
} from "lucide-react";
import { useState, useEffect } from "react";
import { createClient } from "@/lib/supabase/client";

import {
  Sidebar as ShadcnSidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar,
} from "@/components/ui/sidebar";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";

const navItems = [
  { href: "/dashboard", label: "Overview", icon: LayoutDashboard },
  { href: "/dashboard/transactions", label: "Transactions", icon: ArrowLeftRight },
  { href: "/dashboard/developers", label: "Developers", icon: Terminal },
  { href: "/dashboard/settings", label: "Settings", icon: Settings },
];

export function Sidebar() {
  const pathname = usePathname();
  const { isMobile } = useSidebar();
  const [email, setEmail] = useState<string>("merchant@example.com");

  useEffect(() => {
    async function getUser() {
      const supabase = createClient();
      const { data: { user } } = await supabase.auth.getUser();
      if (user?.email) {
        setEmail(user.email);
      }
    }
    getUser();
  }, []);

  const handleLogout = async () => {
    const supabase = createClient();
    await supabase.auth.signOut();
    window.location.href = "/login";
  };

  return (
    <ShadcnSidebar variant="sidebar" collapsible="icon">
      <SidebarHeader className="border-b border-border/50 py-4 h-16">
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton size="lg" render={<Link href="/dashboard" />} className="hover:bg-transparent">
              <div className="flex aspect-square size-8 items-center justify-center rounded-lg bg-primary text-primary-foreground">
                <Building2 className="size-4" />
              </div>
              <div className="flex flex-col gap-0.5 leading-none">
                <span className="font-semibold text-foreground">Payment Gateway</span>
                <span className="text-[10px] text-muted-foreground uppercase tracking-wider font-medium">Dashboard</span>
              </div>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>

      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupContent>
            <SidebarMenu className="mt-4 gap-2">
              {navItems.map((item) => {
                const isActive = pathname === item.href;
                return (
                  <SidebarMenuItem key={item.href}>
                    <SidebarMenuButton
                      render={<Link href={item.href} />}
                      isActive={isActive}
                      tooltip={item.label}
                      className={
                        isActive 
                          ? "bg-primary/10 text-primary hover:bg-primary/15 hover:text-primary" 
                          : "text-muted-foreground hover:bg-secondary hover:text-foreground"
                      }
                    >
                      <item.icon className="size-4" />
                      <span className="font-medium">{item.label}</span>
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                );
              })}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>

      <SidebarFooter className="border-t border-border/50 pb-4">
        {/* Help / Hint Card - only visible when expanded */}
        <SidebarGroup className="group-data-[collapsible=icon]:hidden">
          <div className="mx-2 mb-2 p-3 rounded-xl bg-secondary/50 border border-border/50">
            <div className="flex items-center gap-2 mb-1.5">
              <div className="p-1 rounded-md bg-primary/10">
                <KeyRound className="size-3.5 text-primary" />
              </div>
              <span className="text-[11px] font-semibold text-foreground uppercase tracking-wider">Live API Key</span>
            </div>
            <p className="text-[11px] text-muted-foreground leading-relaxed">
              Use your API key to process production payments.
            </p>
          </div>
        </SidebarGroup>

        <SidebarMenu>
          <SidebarMenuItem>
            <DropdownMenu>
              <DropdownMenuTrigger render={
                <SidebarMenuButton
                  size="lg"
                  className="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
                >
                  <Avatar className="h-8 w-8 rounded-lg">
                    <AvatarFallback className="rounded-lg bg-primary/10 text-primary">
                      {email.substring(0, 2).toUpperCase()}
                    </AvatarFallback>
                  </Avatar>
                  <div className="grid flex-1 text-left text-sm leading-tight">
                    <span className="truncate font-semibold text-foreground">Merchant</span>
                    <span className="truncate text-xs text-muted-foreground">{email}</span>
                  </div>
                  <ChevronsUpDown className="ml-auto size-4 text-muted-foreground" />
                </SidebarMenuButton>
              } />
              <DropdownMenuContent
                className="w-[--radix-dropdown-menu-trigger-width] min-w-56 rounded-lg bg-card border-border/50"
                side={isMobile ? "bottom" : "right"}
                align="end"
                sideOffset={4}
              >
                <div className="p-0 font-normal">
                  <div className="flex items-center gap-2 px-1 py-1.5 text-left text-sm">
                    <Avatar className="h-8 w-8 rounded-lg">
                      <AvatarFallback className="rounded-lg bg-primary/10 text-primary">
                        {email.substring(0, 2).toUpperCase()}
                      </AvatarFallback>
                    </Avatar>
                    <div className="grid flex-1 text-left text-sm leading-tight">
                      <span className="truncate font-semibold text-foreground">Merchant</span>
                      <span className="truncate text-xs text-muted-foreground">{email}</span>
                    </div>
                  </div>
                </div>
                <DropdownMenuSeparator className="bg-border/50" />
                <DropdownMenuItem render={<Link href="/dashboard/settings" />} className="cursor-pointer focus:bg-secondary">
                  <User className="mr-2 size-4 text-muted-foreground" />
                  <span>Account Settings</span>
                </DropdownMenuItem>
                <DropdownMenuSeparator className="bg-border/50" />
                <DropdownMenuItem onClick={handleLogout} className="cursor-pointer text-destructive focus:bg-destructive/10 focus:text-destructive">
                  <LogOut className="mr-2 size-4" />
                  <span>Log out</span>
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarFooter>
    </ShadcnSidebar>
  );
}
