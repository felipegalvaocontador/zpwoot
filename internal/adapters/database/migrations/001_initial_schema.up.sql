-- =====================================================
-- zpwoot Database Schema - Initial Migration
-- Clean Architecture Implementation
-- =====================================================

-- Create function for automatic updatedAt trigger
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW."updatedAt" = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- =====================================================
-- Sessions Table - Core WhatsApp Sessions
-- =====================================================
CREATE TABLE IF NOT EXISTS "zpSessions" (
    "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "name" VARCHAR(255) NOT NULL UNIQUE,
    "deviceJid" VARCHAR(255) UNIQUE,
    "isConnected" BOOLEAN NOT NULL DEFAULT false,
    "connectionError" TEXT,
    "qrCode" TEXT,
    "qrCodeExpiresAt" TIMESTAMP WITH TIME ZONE,
    "proxyConfig" JSONB,
    "createdAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    "updatedAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    "connectedAt" TIMESTAMP WITH TIME ZONE,
    "lastSeen" TIMESTAMP WITH TIME ZONE
);

-- Sessions indexes
CREATE INDEX IF NOT EXISTS "idx_zp_sessions_name" ON "zpSessions" ("name");
CREATE INDEX IF NOT EXISTS "idx_zp_sessions_is_connected" ON "zpSessions" ("isConnected");
CREATE INDEX IF NOT EXISTS "idx_zp_sessions_device_jid" ON "zpSessions" ("deviceJid");
CREATE INDEX IF NOT EXISTS "idx_zp_sessions_created_at" ON "zpSessions" ("createdAt");
CREATE INDEX IF NOT EXISTS "idx_zp_sessions_updated_at" ON "zpSessions" ("updatedAt");
CREATE INDEX IF NOT EXISTS "idx_zp_sessions_connected_at" ON "zpSessions" ("connectedAt");
CREATE INDEX IF NOT EXISTS "idx_zp_sessions_qr_expires" ON "zpSessions" ("qrCodeExpiresAt");

-- Sessions trigger
CREATE TRIGGER update_zp_sessions_updated_at
    BEFORE UPDATE ON "zpSessions"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- =====================================================
-- Webhooks Table - Event Notifications
-- =====================================================
CREATE TABLE IF NOT EXISTS "zpWebhooks" (
    "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "sessionId" UUID REFERENCES "zpSessions"("id") ON DELETE CASCADE,
    "url" VARCHAR(2048) NOT NULL,
    "secret" VARCHAR(255),
    "events" JSONB NOT NULL DEFAULT '[]',
    "enabled" BOOLEAN NOT NULL DEFAULT true,
    "createdAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    "updatedAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Webhooks indexes
CREATE INDEX IF NOT EXISTS "idx_zp_webhooks_session_id" ON "zpWebhooks" ("sessionId");
CREATE INDEX IF NOT EXISTS "idx_zp_webhooks_enabled" ON "zpWebhooks" ("enabled");
CREATE INDEX IF NOT EXISTS "idx_zp_webhooks_created_at" ON "zpWebhooks" ("createdAt");

-- Webhooks trigger
CREATE TRIGGER update_zp_webhooks_updated_at
    BEFORE UPDATE ON "zpWebhooks"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- =====================================================
-- Messages Table - WhatsApp <-> Chatwoot Mapping
-- =====================================================
CREATE TABLE IF NOT EXISTS "zpMessage" (
    "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "sessionId" UUID NOT NULL REFERENCES "zpSessions"("id") ON DELETE CASCADE,

    -- WhatsApp Message Identifiers (from whatsmeow)
    "zpMessageId" VARCHAR(255) NOT NULL,
    "zpSender" VARCHAR(255) NOT NULL,
    "zpChat" VARCHAR(255) NOT NULL,
    "zpTimestamp" TIMESTAMP WITH TIME ZONE NOT NULL,
    "zpFromMe" BOOLEAN NOT NULL,
    "zpType" VARCHAR(50) NOT NULL, -- text, image, audio, video, document, contact, etc.
    "content" TEXT,

    -- Chatwoot Message Identifiers
    "cwMessageId" INTEGER,
    "cwConversationId" INTEGER,

    -- Sync Status
    "syncStatus" VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK ("syncStatus" IN ('pending', 'synced', 'failed')),

    -- Timestamps
    "createdAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    "updatedAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    "syncedAt" TIMESTAMP WITH TIME ZONE
);

-- Messages indexes
CREATE INDEX IF NOT EXISTS "idx_zp_message_session_id" ON "zpMessage" ("sessionId");
CREATE INDEX IF NOT EXISTS "idx_zp_message_zp_message_id" ON "zpMessage" ("zpMessageId");
CREATE INDEX IF NOT EXISTS "idx_zp_message_zp_chat" ON "zpMessage" ("zpChat");
CREATE INDEX IF NOT EXISTS "idx_zp_message_cw_message_id" ON "zpMessage" ("cwMessageId");
CREATE INDEX IF NOT EXISTS "idx_zp_message_cw_conversation_id" ON "zpMessage" ("cwConversationId");
CREATE INDEX IF NOT EXISTS "idx_zp_message_sync_status" ON "zpMessage" ("syncStatus");
CREATE INDEX IF NOT EXISTS "idx_zp_message_timestamp" ON "zpMessage" ("zpTimestamp");
CREATE INDEX IF NOT EXISTS "idx_zp_message_zp_type" ON "zpMessage" ("zpType");
CREATE INDEX IF NOT EXISTS "idx_zp_message_zp_from_me" ON "zpMessage" ("zpFromMe");
CREATE INDEX IF NOT EXISTS "idx_zp_message_created_at" ON "zpMessage" ("createdAt");

-- Composite indexes for common queries
CREATE INDEX IF NOT EXISTS "idx_zp_message_session_chat" ON "zpMessage" ("sessionId", "zpChat");
CREATE INDEX IF NOT EXISTS "idx_zp_message_cw_conversation_status" ON "zpMessage" ("cwConversationId", "syncStatus");

-- Unique constraint to prevent duplicate message mapping
CREATE UNIQUE INDEX IF NOT EXISTS "idx_zp_message_unique_zp" ON "zpMessage" ("sessionId", "zpMessageId");

-- Messages trigger
CREATE TRIGGER update_zp_message_updated_at
    BEFORE UPDATE ON "zpMessage"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- =====================================================
-- Chatwoot Configuration Table
-- =====================================================
CREATE TABLE IF NOT EXISTS "zpChatwootConfig" (
    "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "sessionId" UUID NOT NULL REFERENCES "zpSessions"("id") ON DELETE CASCADE,
    "baseUrl" VARCHAR(512) NOT NULL,
    "accessToken" VARCHAR(255) NOT NULL,
    "accountId" INTEGER NOT NULL,
    "inboxId" INTEGER NOT NULL,
    "enabled" BOOLEAN NOT NULL DEFAULT true,
    "autoCreateContacts" BOOLEAN NOT NULL DEFAULT true,
    "autoCreateConversations" BOOLEAN NOT NULL DEFAULT true,
    "syncHistoryDays" INTEGER DEFAULT 7,
    "createdAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    "updatedAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Chatwoot config indexes
CREATE INDEX IF NOT EXISTS "idx_zp_chatwoot_config_session_id" ON "zpChatwootConfig" ("sessionId");
CREATE INDEX IF NOT EXISTS "idx_zp_chatwoot_config_enabled" ON "zpChatwootConfig" ("enabled");
CREATE INDEX IF NOT EXISTS "idx_zp_chatwoot_config_account_id" ON "zpChatwootConfig" ("accountId");
CREATE INDEX IF NOT EXISTS "idx_zp_chatwoot_config_inbox_id" ON "zpChatwootConfig" ("inboxId");

-- Unique constraint - one config per session
CREATE UNIQUE INDEX IF NOT EXISTS "idx_zp_chatwoot_config_unique_session" ON "zpChatwootConfig" ("sessionId");

-- Chatwoot config trigger
CREATE TRIGGER update_zp_chatwoot_config_updated_at
    BEFORE UPDATE ON "zpChatwootConfig"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- =====================================================
-- Table Comments for Documentation
-- =====================================================

-- Sessions table comments
COMMENT ON TABLE "zpSessions" IS 'WhatsApp sessions management - Clean Architecture implementation';
COMMENT ON COLUMN "zpSessions"."id" IS 'Unique session identifier';
COMMENT ON COLUMN "zpSessions"."name" IS 'Human-readable session name (unique, URL-friendly)';
COMMENT ON COLUMN "zpSessions"."deviceJid" IS 'WhatsApp device JID identifier';
COMMENT ON COLUMN "zpSessions"."isConnected" IS 'Boolean indicating if session is currently connected';
COMMENT ON COLUMN "zpSessions"."connectionError" IS 'Last connection error message if any';
COMMENT ON COLUMN "zpSessions"."qrCode" IS 'Current QR code for session pairing';
COMMENT ON COLUMN "zpSessions"."qrCodeExpiresAt" IS 'QR code expiration timestamp';
COMMENT ON COLUMN "zpSessions"."proxyConfig" IS 'Proxy configuration in JSON format';

-- Webhooks table comments
COMMENT ON TABLE "zpWebhooks" IS 'Webhook configurations for event notifications';
COMMENT ON COLUMN "zpWebhooks"."sessionId" IS 'Associated session ID (NULL for global webhooks)';
COMMENT ON COLUMN "zpWebhooks"."url" IS 'Webhook endpoint URL';
COMMENT ON COLUMN "zpWebhooks"."secret" IS 'Optional webhook secret for verification';
COMMENT ON COLUMN "zpWebhooks"."events" IS 'Array of subscribed event types';

-- Messages table comments
COMMENT ON TABLE "zpMessage" IS 'WhatsApp <-> Chatwoot message mapping table';
COMMENT ON COLUMN "zpMessage"."zpMessageId" IS 'WhatsApp message ID from whatsmeow';
COMMENT ON COLUMN "zpMessage"."zpSender" IS 'WhatsApp sender JID';
COMMENT ON COLUMN "zpMessage"."zpChat" IS 'WhatsApp chat JID (individual or group)';
COMMENT ON COLUMN "zpMessage"."zpType" IS 'WhatsApp message type (text, image, audio, etc.)';
COMMENT ON COLUMN "zpMessage"."syncStatus" IS 'Synchronization status with Chatwoot';

-- Chatwoot config table comments
COMMENT ON TABLE "zpChatwootConfig" IS 'Chatwoot integration configuration per session';
COMMENT ON COLUMN "zpChatwootConfig"."baseUrl" IS 'Chatwoot instance base URL';
COMMENT ON COLUMN "zpChatwootConfig"."accessToken" IS 'Chatwoot API access token';
COMMENT ON COLUMN "zpChatwootConfig"."accountId" IS 'Chatwoot account ID';
COMMENT ON COLUMN "zpChatwootConfig"."inboxId" IS 'Chatwoot inbox ID for WhatsApp integration';