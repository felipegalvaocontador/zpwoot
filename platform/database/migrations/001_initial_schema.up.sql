-- =====================================================
-- zpwoot Database Schema - Initial Migration
-- Clean Architecture Implementation
-- Based on legacy migrations: 001-004
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

-- Create indexes for better performance
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

-- Sessions table comments
COMMENT ON TABLE "zpSessions" IS 'Wameow sessions management table - optimized with boolean connection status';
COMMENT ON COLUMN "zpSessions"."id" IS 'Unique session identifier';
COMMENT ON COLUMN "zpSessions"."name" IS 'Human-readable session name (unique, URL-friendly)';
COMMENT ON COLUMN "zpSessions"."deviceJid" IS 'Wameow device JID identifier';
COMMENT ON COLUMN "zpSessions"."isConnected" IS 'Boolean indicating if session is currently connected to Wameow';
COMMENT ON COLUMN "zpSessions"."connectionError" IS 'Last connection error message if any';
COMMENT ON COLUMN "zpSessions"."qrCode" IS 'Current QR code for session pairing';
COMMENT ON COLUMN "zpSessions"."qrCodeExpiresAt" IS 'QR code expiration timestamp';
COMMENT ON COLUMN "zpSessions"."proxyConfig" IS 'Proxy configuration in JSON format';
COMMENT ON COLUMN "zpSessions"."createdAt" IS 'Session creation timestamp';
COMMENT ON COLUMN "zpSessions"."updatedAt" IS 'Last update timestamp';
COMMENT ON COLUMN "zpSessions"."connectedAt" IS 'Last successful connection timestamp';
COMMENT ON COLUMN "zpSessions"."lastSeen" IS 'Last activity timestamp';

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

-- Webhooks table comments
COMMENT ON TABLE "zpWebhooks" IS 'Webhook configurations for sessions';
COMMENT ON COLUMN "zpWebhooks"."id" IS 'Unique webhook identifier';
COMMENT ON COLUMN "zpWebhooks"."sessionId" IS 'Associated session ID (NULL for global webhooks)';
COMMENT ON COLUMN "zpWebhooks"."url" IS 'Webhook endpoint URL';
COMMENT ON COLUMN "zpWebhooks"."secret" IS 'Optional webhook secret for verification';
COMMENT ON COLUMN "zpWebhooks"."events" IS 'Array of subscribed event types';
COMMENT ON COLUMN "zpWebhooks"."enabled" IS 'Whether webhook is enabled by user';
COMMENT ON COLUMN "zpWebhooks"."createdAt" IS 'Webhook creation timestamp';
COMMENT ON COLUMN "zpWebhooks"."updatedAt" IS 'Last update timestamp';

-- =====================================================
-- Chatwoot Configuration Table
-- =====================================================
CREATE TABLE IF NOT EXISTS "zpChatwoot" (
    "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "sessionId" UUID NOT NULL REFERENCES "zpSessions"("id") ON DELETE CASCADE,
    "url" VARCHAR(2048) NOT NULL,
    "token" VARCHAR(255) NOT NULL,
    "accountId" VARCHAR(50) NOT NULL,
    "inboxId" VARCHAR(50),
    "enabled" BOOLEAN NOT NULL DEFAULT true,

    -- Advanced configuration with shorter names
    "inboxName" VARCHAR(255),
    "autoCreate" BOOLEAN DEFAULT false,
    "signMsg" BOOLEAN DEFAULT false,
    "signDelimiter" VARCHAR(50) DEFAULT E'\n\n',
    "reopenConv" BOOLEAN DEFAULT true,
    "convPending" BOOLEAN DEFAULT false,
    "importContacts" BOOLEAN DEFAULT false,
    "importMessages" BOOLEAN DEFAULT false,
    "importDays" INTEGER DEFAULT 60,
    "mergeBrazil" BOOLEAN DEFAULT true,
    "organization" VARCHAR(255),
    "logo" VARCHAR(2048),
    "number" VARCHAR(20),
    "ignoreJids" TEXT[],

    "createdAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    "updatedAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Chatwoot indexes
CREATE INDEX IF NOT EXISTS "idx_zp_chatwoot_session_id" ON "zpChatwoot" ("sessionId");
CREATE INDEX IF NOT EXISTS "idx_zp_chatwoot_enabled" ON "zpChatwoot" ("enabled");
CREATE INDEX IF NOT EXISTS "idx_zp_chatwoot_created_at" ON "zpChatwoot" ("createdAt");
CREATE INDEX IF NOT EXISTS "idx_zp_chatwoot_auto_create" ON "zpChatwoot" ("autoCreate");
CREATE INDEX IF NOT EXISTS "idx_zp_chatwoot_inbox_name" ON "zpChatwoot" ("inboxName");
CREATE INDEX IF NOT EXISTS "idx_zp_chatwoot_number" ON "zpChatwoot" ("number");

-- Unique constraint: one Chatwoot config per session
CREATE UNIQUE INDEX IF NOT EXISTS "idx_zp_chatwoot_unique_session" ON "zpChatwoot" ("sessionId");

-- Chatwoot trigger
CREATE TRIGGER update_zp_chatwoot_updated_at
    BEFORE UPDATE ON "zpChatwoot"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Chatwoot table comments
COMMENT ON TABLE "zpChatwoot" IS 'Chatwoot integration configuration - one per session';
COMMENT ON COLUMN "zpChatwoot"."id" IS 'Unique configuration identifier';
COMMENT ON COLUMN "zpChatwoot"."sessionId" IS 'Reference to WhatsApp session (one-to-one)';
COMMENT ON COLUMN "zpChatwoot"."url" IS 'Chatwoot instance URL';
COMMENT ON COLUMN "zpChatwoot"."token" IS 'Chatwoot user token';
COMMENT ON COLUMN "zpChatwoot"."accountId" IS 'Chatwoot account ID';
COMMENT ON COLUMN "zpChatwoot"."inboxId" IS 'Optional Chatwoot inbox ID';
COMMENT ON COLUMN "zpChatwoot"."enabled" IS 'Whether configuration is enabled';
COMMENT ON COLUMN "zpChatwoot"."inboxName" IS 'Custom name for Chatwoot inbox';
COMMENT ON COLUMN "zpChatwoot"."autoCreate" IS 'Auto-create inbox and setup integration';
COMMENT ON COLUMN "zpChatwoot"."signMsg" IS 'Add signature to messages';
COMMENT ON COLUMN "zpChatwoot"."signDelimiter" IS 'Delimiter for message signature';
COMMENT ON COLUMN "zpChatwoot"."reopenConv" IS 'Reopen resolved conversations on new message';
COMMENT ON COLUMN "zpChatwoot"."convPending" IS 'Set new conversations as pending';
COMMENT ON COLUMN "zpChatwoot"."importContacts" IS 'Import WhatsApp contacts to Chatwoot';
COMMENT ON COLUMN "zpChatwoot"."importMessages" IS 'Import message history to Chatwoot';
COMMENT ON COLUMN "zpChatwoot"."importDays" IS 'Days limit for message import (default: 60)';
COMMENT ON COLUMN "zpChatwoot"."mergeBrazil" IS 'Merge Brazilian contacts (+55)';
COMMENT ON COLUMN "zpChatwoot"."organization" IS 'Organization name for bot contact';
COMMENT ON COLUMN "zpChatwoot"."logo" IS 'Logo URL for bot contact';
COMMENT ON COLUMN "zpChatwoot"."number" IS 'WhatsApp number for this integration';
COMMENT ON COLUMN "zpChatwoot"."ignoreJids" IS 'Array of JIDs to ignore in sync';
COMMENT ON COLUMN "zpChatwoot"."createdAt" IS 'Configuration creation timestamp';
COMMENT ON COLUMN "zpChatwoot"."updatedAt" IS 'Last update timestamp';

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

-- Messages table comments
COMMENT ON TABLE "zpMessage" IS 'Simple mapping table between WhatsApp messages and Chatwoot messages';
COMMENT ON COLUMN "zpMessage"."id" IS 'Unique message mapping identifier';
COMMENT ON COLUMN "zpMessage"."sessionId" IS 'WhatsApp session identifier';
COMMENT ON COLUMN "zpMessage"."zpMessageId" IS 'WhatsApp message ID from whatsmeow';
COMMENT ON COLUMN "zpMessage"."zpSender" IS 'WhatsApp sender JID';
COMMENT ON COLUMN "zpMessage"."zpChat" IS 'WhatsApp chat JID (individual or group)';
COMMENT ON COLUMN "zpMessage"."zpTimestamp" IS 'WhatsApp message timestamp';
COMMENT ON COLUMN "zpMessage"."zpFromMe" IS 'Whether message was sent by me (true) or received (false)';
COMMENT ON COLUMN "zpMessage"."zpType" IS 'WhatsApp message type (text, image, audio, video, document, contact, etc.)';
COMMENT ON COLUMN "zpMessage"."content" IS 'Message text content';
COMMENT ON COLUMN "zpMessage"."cwMessageId" IS 'Chatwoot message ID';
COMMENT ON COLUMN "zpMessage"."cwConversationId" IS 'Chatwoot conversation ID';
COMMENT ON COLUMN "zpMessage"."syncStatus" IS 'Synchronization status with Chatwoot';
COMMENT ON COLUMN "zpMessage"."createdAt" IS 'Record creation timestamp';
COMMENT ON COLUMN "zpMessage"."updatedAt" IS 'Last update timestamp';
