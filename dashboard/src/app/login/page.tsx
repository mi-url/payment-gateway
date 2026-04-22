"use client";

import { useState } from "react";
import { Building2, Mail, Lock, ArrowRight, Shield } from "lucide-react";
import { createClient } from "@/lib/supabase/client";

export default function LoginPage() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError("");

    const supabase = createClient();
    const { error: authError } = await supabase.auth.signInWithPassword({
      email,
      password,
    });

    if (authError) {
      setError(authError.message);
      setLoading(false);
      return;
    }

    // Successful login — redirect to dashboard.
    // The middleware will handle session persistence.
    window.location.href = "/dashboard";
  };

  return (
    <div
      className="min-h-screen flex items-center justify-center px-4"
      style={{
        background:
          "radial-gradient(ellipse at 50% 0%, rgba(99,102,241,0.08) 0%, var(--bg-primary) 60%)",
      }}
    >
      <div className="w-full max-w-md">
        {/* Logo */}
        <div className="flex flex-col items-center mb-8">
          <div
            className="w-14 h-14 rounded-2xl flex items-center justify-center mb-4"
            style={{
              background: "linear-gradient(135deg, #6366f1, #818cf8)",
              boxShadow: "0 0 30px rgba(99,102,241,0.3)",
            }}
          >
            <Building2 size={28} color="white" />
          </div>
          <h1 className="text-2xl font-bold text-[var(--text-primary)]">
            Payment Gateway
          </h1>
          <p className="text-sm text-[var(--text-secondary)] mt-1">
            Sign in to your merchant dashboard
          </p>
        </div>

        {/* Login Card */}
        <div
          className="rounded-2xl border border-[var(--border)] p-8"
          style={{
            background: "var(--bg-card)",
            boxShadow: "0 4px 24px rgba(0,0,0,0.2)",
          }}
        >
          <form onSubmit={handleLogin} className="space-y-5">
            {/* Email */}
            <div>
              <label
                htmlFor="login-email"
                className="block text-sm font-medium text-[var(--text-primary)] mb-2"
              >
                Email
              </label>
              <div className="relative">
                <Mail
                  size={16}
                  className="absolute left-3 top-1/2 -translate-y-1/2 text-[var(--text-muted)]"
                />
                <input
                  id="login-email"
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  placeholder="merchant@company.com"
                  required
                  className="w-full pl-10 pr-4 py-3 rounded-xl text-sm border border-[var(--border)] placeholder:text-[var(--text-muted)] focus:outline-none focus:border-[var(--accent)] transition-colors cursor-text"
                  style={{
                    background: "var(--bg-primary)",
                    color: "var(--text-primary)",
                  }}
                />
              </div>
            </div>

            {/* Password */}
            <div>
              <div className="flex items-center justify-between mb-2">
                <label
                  htmlFor="login-password"
                  className="block text-sm font-medium text-[var(--text-primary)]"
                >
                  Password
                </label>
                <button
                  type="button"
                  className="text-xs text-[var(--accent)] hover:text-[var(--accent-hover)] transition-colors cursor-pointer"
                >
                  Forgot password?
                </button>
              </div>
              <div className="relative">
                <Lock
                  size={16}
                  className="absolute left-3 top-1/2 -translate-y-1/2 text-[var(--text-muted)]"
                />
                <input
                  id="login-password"
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  placeholder="••••••••••"
                  required
                  className="w-full pl-10 pr-4 py-3 rounded-xl text-sm border border-[var(--border)] placeholder:text-[var(--text-muted)] focus:outline-none focus:border-[var(--accent)] transition-colors cursor-text"
                  style={{
                    background: "var(--bg-primary)",
                    color: "var(--text-primary)",
                  }}
                />
              </div>
            </div>

            {/* Error */}
            {error && (
              <div
                className="px-4 py-3 rounded-xl text-sm"
                style={{
                  background: "rgba(239,68,68,0.1)",
                  color: "var(--danger)",
                  border: "1px solid rgba(239,68,68,0.2)",
                }}
              >
                {error}
              </div>
            )}

            {/* Submit */}
            <button
              type="submit"
              disabled={loading}
              className="w-full flex items-center justify-center gap-2 py-3 rounded-xl text-sm font-semibold text-white transition-all duration-200 cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
              style={{
                background: loading
                  ? "var(--accent)"
                  : "linear-gradient(135deg, #6366f1, #818cf8)",
                boxShadow: loading
                  ? "none"
                  : "0 2px 12px rgba(99,102,241,0.3)",
              }}
            >
              {loading ? (
                <div className="w-5 h-5 border-2 border-white/30 border-t-white rounded-full animate-spin" />
              ) : (
                <>
                  Sign In
                  <ArrowRight size={16} />
                </>
              )}
            </button>
          </form>
        </div>

        {/* Security badge */}
        <div className="flex items-center justify-center gap-2 mt-6">
          <Shield size={14} className="text-[var(--success)]" />
          <span className="text-xs text-[var(--text-muted)]">
            AES-256 encrypted · SOC 2 compliant · Zero-Trust architecture
          </span>
        </div>
      </div>
    </div>
  );
}
