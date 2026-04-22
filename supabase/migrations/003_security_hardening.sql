-- Migration: 003_security_hardening
-- Description: Closes security gaps identified in the audit:
--   1. Revokes all unnecessary anon privileges
--   2. Moves SECURITY DEFINER function to private schema
--   3. Consolidates all GRANTs into a single auditable file
-- Author: Payment Gateway Team
-- Date: 2026-04-21

-- ================================================================
-- 1. Revoke ALL anon privileges on application tables
-- ================================================================
REVOKE ALL ON merchants FROM anon;
REVOKE ALL ON transactions FROM anon;
REVOKE ALL ON merchant_bank_configs FROM anon;

-- ================================================================
-- 2. Revoke excessive privileges from authenticated
--    (only grant what's actually needed)
-- ================================================================
REVOKE ALL ON merchants FROM authenticated;
REVOKE ALL ON transactions FROM authenticated;
REVOKE ALL ON merchant_bank_configs FROM authenticated;

-- Re-grant ONLY what authenticated needs:
-- merchants: read own row, update own webhook_url
GRANT SELECT, UPDATE ON merchants TO authenticated;

-- transactions: read own transactions (inserts come from the Go Gateway via service_role)
GRANT SELECT ON transactions TO authenticated;

-- merchant_bank_configs: read own configs, insert new ones, update existing
GRANT SELECT, INSERT, UPDATE ON merchant_bank_configs TO authenticated;

-- ================================================================
-- 3. Move handle_new_user() from public to private schema
-- ================================================================

-- Create private schema (not exposed via Data API)
CREATE SCHEMA IF NOT EXISTS private;

-- Drop the existing trigger first
DROP TRIGGER IF EXISTS on_auth_user_created ON auth.users;

-- Drop the existing function in public
DROP FUNCTION IF EXISTS public.handle_new_user();

-- Recreate in private schema
CREATE OR REPLACE FUNCTION private.handle_new_user()
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

-- Recreate trigger pointing to private schema
CREATE TRIGGER on_auth_user_created
  AFTER INSERT ON auth.users
  FOR EACH ROW EXECUTE FUNCTION private.handle_new_user();

-- ================================================================
-- 4. Verify final state
-- ================================================================
-- After this migration:
--   anon:          0 privileges on any application table
--   authenticated: SELECT/UPDATE on merchants
--                  SELECT on transactions
--                  SELECT/INSERT/UPDATE on merchant_bank_configs
--   private.handle_new_user(): SECURITY DEFINER in unexposed schema
