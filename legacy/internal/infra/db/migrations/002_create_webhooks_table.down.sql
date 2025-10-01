-- Drop webhooks table
DROP TRIGGER IF EXISTS update_zp_webhooks_updated_at ON "zpWebhooks";
DROP INDEX IF EXISTS "idx_zp_webhooks_enabled";
DROP INDEX IF EXISTS "idx_zp_webhooks_session_id";
DROP INDEX IF EXISTS "idx_zp_webhooks_created_at";
DROP TABLE IF EXISTS "zpWebhooks";
