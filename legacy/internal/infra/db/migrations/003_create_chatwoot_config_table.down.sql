-- Drop chatwoot_config table
DROP TRIGGER IF EXISTS update_zp_chatwoot_updated_at ON "zpChatwoot";
DROP INDEX IF EXISTS "idx_zp_chatwoot_unique_session";
DROP INDEX IF EXISTS "idx_zp_chatwoot_created_at";
DROP INDEX IF EXISTS "idx_zp_chatwoot_enabled";
DROP INDEX IF EXISTS "idx_zp_chatwoot_session_id";
DROP TABLE IF EXISTS "zpChatwoot";
