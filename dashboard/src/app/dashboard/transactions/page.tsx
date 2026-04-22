"use client";

import { useEffect, useState } from "react";
import {
  CheckCircle2,
  XCircle,
  Clock,
  Activity,
  Search,
  Filter,
} from "lucide-react";
import { createClient } from "@/lib/supabase";

interface Transaction {
  id: string;
  idempotency_key: string;
  amount: number;
  currency: string;
  status: string;
  bank_code: string;
  payer_phone: string | null;
  payer_id_document: string | null;
  bank_reference: string | null;
  error_code: string | null;
  initiated_at: string;
}

const statusConfig: Record<string, { color: string; bg: string; icon: React.ElementType }> = {
  SUCCESS: { color: "var(--success)", bg: "rgba(16,185,129,0.1)", icon: CheckCircle2 },
  DECLINED: { color: "var(--danger)", bg: "rgba(239,68,68,0.1)", icon: XCircle },
  PENDING_RECONCILIATION: { color: "var(--pending)", bg: "rgba(59,130,246,0.1)", icon: Clock },
  PROCESSING: { color: "var(--warning)", bg: "rgba(245,158,11,0.1)", icon: Activity },
  INITIATED: { color: "var(--text-muted)", bg: "rgba(148,163,184,0.1)", icon: Clock },
  BANK_NETWORK_ERROR: { color: "var(--danger)", bg: "rgba(239,68,68,0.1)", icon: XCircle },
};

const bankNames: Record<string, string> = {
  "0191": "BNC (0191)",
  "0105": "Mercantil (0105)",
  "0134": "Banesco (0134)",
  "0102": "Venezuela (0102)",
  "0172": "Bancamiga (0172)",
};

function formatAmount(amount: number): string {
  return `Bs. ${amount.toLocaleString("es-VE", { minimumFractionDigits: 2 })}`;
}

function formatDate(dateStr: string): string {
  const d = new Date(dateStr);
  return d.toLocaleString("es-VE", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  });
}

export default function TransactionsPage() {
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState("");

  useEffect(() => {
    async function fetchData() {
      const supabase = createClient();
      const { data, error } = await supabase
        .from("transactions")
        .select("*")
        .order("initiated_at", { ascending: false });

      if (!error && data) {
        setTransactions(data);
      }
      setLoading(false);
    }
    fetchData();
  }, []);

  // Client-side search filter
  const filtered = transactions.filter((txn) => {
    if (!searchQuery) return true;
    const q = searchQuery.toLowerCase();
    return (
      txn.idempotency_key.toLowerCase().includes(q) ||
      (txn.payer_phone && txn.payer_phone.includes(q)) ||
      (txn.bank_reference && txn.bank_reference.includes(q)) ||
      txn.status.toLowerCase().includes(q)
    );
  });

  return (
    <div className="max-w-7xl">
      {/* Header */}
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-2xl font-bold text-[var(--text-primary)]">Transactions</h1>
          <p className="text-sm text-[var(--text-secondary)] mt-1">
            {loading
              ? "Loading from Supabase..."
              : `${transactions.length} transactions · ${filtered.length} shown`}
          </p>
        </div>
        <button
          className="flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium text-white transition-colors cursor-pointer"
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
          id="transactions-search"
          type="text"
          placeholder="Search by transaction ID, phone, or reference..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="w-full pl-10 pr-4 py-2.5 rounded-lg text-sm border border-[var(--border)] placeholder:text-[var(--text-muted)] focus:outline-none focus:border-[var(--accent)] transition-colors cursor-text"
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
              {loading ? (
                <tr>
                  <td colSpan={6} className="px-6 py-8 text-center">
                    <div className="w-6 h-6 border-2 border-[var(--accent)]/30 border-t-[var(--accent)] rounded-full animate-spin mx-auto mb-2" />
                    <p className="text-sm text-[var(--text-muted)]">Loading transactions...</p>
                  </td>
                </tr>
              ) : filtered.length === 0 ? (
                <tr>
                  <td colSpan={6} className="px-6 py-8 text-center">
                    <p className="text-sm text-[var(--text-muted)]">
                      {searchQuery ? "No transactions match your search." : "No transactions yet."}
                    </p>
                  </td>
                </tr>
              ) : (
                filtered.map((txn) => {
                  const config = statusConfig[txn.status] || statusConfig.PROCESSING;
                  return (
                    <tr
                      key={txn.id}
                      className="hover:bg-[var(--bg-hover)] transition-colors cursor-pointer"
                    >
                      <td className="px-6 py-4">
                        <p className="text-sm font-mono text-[var(--text-primary)]">
                          {txn.idempotency_key}
                        </p>
                        <p className="text-xs text-[var(--text-muted)]">
                          {txn.payer_phone || "—"}
                        </p>
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
                        {txn.error_code && (
                          <p className="text-xs text-[var(--danger)] mt-0.5">{txn.error_code}</p>
                        )}
                      </td>
                      <td className="px-6 py-4">
                        <p className="text-sm text-[var(--text-secondary)]">
                          {bankNames[txn.bank_code] || txn.bank_code}
                        </p>
                      </td>
                      <td className="px-6 py-4">
                        <p className="text-sm font-mono text-[var(--text-secondary)]">
                          {txn.bank_reference || "—"}
                        </p>
                      </td>
                      <td className="px-6 py-4">
                        <p className="text-sm text-[var(--text-secondary)]">
                          {formatDate(txn.initiated_at)}
                        </p>
                      </td>
                    </tr>
                  );
                })
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
