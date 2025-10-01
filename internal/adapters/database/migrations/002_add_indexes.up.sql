-- =====================================================
-- zpwoot Database Schema - Additional Indexes
-- Clean Architecture Implementation
-- Performance optimization indexes
-- =====================================================

-- Additional performance indexes for zpSessions
CREATE INDEX IF NOT EXISTS "idx_zp_sessions_last_seen" ON "zpSessions" ("lastSeen");
CREATE INDEX IF NOT EXISTS "idx_zp_sessions_connected_status" ON "zpSessions" ("isConnected", "connectedAt");

-- Additional performance indexes for zpWebhooks
CREATE INDEX IF NOT EXISTS "idx_zp_webhooks_session_enabled" ON "zpWebhooks" ("sessionId", "enabled");
CREATE INDEX IF NOT EXISTS "idx_zp_webhooks_updated_at" ON "zpWebhooks" ("updatedAt");

-- Additional performance indexes for zpChatwoot
CREATE INDEX IF NOT EXISTS "idx_zp_chatwoot_account_inbox" ON "zpChatwoot" ("accountId", "inboxId");
CREATE INDEX IF NOT EXISTS "idx_zp_chatwoot_enabled_session" ON "zpChatwoot" ("enabled", "sessionId");
CREATE INDEX IF NOT EXISTS "idx_zp_chatwoot_updated_at" ON "zpChatwoot" ("updatedAt");

-- Additional performance indexes for zpMessage
CREATE INDEX IF NOT EXISTS "idx_zp_message_synced_at" ON "zpMessage" ("syncedAt");
CREATE INDEX IF NOT EXISTS "idx_zp_message_session_timestamp" ON "zpMessage" ("sessionId", "zpTimestamp");
CREATE INDEX IF NOT EXISTS "idx_zp_message_chat_timestamp" ON "zpMessage" ("zpChat", "zpTimestamp");
CREATE INDEX IF NOT EXISTS "idx_zp_message_sender_timestamp" ON "zpMessage" ("zpSender", "zpTimestamp");
CREATE INDEX IF NOT EXISTS "idx_zp_message_type_timestamp" ON "zpMessage" ("zpType", "zpTimestamp");

-- Partial indexes for common filtered queries
CREATE INDEX IF NOT EXISTS "idx_zp_sessions_connected_only" ON "zpSessions" ("id", "name") WHERE "isConnected" = true;
CREATE INDEX IF NOT EXISTS "idx_zp_webhooks_enabled_only" ON "zpWebhooks" ("sessionId", "url") WHERE "enabled" = true;
CREATE INDEX IF NOT EXISTS "idx_zp_chatwoot_enabled_only" ON "zpChatwoot" ("sessionId", "accountId") WHERE "enabled" = true;
CREATE INDEX IF NOT EXISTS "idx_zp_message_pending_sync" ON "zpMessage" ("sessionId", "createdAt") WHERE "syncStatus" = 'pending';
CREATE INDEX IF NOT EXISTS "idx_zp_message_failed_sync" ON "zpMessage" ("sessionId", "createdAt") WHERE "syncStatus" = 'failed';

-- Text search indexes for better search performance
CREATE INDEX IF NOT EXISTS "idx_zp_sessions_name_text" ON "zpSessions" USING gin(to_tsvector('english', "name"));
CREATE INDEX IF NOT EXISTS "idx_zp_chatwoot_inbox_name_text" ON "zpChatwoot" USING gin(to_tsvector('english', "inboxName")) WHERE "inboxName" IS NOT NULL;
CREATE INDEX IF NOT EXISTS "idx_zp_message_content_text" ON "zpMessage" USING gin(to_tsvector('english', "content")) WHERE "content" IS NOT NULL;

-- Comments for documentation
COMMENT ON INDEX "idx_zp_sessions_connected_status" IS 'Optimizes queries for connected sessions with connection time';
COMMENT ON INDEX "idx_zp_message_session_timestamp" IS 'Optimizes message queries by session and time range';
COMMENT ON INDEX "idx_zp_message_pending_sync" IS 'Optimizes queries for messages pending synchronization';
COMMENT ON INDEX "idx_zp_sessions_name_text" IS 'Full-text search index for session names';
COMMENT ON INDEX "idx_zp_message_content_text" IS 'Full-text search index for message content';