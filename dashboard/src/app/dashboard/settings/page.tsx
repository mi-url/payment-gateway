import { KeyRound, Building2, Copy, Eye, EyeOff, Shield } from "lucide-react";

export default function SettingsPage() {
  return (
    <div className="max-w-4xl">
      {/* Header */}
      <div className="mb-8">
        <h1 className="text-2xl font-bold text-[var(--text-primary)]">Settings</h1>
        <p className="text-sm text-[var(--text-secondary)] mt-1">
          Manage your API keys, bank configurations, and webhook settings.
        </p>
      </div>

      {/* API Keys Section */}
      <section className="mb-8">
        <h2 className="text-sm font-semibold text-[var(--text-primary)] uppercase tracking-wider mb-4 flex items-center gap-2">
          <KeyRound size={16} className="text-[var(--accent)]" />
          API Keys
        </h2>
        <div
          className="rounded-xl border border-[var(--border)] p-6"
          style={{ background: "var(--bg-card)" }}
        >
          <div className="flex items-center justify-between mb-4">
            <div>
              <p className="text-sm font-medium text-[var(--text-primary)]">Live API Key</p>
              <p className="text-xs text-[var(--text-muted)] mt-0.5">
                Use this key for production transactions
              </p>
            </div>
            <button
              className="px-3 py-1.5 rounded-lg text-xs font-medium text-white transition-colors"
              style={{ background: "var(--accent)" }}
            >
              Regenerate
            </button>
          </div>
          <div
            className="flex items-center justify-between px-4 py-3 rounded-lg font-mono text-sm"
            style={{ background: "var(--bg-primary)" }}
          >
            <span className="text-[var(--text-secondary)]">
              gw_live_••••••••••••••••••••••••
            </span>
            <div className="flex items-center gap-2">
              <button className="text-[var(--text-muted)] hover:text-[var(--text-primary)] transition-colors">
                <Eye size={14} />
              </button>
              <button className="text-[var(--text-muted)] hover:text-[var(--text-primary)] transition-colors">
                <Copy size={14} />
              </button>
            </div>
          </div>
          <div className="flex items-center gap-2 mt-3">
            <Shield size={12} className="text-[var(--success)]" />
            <span className="text-xs text-[var(--text-muted)]">
              Keys are hashed with SHA-256 — we never store the raw key
            </span>
          </div>
        </div>
      </section>

      {/* Bank Configurations */}
      <section className="mb-8">
        <h2 className="text-sm font-semibold text-[var(--text-primary)] uppercase tracking-wider mb-4 flex items-center gap-2">
          <Building2 size={16} className="text-[var(--accent)]" />
          Bank Configurations
        </h2>
        <div
          className="rounded-xl border border-[var(--border)] overflow-hidden"
          style={{ background: "var(--bg-card)" }}
        >
          {/* BNC */}
          <div className="p-6 border-b border-[var(--border)] flex items-center justify-between">
            <div className="flex items-center gap-4">
              <div
                className="w-10 h-10 rounded-lg flex items-center justify-center text-sm font-bold"
                style={{ background: "rgba(99,102,241,0.1)", color: "var(--accent)" }}
              >
                BNC
              </div>
              <div>
                <p className="text-sm font-medium text-[var(--text-primary)]">
                  Banco Nacional de Crédito
                </p>
                <p className="text-xs text-[var(--text-muted)]">Code: 0191 · C2P Active</p>
              </div>
            </div>
            <div className="flex items-center gap-3">
              <span className="text-xs font-medium px-2.5 py-1 rounded-full text-emerald-400"
                style={{ background: "rgba(16,185,129,0.1)" }}
              >
                Connected
              </span>
              <button className="text-xs text-[var(--text-secondary)] hover:text-[var(--text-primary)] transition-colors">
                Configure →
              </button>
            </div>
          </div>

          {/* Mercantil (locked) */}
          <div className="p-6 border-b border-[var(--border)] flex items-center justify-between opacity-50">
            <div className="flex items-center gap-4">
              <div
                className="w-10 h-10 rounded-lg flex items-center justify-center text-sm font-bold"
                style={{ background: "var(--bg-hover)", color: "var(--text-muted)" }}
              >
                MER
              </div>
              <div>
                <p className="text-sm font-medium text-[var(--text-primary)]">Mercantil</p>
                <p className="text-xs text-[var(--text-muted)]">Code: 0105 · Coming soon</p>
              </div>
            </div>
            <span className="text-xs text-[var(--text-muted)]">Pending integration</span>
          </div>

          {/* Banesco (locked) */}
          <div className="p-6 flex items-center justify-between opacity-50">
            <div className="flex items-center gap-4">
              <div
                className="w-10 h-10 rounded-lg flex items-center justify-center text-sm font-bold"
                style={{ background: "var(--bg-hover)", color: "var(--text-muted)" }}
              >
                BAN
              </div>
              <div>
                <p className="text-sm font-medium text-[var(--text-primary)]">Banesco</p>
                <p className="text-xs text-[var(--text-muted)]">Code: 0134 · Coming soon</p>
              </div>
            </div>
            <span className="text-xs text-[var(--text-muted)]">Pending integration</span>
          </div>
        </div>
      </section>

      {/* Webhook Configuration */}
      <section>
        <h2 className="text-sm font-semibold text-[var(--text-primary)] uppercase tracking-wider mb-4 flex items-center gap-2">
          <Shield size={16} className="text-[var(--accent)]" />
          Webhook
        </h2>
        <div
          className="rounded-xl border border-[var(--border)] p-6"
          style={{ background: "var(--bg-card)" }}
        >
          <label className="block text-sm font-medium text-[var(--text-primary)] mb-2">
            Webhook URL
          </label>
          <input
            type="url"
            placeholder="https://your-app.com/webhooks/payments"
            className="w-full px-4 py-2.5 rounded-lg text-sm border border-[var(--border)] placeholder:text-[var(--text-muted)] focus:outline-none focus:border-[var(--accent)] transition-colors"
            style={{ background: "var(--bg-primary)", color: "var(--text-primary)" }}
          />
          <p className="text-xs text-[var(--text-muted)] mt-2">
            We&apos;ll send POST requests to this URL when transaction statuses change.
          </p>
          <button
            className="mt-4 px-4 py-2 rounded-lg text-sm font-medium text-white transition-colors"
            style={{ background: "var(--accent)" }}
          >
            Save Webhook
          </button>
        </div>
      </section>
    </div>
  );
}
