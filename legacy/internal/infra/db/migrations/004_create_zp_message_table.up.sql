-- Create zpMessage table for WhatsApp <-> Chatwoot message mapping
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

-- Create indexes for better performance
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

-- Create trigger to automatically update updatedAt
CREATE TRIGGER update_zp_message_updated_at
    BEFORE UPDATE ON "zpMessage"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Add comments for documentation
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
