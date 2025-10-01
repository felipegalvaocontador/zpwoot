-- =====================================================
-- zpwoot Database Schema - Rollback Additional Indexes
-- Clean Architecture Implementation
-- =====================================================

-- Drop text search indexes
DROP INDEX IF EXISTS "idx_zp_message_content_text";
DROP INDEX IF EXISTS "idx_zp_chatwoot_inbox_name_text";
DROP INDEX IF EXISTS "idx_zp_sessions_name_text";

-- Drop partial indexes
DROP INDEX IF EXISTS "idx_zp_message_failed_sync";
DROP INDEX IF EXISTS "idx_zp_message_pending_sync";
DROP INDEX IF EXISTS "idx_zp_chatwoot_enabled_only";
DROP INDEX IF EXISTS "idx_zp_webhooks_enabled_only";
DROP INDEX IF EXISTS "idx_zp_sessions_connected_only";

-- Drop additional performance indexes for zpMessage
DROP INDEX IF EXISTS "idx_zp_message_type_timestamp";
DROP INDEX IF EXISTS "idx_zp_message_sender_timestamp";
DROP INDEX IF EXISTS "idx_zp_message_chat_timestamp";
DROP INDEX IF EXISTS "idx_zp_message_session_timestamp";
DROP INDEX IF EXISTS "idx_zp_message_synced_at";

-- Drop additional performance indexes for zpChatwoot
DROP INDEX IF EXISTS "idx_zp_chatwoot_updated_at";
DROP INDEX IF EXISTS "idx_zp_chatwoot_enabled_session";
DROP INDEX IF EXISTS "idx_zp_chatwoot_account_inbox";

-- Drop additional performance indexes for zpWebhooks
DROP INDEX IF EXISTS "idx_zp_webhooks_updated_at";
DROP INDEX IF EXISTS "idx_zp_webhooks_session_enabled";

-- Drop additional performance indexes for zpSessions
DROP INDEX IF EXISTS "idx_zp_sessions_connected_status";
DROP INDEX IF EXISTS "idx_zp_sessions_last_seen";