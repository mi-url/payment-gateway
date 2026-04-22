"use client";

import { useEffect, useState, useCallback } from "react";
import {
  KeyRound,
  Copy,
  Eye,
  EyeOff,
  Shield,
  Save,
  Link2,
  Loader2,
  Terminal,
  Code2,
} from "lucide-react";
import { createClient } from "@/lib/supabase/client";
import { toast } from "sonner";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

const webhookSchema = z.object({
  webhook_url: z.string().url({ message: "Must be a valid URL (e.g. https://api.yoursite.com)" }).min(1, { message: "URL is required" }),
});

type WebhookFormValues = z.infer<typeof webhookSchema>;

export default function DevelopersPage() {
  const [webhookSaving, setWebhookSaving] = useState(false);
  const [showApiKey, setShowApiKey] = useState(false);
  const [merchantApiKey, setMerchantApiKey] = useState("");

  const webhookForm = useForm<WebhookFormValues>({
    resolver: zodResolver(webhookSchema),
    defaultValues: { webhook_url: "" },
  });

  useEffect(() => {
    async function loadConfig() {
      const supabase = createClient();
      const { data: merchant } = await supabase
        .from("merchants")
        .select("webhook_url, api_key_hash")
        .single();

      if (merchant) {
        webhookForm.reset({ webhook_url: merchant.webhook_url || "" });
        setMerchantApiKey(merchant.api_key_hash?.substring(0, 12) || "test_key");
      }
    }
    loadConfig();
  }, [webhookForm]);

  const onWebhookSubmit = async (data: WebhookFormValues) => {
    setWebhookSaving(true);
    const supabase = createClient();
    const { error } = await supabase
      .from("merchants")
      .update({ webhook_url: data.webhook_url, updated_at: new Date().toISOString() })
      .eq("id", (await supabase.auth.getUser()).data.user?.id);

    setWebhookSaving(false);
    if (!error) {
      toast.success("Webhook URL saved successfully");
    } else {
      toast.error("Failed to save webhook URL");
    }
  };

  const handleCopy = useCallback((text: string, message: string) => {
    navigator.clipboard.writeText(text);
    toast.success(message);
  }, []);

  const fullApiKey = `gw_live_${merchantApiKey}`;
  const maskedApiKey = `gw_live_••••••••••••••••••••••••`;

  const curlSnippet = `curl -X POST https://api.faloppa.com/v1/charges \\
  -H "Authorization: Bearer ${fullApiKey}" \\
  -H "Content-Type: application/json" \\
  -d '{
    "amount": 150.00,
    "bank_code": "0191",
    "payer_phone": "04141234567",
    "payer_id_document": "V12345678"
  }'`;

  const nodeSnippet = `const fetch = require('node-fetch');

const response = await fetch('https://api.faloppa.com/v1/charges', {
  method: 'POST',
  headers: {
    'Authorization': 'Bearer ${fullApiKey}',
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    amount: 150.00,
    bank_code: "0191",
    payer_phone: "04141234567",
    payer_id_document: "V12345678"
  })
});

const data = await response.json();
console.log(data);`;

  const goSnippet = `package main

import (
	"bytes"
	"encoding/json"
	"net/http"
)

func main() {
	payload := map[string]interface{}{
		"amount":            150.00,
		"bank_code":         "0191",
		"payer_phone":       "04141234567",
		"payer_id_document": "V12345678",
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "https://api.faloppa.com/v1/charges", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer ${fullApiKey}")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	client.Do(req)
}`;

  return (
    <div className="max-w-5xl space-y-8 animate-in fade-in duration-500 pb-10">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Developers</h1>
        <p className="text-muted-foreground mt-1">
          API Keys, Webhooks, and Integration Guides for your Faloppa Gateway.
        </p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="space-y-6">
          {/* API Keys */}
          <Card className="bg-card/50 backdrop-blur-sm border-border/50">
            <CardHeader>
              <div className="flex items-center justify-between">
                <div>
                  <CardTitle className="flex items-center gap-2">
                    <KeyRound className="h-4 w-4" /> Live API Key
                  </CardTitle>
                  <CardDescription className="mt-1">Authenticate requests from your backend.</CardDescription>
                </div>
                <Button variant="outline" size="sm">Roll Key</Button>
              </div>
            </CardHeader>
            <CardContent>
              <div className="flex items-center justify-between px-4 py-3 rounded-md bg-secondary/50 border border-border font-mono text-sm">
                <span className="text-foreground/80 break-all mr-2">
                  {showApiKey ? fullApiKey : maskedApiKey}
                </span>
                <div className="flex items-center gap-2 shrink-0">
                  <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => setShowApiKey(!showApiKey)}>
                    {showApiKey ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                  </Button>
                  <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => handleCopy(fullApiKey, "API Key copied to clipboard")}>
                    <Copy className="h-4 w-4" />
                  </Button>
                </div>
              </div>
              <div className="flex items-center gap-2 mt-4 text-xs text-muted-foreground">
                <Shield className="h-4 w-4 text-emerald-500" />
                Never expose this key in client-side code (browsers or mobile apps).
              </div>
            </CardContent>
          </Card>

          {/* Webhooks */}
          <Card className="bg-card/50 backdrop-blur-sm border-border/50">
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Link2 className="h-4 w-4" /> Webhook Events
              </CardTitle>
              <CardDescription className="mt-1">Receive real-time HTTP POST notifications when a transaction succeeds or fails.</CardDescription>
            </CardHeader>
            <CardContent>
              <form onSubmit={webhookForm.handleSubmit(onWebhookSubmit)} className="space-y-4">
                <div className="grid gap-2">
                  <Label htmlFor="webhook-url">Endpoint URL</Label>
                  <Input
                    id="webhook-url"
                    type="url"
                    placeholder="https://api.yourdomain.com/webhooks/faloppa"
                    {...webhookForm.register("webhook_url")}
                  />
                  {webhookForm.formState.errors.webhook_url && (
                    <p className="text-[13px] text-destructive font-medium mt-1">
                      {webhookForm.formState.errors.webhook_url.message}
                    </p>
                  )}
                </div>
                <Button type="submit" disabled={webhookSaving} variant="secondary">
                  {webhookSaving ? (
                    <>
                      <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                      Saving...
                    </>
                  ) : (
                    <>
                      Save Webhook
                      <Save className="ml-2 h-4 w-4" />
                    </>
                  )}
                </Button>
              </form>
            </CardContent>
          </Card>
        </div>

        <div className="space-y-6">
          {/* Quick Integration */}
          <Card className="bg-[#0D0D0D] border-border/50 shadow-xl overflow-hidden text-slate-300 h-full">
            <CardHeader className="bg-white/5 border-b border-white/10 pb-4">
              <CardTitle className="text-white flex items-center gap-2 text-base">
                <Terminal className="h-4 w-4 text-emerald-400" /> Quick Integration
              </CardTitle>
              <CardDescription className="text-slate-400">
                Create a C2P Charge using your actual API Key.
              </CardDescription>
            </CardHeader>
            <CardContent className="p-0">
              <Tabs defaultValue="curl" className="w-full">
                <TabsList className="w-full justify-start rounded-none border-b border-white/10 bg-transparent p-0 h-auto">
                  <TabsTrigger 
                    value="curl" 
                    className="rounded-none data-[state=active]:bg-white/5 data-[state=active]:text-white border-b-2 border-transparent data-[state=active]:border-emerald-500 px-4 py-2 text-sm"
                  >
                    cURL
                  </TabsTrigger>
                  <TabsTrigger 
                    value="node" 
                    className="rounded-none data-[state=active]:bg-white/5 data-[state=active]:text-white border-b-2 border-transparent data-[state=active]:border-emerald-500 px-4 py-2 text-sm"
                  >
                    Node.js
                  </TabsTrigger>
                  <TabsTrigger 
                    value="go" 
                    className="rounded-none data-[state=active]:bg-white/5 data-[state=active]:text-white border-b-2 border-transparent data-[state=active]:border-emerald-500 px-4 py-2 text-sm"
                  >
                    Go
                  </TabsTrigger>
                </TabsList>
                
                <div className="relative group">
                  <TabsContent value="curl" className="p-4 m-0">
                    <pre className="text-xs font-mono leading-relaxed overflow-x-auto text-emerald-300/90">
                      {curlSnippet}
                    </pre>
                    <Button 
                      variant="ghost" 
                      size="icon" 
                      className="absolute top-4 right-4 h-8 w-8 opacity-0 group-hover:opacity-100 transition-opacity bg-white/10 hover:bg-white/20 text-white"
                      onClick={() => handleCopy(curlSnippet, "Code copied")}
                    >
                      <Copy className="h-4 w-4" />
                    </Button>
                  </TabsContent>

                  <TabsContent value="node" className="p-4 m-0">
                    <pre className="text-xs font-mono leading-relaxed overflow-x-auto text-blue-300/90">
                      {nodeSnippet}
                    </pre>
                    <Button 
                      variant="ghost" 
                      size="icon" 
                      className="absolute top-4 right-4 h-8 w-8 opacity-0 group-hover:opacity-100 transition-opacity bg-white/10 hover:bg-white/20 text-white"
                      onClick={() => handleCopy(nodeSnippet, "Code copied")}
                    >
                      <Copy className="h-4 w-4" />
                    </Button>
                  </TabsContent>

                  <TabsContent value="go" className="p-4 m-0">
                    <pre className="text-xs font-mono leading-relaxed overflow-x-auto text-cyan-300/90">
                      {goSnippet}
                    </pre>
                    <Button 
                      variant="ghost" 
                      size="icon" 
                      className="absolute top-4 right-4 h-8 w-8 opacity-0 group-hover:opacity-100 transition-opacity bg-white/10 hover:bg-white/20 text-white"
                      onClick={() => handleCopy(goSnippet, "Code copied")}
                    >
                      <Copy className="h-4 w-4" />
                    </Button>
                  </TabsContent>
                </div>
              </Tabs>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
