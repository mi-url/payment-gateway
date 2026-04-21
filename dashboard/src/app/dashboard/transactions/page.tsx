import {
  CheckCircle2,
  XCircle,
  Clock,
  Activity,
  Search,
  Filter,
} from "lucide-react";

// Mock transaction data — in production, fetched from Gateway API.
const transactions = [
  { id: "txn_a1b2c3d4", amount: 2500.0, currency: "VES", status: "SUCCESS", bank: "BNC (0191)", payerPhone: "0414***4567", reference: "123456789012", time: "2026-04-21 08:30:12" },
  { id: "txn_e5f6g7h8", amount: 800.0, currency: "VES", status: "SUCCESS", bank: "BNC (0191)", payerPhone: "0412***8901", reference: "234567890123", time: "2026-04-21 08:25:44" },
  { id: "txn_i9j0k1l2", amount: 15000.0, currency: "VES", status: "DECLINED", bank: "BNC (0191)", payerPhone: "0416***2345", reference: "—", time: "2026-04-21 08:20:33", error: "INCORRECT_OTP" },
  { id: "txn_m3n4o5p6", amount: 3200.0, currency: "VES", status: "SUCCESS", bank: "BNC (0191)", payerPhone: "0424***6789", reference: "345678901234", time: "2026-04-21 08:15:17" },
  { id: "txn_q7r8s9t0", amount: 450.0, currency: "VES", status: "PENDING_RECONCILIATION", bank: "BNC (0191)", payerPhone: "0414***0123", reference: "—", time: "2026-04-21 08:10:05" },
  { id: "txn_u1v2w3x4", amount: 7800.0, currency: "VES", status: "SUCCESS", bank: "BNC (0191)", payerPhone: "0412***4567", reference: "456789012345", time: "2026-04-21 07:58:22" },
  { id: "txn_y5z6a7b8", amount: 1250.0, currency: "VES", status: "SUCCESS", bank: "BNC (0191)", payerPhone: "0416***8901", reference: "567890123456", time: "2026-04-21 07:45:11" },
  { id: "txn_c9d0e1f2", amount: 22000.0, currency: "VES", status: "DECLINED", bank: "BNC (0191)", payerPhone: "0424***2345", reference: "—", time: "2026-04-21 07:30:59", error: "INSUFFICIENT_FUNDS" },
];

const statusConfig: Record<string, { color: string; bg: string; icon: React.ElementType }> = {
  SUCCESS: { color: "var(--success)", bg: "rgba(16,185,129,0.1)", icon: CheckCircle2 },
  DECLINED: { color: "var(--danger)", bg: "rgba(239,68,68,0.1)", icon: XCircle },
  PENDING_RECONCILIATION: { color: "var(--pending)", bg: "rgba(59,130,246,0.1)", icon: Clock },
  PROCESSING: { color: "var(--warning)", bg: "rgba(245,158,11,0.1)", icon: Activity },
};

function formatAmount(amount: number): string {
  return `Bs. ${amount.toLocaleString("es-VE", { minimumFractionDigits: 2 })}`;
}

export default function TransactionsPage() {
  return (
    <div className="max-w-7xl">
      {/* Header */}
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-2xl font-bold text-[var(--text-primary)]">Transactions</h1>
          <p className="text-sm text-[var(--text-secondary)] mt-1">
            View and filter all payment transactions.
          </p>
        </div>
        <button
          className="flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium text-white transition-colors"
          style={{ background: "var(--accent)" }}
        >
          <Filter size={14} />
          Filter
        </button>
      </div>

      {/* Search bar */}
      <div className="relative mb-6">
        <Search
          size={16}
          className="absolute left-3 top-1/2 -translate-y-1/2 text-[var(--text-muted)]"
        />
        <input
          type="text"
          placeholder="Search by transaction ID, phone, or reference..."
          className="w-full pl-10 pr-4 py-2.5 rounded-lg text-sm border border-[var(--border)] placeholder:text-[var(--text-muted)] focus:outline-none focus:border-[var(--accent)] transition-colors"
          style={{ background: "var(--bg-card)", color: "var(--text-primary)" }}
        />
      </div>

      {/* Table */}
      <div
        className="rounded-xl border border-[var(--border)] overflow-hidden"
        style={{ background: "var(--bg-card)" }}
      >
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b border-[var(--border)]">
                <th className="text-left text-xs font-medium text-[var(--text-muted)] px-6 py-3 uppercase tracking-wider">
                  Transaction
                </th>
                <th className="text-left text-xs font-medium text-[var(--text-muted)] px-6 py-3 uppercase tracking-wider">
                  Amount
                </th>
                <th className="text-left text-xs font-medium text-[var(--text-muted)] px-6 py-3 uppercase tracking-wider">
                  Status
                </th>
                <th className="text-left text-xs font-medium text-[var(--text-muted)] px-6 py-3 uppercase tracking-wider">
                  Bank
                </th>
                <th className="text-left text-xs font-medium text-[var(--text-muted)] px-6 py-3 uppercase tracking-wider">
                  Reference
                </th>
                <th className="text-left text-xs font-medium text-[var(--text-muted)] px-6 py-3 uppercase tracking-wider">
                  Date
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-[var(--border)]">
              {transactions.map((txn) => {
                const config = statusConfig[txn.status] || statusConfig.PROCESSING;
                return (
                  <tr
                    key={txn.id}
                    className="hover:bg-[var(--bg-hover)] transition-colors cursor-pointer"
                  >
                    <td className="px-6 py-4">
                      <p className="text-sm font-mono text-[var(--text-primary)]">{txn.id}</p>
                      <p className="text-xs text-[var(--text-muted)]">{txn.payerPhone}</p>
                    </td>
                    <td className="px-6 py-4">
                      <p className="text-sm font-semibold text-[var(--text-primary)]">
                        {formatAmount(txn.amount)}
                      </p>
                    </td>
                    <td className="px-6 py-4">
                      <span
                        className="inline-flex items-center gap-1.5 text-xs font-medium px-2.5 py-1 rounded-full"
                        style={{ color: config.color, background: config.bg }}
                      >
                        <config.icon size={12} />
                        {txn.status.replace(/_/g, " ")}
                      </span>
                      {txn.error && (
                        <p className="text-xs text-[var(--danger)] mt-0.5">{txn.error}</p>
                      )}
                    </td>
                    <td className="px-6 py-4">
                      <p className="text-sm text-[var(--text-secondary)]">{txn.bank}</p>
                    </td>
                    <td className="px-6 py-4">
                      <p className="text-sm font-mono text-[var(--text-secondary)]">
                        {txn.reference}
                      </p>
                    </td>
                    <td className="px-6 py-4">
                      <p className="text-sm text-[var(--text-secondary)]">{txn.time}</p>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
