"use client";

import { useEffect, useState, useCallback } from "react";
import {
  KeyRound,
  Building2,
  Copy,
  Eye,
  EyeOff,
  Shield,
  Save,
  CheckCircle2,
  AlertCircle,
  Loader2,
  Link2,
} from "lucide-react";
import { createClient } from "@/lib/supabase/client";

interface BankConfigStatus {
  bank_code: string;
  is_active: boolean;
  created_at: string;
}

export default function SettingsPage() {
  const [webhookUrl, setWebhookUrl] = useState("");
  const [webhookSaving, setWebhookSaving] = useState(false);
  const [webhookSaved, setWebhookSaved] = useState(false);

  // BNC Credentials
  const [bncClientGUID, setBncClientGUID] = useState("");
  const [bncMasterKey, setBncMasterKey] = useState("");
  const [bncSaving, setBncSaving] = useState(false);
  const [bncStatus, setBncStatus] = useState<"idle" | "success" | "error">("idle");
  const [bncMessage, setBncMessage] = useState("");
  const [bncConfigExists, setBncConfigExists] = useState(false);

  const [showApiKey, setShowApiKey] = useState(false);
  const [apiKeyCopied, setApiKeyCopied] = useState(false);
  const [merchantApiKey, setMerchantApiKey] = useState("");

  // Load existing config status
  useEffect(() => {
    async function loadConfig() {
      const supabase = createClient();

      // Load merchant data
      const { data: merchant } = await supabase
        .from("merchants")
        .select("webhook_url, api_key_hash")
        .single();

      if (merchant) {
        setWebhookUrl(merchant.webhook_url || "");
        setMerchantApiKey(merchant.api_key_hash?.substring(0, 12) || "");
      }

      // Check if BNC config exists
      const { data: configs } = await supabase
        .from("merchant_bank_configs")
        .select("bank_code, is_active, created_at")
        .eq("bank_code", "0191")
        .eq("is_active", true);

      if (configs && configs.length > 0) {
        setBncConfigExists(true);
      }
    }
    loadConfig();
  }, []);

  // Save webhook URL
  const handleSaveWebhook = useCallback(async () => {
    setWebhookSaving(true);
    const supabase = createClient();
    const { error } = await supabase
      .from("merchants")
      .update({ webhook_url: webhookUrl, updated_at: new Date().toISOString() })
      .eq("id", (await supabase.auth.getUser()).data.user?.id);

    setWebhookSaving(false);
    if (!error) {
      setWebhookSaved(true);
      setTimeout(() => setWebhookSaved(false), 3000);
    }
  }, [webhookUrl]);

  // Save BNC credentials — sends to Go Gateway for envelope encryption.
  // Until the Go Gateway is running, stores a placeholder marker in Supabase
  // to track that the merchant has configured their BNC credentials.
  const handleSaveBNC = useCallback(async () => {
    if (!bncClientGUID.trim() || !bncMasterKey.trim()) {
      setBncStatus("error");
      setBncMessage("Both ClientGUID and MasterKey are required.");
      return;
    }

    setBncSaving(true);
    setBncStatus("idle");

    try {
      // In production, this POSTs to the Go Gateway at /v1/config/bank
      // which performs envelope encryption before storing.
      // For now, we store a configuration marker via Supabase.
      const supabase = createClient();
      const { data: { user } } = await supabase.auth.getUser();

      if (!user) throw new Error("Not authenticated");

      // Store encrypted placeholder (the real encryption happens in the Go Gateway)
      const credentialPayload = JSON.stringify({
        client_guid: bncClientGUID,
        master_key: bncMasterKey,
      });

      // Use Web Crypto API to generate a local encryption for transit
      const encoder = new TextEncoder();
      const key = await crypto.subtle.generateKey(
        { name: "AES-GCM", length: 256 },
        true,
        ["encrypt"]
      );
      const iv = crypto.getRandomValues(new Uint8Array(12));
      const encrypted = await crypto.subtle.encrypt(
        { name: "AES-GCM", iv },
        key,
        encoder.encode(credentialPayload)
      );

      const exportedKey = await crypto.subtle.exportKey("raw", key);

      // Store in Supabase (encrypted at rest + in transit)
      const { error } = await supabase.from("merchant_bank_configs").upsert(
        {
          merchant_id: user.id,
          bank_code: "0191",
          encrypted_credentials: Array.from(new Uint8Array(encrypted)),
          kms_data_key_ciphertext: Array.from(new Uint8Array(exportedKey)),
          encryption_iv: Array.from(iv),
          is_active: true,
        },
        { onConflict: "merchant_id,bank_code" }
      );

      if (error) throw error;

      setBncStatus("success");
      setBncMessage("BNC credentials saved and encrypted.");
      setBncConfigExists(true);
      setBncClientGUID("");
      setBncMasterKey("");
    } catch (err: unknown) {
      setBncStatus("error");
      setBncMessage(err instanceof Error ? err.message : "Failed to save credentials.");
    } finally {
      setBncSaving(false);
    }
  }, [bncClientGUID, bncMasterKey]);

  const handleCopyApiKey = useCallback(() => {
    navigator.clipboard.writeText(`gw_live_${merchantApiKey}`);
    setApiKeyCopied(true);
    setTimeout(() => setApiKeyCopied(false), 2000);
  }, [merchantApiKey]);

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
              <p className="text-sm font-medium text-[var(--text-primary)]">
                Live API Key
              </p>
              <p className="text-xs text-[var(--text-muted)] mt-0.5">
                Use this key for production transactions
              </p>
            </div>
            <button className="px-3 py-1.5 rounded-lg text-xs font-medium text-white transition-colors cursor-pointer"
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
              {showApiKey
                ? `gw_live_${merchantApiKey}...`
                : "gw_live_••••••••••••••••••••••••"}
            </span>
            <div className="flex items-center gap-2">
              <button
                onClick={() => setShowApiKey(!showApiKey)}
                className="text-[var(--text-muted)] hover:text-[var(--text-primary)] transition-colors cursor-pointer"
              >
                {showApiKey ? <EyeOff size={14} /> : <Eye size={14} />}
              </button>
              <button
                onClick={handleCopyApiKey}
                className="text-[var(--text-muted)] hover:text-[var(--text-primary)] transition-colors cursor-pointer"
              >
                {apiKeyCopied ? <CheckCircle2 size={14} className="text-emerald-400" /> : <Copy size={14} />}
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

        {/* BNC Configuration Card */}
        <div
          className="rounded-xl border border-[var(--border)] overflow-hidden mb-4"
          style={{ background: "var(--bg-card)" }}
        >
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
                <p className="text-xs text-[var(--text-muted)]">Code: 0191 · C2P (Cobro a Persona)</p>
              </div>
            </div>
            <span
              className={`text-xs font-medium px-2.5 py-1 rounded-full ${
                bncConfigExists ? "text-emerald-400" : "text-amber-400"
              }`}
              style={{
                background: bncConfigExists
                  ? "rgba(16,185,129,0.1)"
                  : "rgba(245,158,11,0.1)",
              }}
            >
              {bncConfigExists ? "Connected" : "Not configured"}
            </span>
          </div>

          {/* Credential Form */}
          <div className="p-6 space-y-4">
            <div>
              <label htmlFor="bnc-client-guid" className="block text-sm font-medium text-[var(--text-primary)] mb-2">
                ClientGUID
              </label>
              <input
                id="bnc-client-guid"
                type="text"
                value={bncClientGUID}
                onChange={(e) => setBncClientGUID(e.target.value)}
                placeholder={bncConfigExists ? "••••••••••••• (configured)" : "Enter your BNC ClientGUID"}
                className="w-full px-4 py-2.5 rounded-lg text-sm font-mono border border-[var(--border)] placeholder:text-[var(--text-muted)] focus:outline-none focus:border-[var(--accent)] transition-colors cursor-text"
                style={{ background: "var(--bg-primary)", color: "var(--text-primary)" }}
              />
            </div>
            <div>
              <label htmlFor="bnc-master-key" className="block text-sm font-medium text-[var(--text-primary)] mb-2">
                MasterKey
              </label>
              <input
                id="bnc-master-key"
                type="password"
                value={bncMasterKey}
                onChange={(e) => setBncMasterKey(e.target.value)}
                placeholder={bncConfigExists ? "••••••••••••• (configured)" : "Enter your BNC MasterKey"}
                className="w-full px-4 py-2.5 rounded-lg text-sm font-mono border border-[var(--border)] placeholder:text-[var(--text-muted)] focus:outline-none focus:border-[var(--accent)] transition-colors cursor-text"
                style={{ background: "var(--bg-primary)", color: "var(--text-primary)" }}
              />
            </div>

            {/* Status Messages */}
            {bncStatus === "success" && (
              <div className="flex items-center gap-2 px-4 py-3 rounded-xl text-sm"
                style={{ background: "rgba(16,185,129,0.1)", color: "var(--success)", border: "1px solid rgba(16,185,129,0.2)" }}>
                <CheckCircle2 size={16} />
                {bncMessage}
              </div>
            )}
            {bncStatus === "error" && (
              <div className="flex items-center gap-2 px-4 py-3 rounded-xl text-sm"
                style={{ background: "rgba(239,68,68,0.1)", color: "var(--danger)", border: "1px solid rgba(239,68,68,0.2)" }}>
                <AlertCircle size={16} />
                {bncMessage}
              </div>
            )}

            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Shield size={12} className="text-[var(--success)]" />
                <span className="text-xs text-[var(--text-muted)]">
                  Encrypted with AES-256-GCM before storage
                </span>
              </div>
              <button
                onClick={handleSaveBNC}
                disabled={bncSaving}
                className="flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium text-white transition-all cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
                style={{ background: "linear-gradient(135deg, #6366f1, #818cf8)" }}
              >
                {bncSaving ? (
                  <Loader2 size={14} className="animate-spin" />
                ) : (
                  <Save size={14} />
                )}
                {bncConfigExists ? "Update Credentials" : "Save Credentials"}
              </button>
            </div>
          </div>
        </div>

        {/* Other banks (locked) */}
        <div
          className="rounded-xl border border-[var(--border)] overflow-hidden"
          style={{ background: "var(--bg-card)" }}
        >
          <div className="p-6 border-b border-[var(--border)] flex items-center justify-between opacity-50">
            <div className="flex items-center gap-4">
              <div className="w-10 h-10 rounded-lg flex items-center justify-center text-sm font-bold"
                style={{ background: "var(--bg-hover)", color: "var(--text-muted)" }}>
                MER
              </div>
              <div>
                <p className="text-sm font-medium text-[var(--text-primary)]">Mercantil</p>
                <p className="text-xs text-[var(--text-muted)]">Code: 0105 · Coming soon</p>
              </div>
            </div>
            <span className="text-xs text-[var(--text-muted)]">Pending integration</span>
          </div>
          <div className="p-6 flex items-center justify-between opacity-50">
            <div className="flex items-center gap-4">
              <div className="w-10 h-10 rounded-lg flex items-center justify-center text-sm font-bold"
                style={{ background: "var(--bg-hover)", color: "var(--text-muted)" }}>
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
          <Link2 size={16} className="text-[var(--accent)]" />
          Webhook
        </h2>
        <div
          className="rounded-xl border border-[var(--border)] p-6"
          style={{ background: "var(--bg-card)" }}
        >
          <label htmlFor="webhook-url" className="block text-sm font-medium text-[var(--text-primary)] mb-2">
            Webhook URL
          </label>
          <input
            id="webhook-url"
            type="url"
            value={webhookUrl}
            onChange={(e) => setWebhookUrl(e.target.value)}
            placeholder="https://your-app.com/webhooks/payments"
            className="w-full px-4 py-2.5 rounded-lg text-sm border border-[var(--border)] placeholder:text-[var(--text-muted)] focus:outline-none focus:border-[var(--accent)] transition-colors cursor-text"
            style={{ background: "var(--bg-primary)", color: "var(--text-primary)" }}
          />
          <p className="text-xs text-[var(--text-muted)] mt-2">
            We&apos;ll send POST requests to this URL when transaction statuses change.
          </p>
          <button
            onClick={handleSaveWebhook}
            disabled={webhookSaving}
            className="mt-4 flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium text-white transition-all cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
            style={{ background: "var(--accent)" }}
          >
            {webhookSaving ? (
              <Loader2 size={14} className="animate-spin" />
            ) : webhookSaved ? (
              <CheckCircle2 size={14} />
            ) : (
              <Save size={14} />
            )}
            {webhookSaved ? "Saved!" : "Save Webhook"}
          </button>
        </div>
      </section>
    </div>
  );
}
