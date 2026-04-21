"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  LayoutDashboard,
  ArrowLeftRight,
  Settings,
  KeyRound,
  Building2,
  ChevronLeft,
} from "lucide-react";
import { useState } from "react";

const navItems = [
  { href: "/dashboard", label: "Overview", icon: LayoutDashboard },
  { href: "/dashboard/transactions", label: "Transactions", icon: ArrowLeftRight },
  { href: "/dashboard/settings", label: "Settings", icon: Settings },
];

export function Sidebar() {
  const pathname = usePathname();
  const [collapsed, setCollapsed] = useState(false);

  return (
    <aside
      className={`fixed left-0 top-0 h-screen flex flex-col transition-all duration-300 ease-in-out z-50 ${
        collapsed ? "w-[72px]" : "w-[260px]"
      }`}
      style={{
        background: "var(--bg-secondary)",
        borderRight: "1px solid var(--border)",
      }}
    >
      {/* Logo */}
      <div className="flex items-center gap-3 px-5 h-16 border-b border-[var(--border)]">
        <div
          className="w-8 h-8 rounded-lg flex items-center justify-center flex-shrink-0"
          style={{ background: "var(--accent)" }}
        >
          <Building2 size={18} color="white" />
        </div>
        {!collapsed && (
          <div className="overflow-hidden">
            <p className="text-sm font-semibold text-[var(--text-primary)] truncate">
              Payment Gateway
            </p>
            <p className="text-xs text-[var(--text-muted)]">Dashboard</p>
          </div>
        )}
      </div>

      {/* Navigation */}
      <nav className="flex-1 py-4 px-3 space-y-1">
        {navItems.map((item) => {
          const isActive = pathname === item.href;
          return (
            <Link
              key={item.href}
              href={item.href}
              className={`flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-all duration-200 ${
                isActive
                  ? "text-white"
                  : "text-[var(--text-secondary)] hover:text-[var(--text-primary)] hover:bg-[var(--bg-hover)]"
              }`}
              style={isActive ? { background: "var(--accent)" } : undefined}
            >
              <item.icon size={18} className="flex-shrink-0" />
              {!collapsed && <span>{item.label}</span>}
            </Link>
          );
        })}
      </nav>

      {/* API Key hint */}
      {!collapsed && (
        <div className="mx-3 mb-3 p-3 rounded-lg" style={{ background: "var(--bg-card)" }}>
          <div className="flex items-center gap-2 mb-1">
            <KeyRound size={14} className="text-[var(--accent)]" />
            <span className="text-xs font-medium text-[var(--text-primary)]">API Key</span>
          </div>
          <p className="text-xs text-[var(--text-muted)] leading-relaxed">
            Use your API key to integrate payments into your app.
          </p>
        </div>
      )}

      {/* Collapse toggle */}
      <button
        onClick={() => setCollapsed(!collapsed)}
        className="flex items-center justify-center h-12 border-t border-[var(--border)] text-[var(--text-muted)] hover:text-[var(--text-primary)] transition-colors"
      >
        <ChevronLeft
          size={18}
          className={`transition-transform duration-300 ${collapsed ? "rotate-180" : ""}`}
        />
      </button>
    </aside>
  );
}
