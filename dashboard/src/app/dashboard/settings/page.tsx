"use client";

import { useEffect, useState } from "react";
import { Building2, Shield, Save, Loader2 } from "lucide-react";
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
import { Badge } from "@/components/ui/badge";

const bncSchema = z.object({
  client_guid: z.string().min(8, { message: "ClientGUID must be at least 8 characters" }),
  master_key: z.string().min(8, { message: "MasterKey must be at least 8 characters" }),
});

type BncFormValues = z.infer<typeof bncSchema>;

export default function SettingsPage() {
  const [bncSaving, setBncSaving] = useState(false);
  const [bncConfigExists, setBncConfigExists] = useState(false);

  const bncForm = useForm<BncFormValues>({
    resolver: zodResolver(bncSchema),
    defaultValues: { client_guid: "", master_key: "" },
  });

  useEffect(() => {
    async function loadConfig() {
      const supabase = createClient();
      const { data: configs } = await supabase
        .from("merchant_bank_configs")
        .select("bank_code, is_active")
        .eq("bank_code", "0191")
        .eq("is_active", true);

      if (configs && configs.length > 0) {
        setBncConfigExists(true);
      }
    }
    loadConfig();
  }, []);

  const onBncSubmit = async (data: BncFormValues) => {
    setBncSaving(true);
    try {
      const res = await fetch("/api/config/bank", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          bank_code: "0191",
          credentials: {
            client_guid: data.client_guid,
            master_key: data.master_key,
          },
        }),
      });

      const resData = await res.json();
      if (!res.ok) throw new Error(resData.error || "Failed to save credentials");

      toast.success("BNC credentials securely encrypted and saved.");
      setBncConfigExists(true);
      bncForm.reset();
    } catch (err: unknown) {
      toast.error(err instanceof Error ? err.message : "Failed to save credentials.");
    } finally {
      setBncSaving(false);
    }
  };

  return (
    <div className="max-w-4xl space-y-6 animate-in fade-in duration-500">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Settings</h1>
        <p className="text-muted-foreground mt-1">
          Manage your business and bank integration configurations.
        </p>
      </div>

      <div className="space-y-4">
        <Card className="bg-card/50 backdrop-blur-sm border-border/50 border-primary/20 shadow-sm">
          <CardHeader className="pb-4">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-4">
                <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-primary/10 text-primary font-bold text-sm">
                  <Building2 className="h-5 w-5" />
                </div>
                <div>
                  <CardTitle className="text-base">Banco Nacional de Crédito</CardTitle>
                  <CardDescription>Code: 0191 • C2P (Cobro a Persona)</CardDescription>
                </div>
              </div>
              <Badge variant={bncConfigExists ? "default" : "secondary"} className={bncConfigExists ? "bg-emerald-500/10 text-emerald-500 hover:bg-emerald-500/20" : ""}>
                {bncConfigExists ? "Connected" : "Not configured"}
              </Badge>
            </div>
          </CardHeader>
          <CardContent>
            <form onSubmit={bncForm.handleSubmit(onBncSubmit)} className="space-y-4">
              <div className="grid gap-2">
                <Label htmlFor="bnc-client-guid">ClientGUID</Label>
                <Input
                  id="bnc-client-guid"
                  placeholder={bncConfigExists ? "••••••••••••• (configured)" : "Enter your BNC ClientGUID"}
                  className="font-mono"
                  {...bncForm.register("client_guid")}
                />
                {bncForm.formState.errors.client_guid && (
                  <p className="text-[13px] text-destructive font-medium mt-1">
                    {bncForm.formState.errors.client_guid.message}
                  </p>
                )}
              </div>
              <div className="grid gap-2">
                <Label htmlFor="bnc-master-key">MasterKey</Label>
                <Input
                  id="bnc-master-key"
                  type="password"
                  placeholder={bncConfigExists ? "••••••••••••• (configured)" : "Enter your BNC MasterKey"}
                  className="font-mono"
                  {...bncForm.register("master_key")}
                />
                {bncForm.formState.errors.master_key && (
                  <p className="text-[13px] text-destructive font-medium mt-1">
                    {bncForm.formState.errors.master_key.message}
                  </p>
                )}
              </div>
              <div className="flex items-center justify-between pt-2">
                <div className="flex items-center gap-2 text-xs text-muted-foreground">
                  <Shield className="h-4 w-4 text-emerald-500" />
                  Encrypted with AES-256-GCM before storage
                </div>
                <Button type="submit" disabled={bncSaving}>
                  {bncSaving ? (
                    <>
                      <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                      Saving...
                    </>
                  ) : (
                    <>
                      {bncConfigExists ? "Update Credentials" : "Save Credentials"}
                      <Save className="ml-2 h-4 w-4" />
                    </>
                  )}
                </Button>
              </div>
            </form>
          </CardContent>
        </Card>

        <Card className="bg-card/50 backdrop-blur-sm border-border/50 opacity-60">
          <CardHeader>
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-4">
                <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-secondary text-muted-foreground font-bold text-sm">
                  MER
                </div>
                <div>
                  <CardTitle className="text-base text-muted-foreground">Mercantil</CardTitle>
                  <CardDescription>Code: 0105 • Coming soon</CardDescription>
                </div>
              </div>
              <Badge variant="outline">Pending Integration</Badge>
            </div>
          </CardHeader>
        </Card>
      </div>
    </div>
  );
}
