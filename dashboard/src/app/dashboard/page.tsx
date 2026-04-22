"use client";

import { useEffect, useState } from "react";
import {
  ArrowUpRight,
  ArrowDownRight,
  DollarSign,
  Activity,
  CheckCircle2,
  XCircle,
  Clock,
} from "lucide-react";
import { createClient } from "@/lib/supabase";

interface Transaction {
  id: string;
  amount: number;
  status: string;
  bank_code: string;
  payer_phone: string;
  initiated_at: string;
  idempotency_key: string;
}

const statusConfig = {
  SUCCESS: { color: "var(--success)", bg: "rgba(16,185,129,0.1)", icon: CheckCircle2 },
  DECLINED: { color: "var(--danger)", bg: "rgba(239,68,68,0.1)", icon: XCircle },
  PENDING_RECONCILIATION: { color: "var(--pending)", bg: "rgba(59,130,246,0.1)", icon: Clock },
  PROCESSING: { color: "var(--warning)", bg: "rgba(245,158,11,0.1)", icon: Activity },
  INITIATED: { color: "var(--text-muted)", bg: "rgba(148,163,184,0.1)", icon: Clock },
  BANK_NETWORK_ERROR: { color: "var(--danger)", bg: "rgba(239,68,68,0.1)", icon: XCircle },
};

const bankNames: Record<string, string> = {
  "0191": "BNC",
  "0105": "Mercantil",
  "0134": "Banesco",
  "0102": "Venezuela",
  "0172": "Bancamiga",
};

function timeAgo(dateStr: string): string {
  const now = new Date();
  const date = new Date(dateStr);
  const seconds = Math.floor((now.getTime() - date.getTime()) / 1000);
  if (seconds < 60) return `${seconds}s ago`;
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes} min ago`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h ago`;
  return `${Math.floor(hours / 24)}d ago`;
}

function formatBs(amount: number): string {
  return `Bs. ${amount.toLocaleString("es-VE", { minimumFractionDigits: 2 })}`;
}

export default function DashboardPage() {
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function fetchData() {
      const supabase = createClient();
      const { data, error } = await supabase
        .from("transactions")
        .select("id, amount, status, bank_code, payer_phone, initiated_at, idempotency_key")
        .order("initiated_at", { ascending: false })
        .limit(10);

      if (!error && data) {
        setTransactions(data);
      }
      setLoading(false);
    }
    fetchData();
  }, []);

  // Compute stats from real data
  const totalRevenue = transactions
    .filter((t) => t.status === "SUCCESS")
    .reduce((sum, t) => sum + t.amount, 0);
  const totalCount = transactions.length;
  const successCount = transactions.filter((t) => t.status === "SUCCESS").length;
  const successRate = totalCount > 0 ? ((successCount / totalCount) * 100).toFixed(1) : "0.0";
  const pendingCount = transactions.filter(
    (t) => t.status === "PENDING_RECONCILIATION" || t.status === "PROCESSING"
  ).length;

  const stats = [
    {
      label: "Total Revenue",
      value: formatBs(totalRevenue),
      change: successCount > 0 ? `${successCount} txns` : "—",
      trend: "up" as const,
      icon: DollarSign,
    },
    {
      label: "Transactions",
      value: String(totalCount),
      change: "all time",
      trend: "up" as const,
      icon: Activity,
    },
    {
      label: "Success Rate",
      value: `${successRate}%`,
      change: `${successCount}/${totalCount}`,
      trend: Number(successRate) >= 90 ? ("up" as const) : ("down" as const),
      icon: CheckCircle2,
    },
    {
      label: "Pending",
      value: String(pendingCount),
      change: pendingCount === 0 ? "all clear" : "needs review",
      trend: pendingCount === 0 ? ("up" as const) : ("down" as const),
      icon: Clock,
    },
  ];

  return (
    <div className="max-w-7xl">
      {/* Header */}
      <div className="mb-8">
        <h1 className="text-2xl font-bold text-[var(--text-primary)]">Overview</h1>
        <p className="text-sm text-[var(--text-secondary)] mt-1">
          {loading ? "Loading data from Supabase..." : "Live data from your payment gateway."}
        </p>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
        {stats.map((stat) => (
          <div
            key={stat.label}
            className="rounded-xl p-5 border border-[var(--border)] transition-all duration-200 hover:border-[var(--accent)] cursor-pointer"
            style={{ background: "var(--bg-card)" }}
          >
            <div className="flex items-center justify-between mb-3">
              <div
                className="w-10 h-10 rounded-lg flex items-center justify-center"
                style={{ background: "var(--bg-hover)" }}
              >
                <stat.icon size={18} className="text-[var(--accent)]" />
              </div>
              <div
                className={`flex items-center gap-1 text-xs font-medium px-2 py-1 rounded-full ${
                  stat.trend === "up" ? "text-emerald-400" : "text-red-400"
                }`}
                style={{
                  background:
                    stat.trend === "up" ? "rgba(16,185,129,0.1)" : "rgba(239,68,68,0.1)",
                }}
              >
                {stat.trend === "up" ? (
                  <ArrowUpRight size={12} />
                ) : (
                  <ArrowDownRight size={12} />
                )}
                {stat.change}
              </div>
            </div>
            <p className="text-2xl font-bold text-[var(--text-primary)]">{stat.value}</p>
            <p className="text-xs text-[var(--text-muted)] mt-1">{stat.label}</p>
          </div>
        ))}
      </div>

      {/* Recent Transactions */}
      <div
        className="rounded-xl border border-[var(--border)] overflow-hidden"
        style={{ background: "var(--bg-card)" }}
      >
        <div className="px-6 py-4 border-b border-[var(--border)] flex items-center justify-between">
          <h2 className="text-sm font-semibold text-[var(--text-primary)]">
            Recent Transactions
          </h2>
          <a
            href="/dashboard/transactions"
            className="text-xs text-[var(--accent)] hover:text-[var(--accent-hover)] transition-colors cursor-pointer"
          >
            View all →
          </a>
        </div>
        <div className="divide-y divide-[var(--border)]">
          {loading ? (
            <div className="px-6 py-8 text-center">
              <div className="w-6 h-6 border-2 border-[var(--accent)]/30 border-t-[var(--accent)] rounded-full animate-spin mx-auto mb-2" />
              <p className="text-sm text-[var(--text-muted)]">Loading transactions...</p>
            </div>
          ) : transactions.length === 0 ? (
            <div className="px-6 py-8 text-center">
              <p className="text-sm text-[var(--text-muted)]">No transactions yet.</p>
            </div>
          ) : (
            transactions.map((txn) => {
              const config = statusConfig[txn.status as keyof typeof statusConfig];
              const StatusIcon = config?.icon || Activity;
              return (
                <div
                  key={txn.id}
                  className="px-6 py-3.5 flex items-center justify-between hover:bg-[var(--bg-hover)] transition-colors cursor-pointer"
                >
                  <div className="flex items-center gap-4">
                    <div
                      className="w-8 h-8 rounded-full flex items-center justify-center"
                      style={{ background: config?.bg }}
                    >
                      <StatusIcon size={14} style={{ color: config?.color }} />
                    </div>
                    <div>
                      <p className="text-sm font-medium text-[var(--text-primary)]">
                        {formatBs(txn.amount)}
                      </p>
                      <p className="text-xs text-[var(--text-muted)]">
                        {txn.idempotency_key} · {bankNames[txn.bank_code] || txn.bank_code}
                      </p>
                    </div>
                  </div>
                  <div className="text-right">
                    <span
                      className="text-xs font-medium px-2 py-1 rounded-full"
                      style={{ color: config?.color, background: config?.bg }}
                    >
                      {txn.status.replace(/_/g, " ")}
                    </span>
                    <p className="text-xs text-[var(--text-muted)] mt-1">
                      {timeAgo(txn.initiated_at)}
                    </p>
                  </div>
                </div>
              );
            })
          )}
        </div>
      </div>
    </div>
  );
}
