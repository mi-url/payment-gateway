-- Migration: 002_strict_rls
-- Description: Replaces dev-time permissive RLS policies with strict
--   per-merchant isolation based on auth.uid(). Adds auto-provisioning
--   trigger so a merchant row is created on Supabase Auth signup.
-- Author: Payment Gateway Team
-- Date: 2026-04-21

-- ================================================================
-- 1. Drop the temporary dev policies
-- ================================================================
DROP POLICY IF EXISTS merchants_select ON merchants;
DROP POLICY IF EXISTS merchants_anon_select ON merchants;
DROP POLICY IF EXISTS transactions_select ON transactions;
DROP POLICY IF EXISTS transactions_anon_select ON transactions;
DROP POLICY IF EXISTS bank_configs_select ON merchant_bank_configs;

-- ================================================================
-- 2. Strict per-merchant RLS policies
-- ================================================================

-- merchants: users can only read their own row (id = auth.uid())
CREATE POLICY merchants_own_select ON merchants
  FOR SELECT TO authenticated
  USING (id = auth.uid());

CREATE POLICY merchants_own_update ON merchants
  FOR UPDATE TO authenticated
  USING (id = auth.uid())
  WITH CHECK (id = auth.uid());

-- transactions: users see only their own transactions
CREATE POLICY transactions_own_select ON transactions
  FOR SELECT TO authenticated
  USING (merchant_id = auth.uid());

CREATE POLICY transactions_own_insert ON transactions
  FOR INSERT TO authenticated
  WITH CHECK (merchant_id = auth.uid());

-- merchant_bank_configs: users manage only their own configs
CREATE POLICY bank_configs_own_select ON merchant_bank_configs
  FOR SELECT TO authenticated
  USING (merchant_id = auth.uid());

CREATE POLICY bank_configs_own_insert ON merchant_bank_configs
  FOR INSERT TO authenticated
  WITH CHECK (merchant_id = auth.uid());

CREATE POLICY bank_configs_own_update ON merchant_bank_configs
  FOR UPDATE TO authenticated
  USING (merchant_id = auth.uid())
  WITH CHECK (merchant_id = auth.uid());

-- Also enable RLS on merchants table (was missing in 001)
ALTER TABLE merchants ENABLE ROW LEVEL SECURITY;

-- ================================================================
-- 3. Auto-provision merchant on Supabase Auth signup
-- ================================================================

-- This function runs after a new user is inserted into auth.users.
-- It creates a corresponding row in public.merchants so the user
-- immediately has a merchant_id matching their auth.uid().
CREATE OR REPLACE FUNCTION public.handle_new_user()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER SET search_path = ''
AS $$
BEGIN
  INSERT INTO public.merchants (id, name, api_key_hash)
  VALUES (
    new.id,
    COALESCE(new.raw_user_meta_data ->> 'company_name', new.email),
    encode(sha256(gen_random_uuid()::text::bytea), 'hex')
  );
  RETURN new;
END;
$$;

-- Trigger: fires after each INSERT on auth.users
CREATE OR REPLACE TRIGGER on_auth_user_created
  AFTER INSERT ON auth.users
  FOR EACH ROW EXECUTE FUNCTION public.handle_new_user();
