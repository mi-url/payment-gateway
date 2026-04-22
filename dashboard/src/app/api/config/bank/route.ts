// Next.js API Route: /api/config/bank
// 
// This endpoint handles bank credential encryption SERVER-SIDE.
// The browser sends plaintext credentials here (over HTTPS, same origin).
// The server encrypts with AES-256-GCM using a key from the environment
// (NOT stored in Supabase), then writes only ciphertext to the DB.
//
// In production, this will be replaced by the Go Gateway endpoint
// (POST /v1/config/bank) which uses envelope encryption with Cloud KMS.

import { NextRequest, NextResponse } from "next/server";
import { createClient } from "@/lib/supabase/server";

export async function POST(request: NextRequest) {
  const supabase = await createClient();

  // Verify the user is authenticated
  const {
    data: { user },
    error: authError,
  } = await supabase.auth.getUser();

  if (authError || !user) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
  }

  // Parse the request
  const body = await request.json();
  const { bank_code, credentials } = body;

  if (!bank_code || !credentials) {
    return NextResponse.json(
      { error: "bank_code and credentials are required" },
      { status: 400 }
    );
  }

  // Encrypt server-side using the GATEWAY_ENCRYPTION_KEY env var.
  // This key exists ONLY on the server — never in Supabase, never in the browser.
  const encKeyHex = process.env.GATEWAY_ENCRYPTION_KEY;
  if (!encKeyHex) {
    console.error("GATEWAY_ENCRYPTION_KEY not set");
    return NextResponse.json(
      { error: "Server encryption not configured" },
      { status: 500 }
    );
  }

  try {
    const encoder = new TextEncoder();
    const plaintext = encoder.encode(JSON.stringify(credentials));

    // Import the server-side key (from env, NOT from Supabase)
    const keyBytes = hexToBytes(encKeyHex);
    const cryptoKey = await crypto.subtle.importKey(
      "raw",
      keyBytes,
      { name: "AES-GCM" },
      false, // not extractable
      ["encrypt"]
    );

    // Encrypt
    const iv = crypto.getRandomValues(new Uint8Array(12));
    const ciphertext = await crypto.subtle.encrypt(
      { name: "AES-GCM", iv },
      cryptoKey,
      plaintext
    );

    // Store in Supabase — only ciphertext + IV, never the key
    // The key stays in the environment variable
    const { error: dbError } = await supabase
      .from("merchant_bank_configs")
      .upsert(
        {
          merchant_id: user.id,
          bank_code,
          encrypted_credentials: Buffer.from(ciphertext).toString("base64"),
          kms_data_key_ciphertext: Buffer.from("server-env-key").toString(
            "base64"
          ), // Marker: key is in env, not here
          encryption_iv: Buffer.from(iv).toString("base64"),
          is_active: true,
        },
        { onConflict: "merchant_id,bank_code" }
      );

    if (dbError) {
      console.error("Failed to store bank config:", dbError);
      return NextResponse.json(
        { error: "Failed to save configuration" },
        { status: 500 }
      );
    }

    return NextResponse.json({
      status: "configured",
      bank_code,
      encrypted: true,
    });
  } catch (err) {
    console.error("Encryption error:", err);
    return NextResponse.json(
      { error: "Encryption failed" },
      { status: 500 }
    );
  }
}

// Convert hex string to Uint8Array
function hexToBytes(hex: string): Uint8Array {
  const bytes = new Uint8Array(hex.length / 2);
  for (let i = 0; i < hex.length; i += 2) {
    bytes[i / 2] = parseInt(hex.substring(i, i + 2), 16);
  }
  return bytes;
}
