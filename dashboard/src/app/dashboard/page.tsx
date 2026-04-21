import {
  ArrowUpRight,
  ArrowDownRight,
  DollarSign,
  Activity,
  CheckCircle2,
  XCircle,
  Clock,
} from "lucide-react";

// Mock data for the overview dashboard.
// In production, this comes from the Gateway API.
const stats = [
  {
    label: "Total Revenue",
    value: "Bs. 125,430.00",
    change: "+12.5%",
    trend: "up" as const,
    icon: DollarSign,
  },
  {
    label: "Transactions Today",
    value: "47",
    change: "+8.2%",
    trend: "up" as const,
    icon: Activity,
  },
  {
    label: "Success Rate",
    value: "96.8%",
    change: "+0.3%",
    trend: "up" as const,
    icon: CheckCircle2,
  },
  {
    label: "Pending",
    value: "3",
    change: "-2",
    trend: "down" as const,
    icon: Clock,
  },
];

const recentTransactions = [
  { id: "txn_a1b2c3", amount: "Bs. 2,500.00", status: "SUCCESS", bank: "BNC", time: "2 min ago" },
  { id: "txn_d4e5f6", amount: "Bs. 800.00", status: "SUCCESS", bank: "BNC", time: "5 min ago" },
  { id: "txn_g7h8i9", amount: "Bs. 15,000.00", status: "DECLINED", bank: "BNC", time: "8 min ago" },
  { id: "txn_j0k1l2", amount: "Bs. 3,200.00", status: "SUCCESS", bank: "BNC", time: "12 min ago" },
  { id: "txn_m3n4o5", amount: "Bs. 450.00", status: "PENDING_RECONCILIATION", bank: "BNC", time: "15 min ago" },
  { id: "txn_p6q7r8", amount: "Bs. 7,800.00", status: "SUCCESS", bank: "BNC", time: "22 min ago" },
];

const statusConfig = {
  SUCCESS: { color: "var(--success)", bg: "rgba(16,185,129,0.1)", icon: CheckCircle2 },
  DECLINED: { color: "var(--danger)", bg: "rgba(239,68,68,0.1)", icon: XCircle },
  PENDING_RECONCILIATION: { color: "var(--pending)", bg: "rgba(59,130,246,0.1)", icon: Clock },
  PROCESSING: { color: "var(--warning)", bg: "rgba(245,158,11,0.1)", icon: Activity },
};

export default function DashboardPage() {
  return (
    <div className="max-w-7xl">
      {/* Header */}
      <div className="mb-8">
        <h1 className="text-2xl font-bold text-[var(--text-primary)]">Overview</h1>
        <p className="text-sm text-[var(--text-secondary)] mt-1">
          Monitor your payment gateway performance in real-time.
        </p>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
        {stats.map((stat) => (
          <div
            key={stat.label}
            className="rounded-xl p-5 border border-[var(--border)] transition-all duration-200 hover:border-[var(--accent)]"
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
            className="text-xs text-[var(--accent)] hover:text-[var(--accent-hover)] transition-colors"
          >
            View all →
          </a>
        </div>
        <div className="divide-y divide-[var(--border)]">
          {recentTransactions.map((txn) => {
            const config = statusConfig[txn.status as keyof typeof statusConfig];
            const StatusIcon = config?.icon || Activity;
            return (
              <div
                key={txn.id}
                className="px-6 py-3.5 flex items-center justify-between hover:bg-[var(--bg-hover)] transition-colors"
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
                      {txn.amount}
                    </p>
                    <p className="text-xs text-[var(--text-muted)]">
                      {txn.id} · {txn.bank}
                    </p>
                  </div>
                </div>
                <div className="text-right">
                  <span
                    className="text-xs font-medium px-2 py-1 rounded-full"
                    style={{ color: config?.color, background: config?.bg }}
                  >
                    {txn.status.replace("_", " ")}
                  </span>
                  <p className="text-xs text-[var(--text-muted)] mt-1">{txn.time}</p>
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}
