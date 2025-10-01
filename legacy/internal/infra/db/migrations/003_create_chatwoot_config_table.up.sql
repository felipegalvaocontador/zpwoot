-- Create chatwoot_config table with all advanced features
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

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS "idx_zp_chatwoot_session_id" ON "zpChatwoot" ("sessionId");
CREATE INDEX IF NOT EXISTS "idx_zp_chatwoot_enabled" ON "zpChatwoot" ("enabled");
CREATE INDEX IF NOT EXISTS "idx_zp_chatwoot_created_at" ON "zpChatwoot" ("createdAt");
CREATE INDEX IF NOT EXISTS "idx_zp_chatwoot_auto_create" ON "zpChatwoot" ("autoCreate");
CREATE INDEX IF NOT EXISTS "idx_zp_chatwoot_inbox_name" ON "zpChatwoot" ("inboxName");
CREATE INDEX IF NOT EXISTS "idx_zp_chatwoot_number" ON "zpChatwoot" ("number");

-- Unique constraint: one Chatwoot config per session
CREATE UNIQUE INDEX IF NOT EXISTS "idx_zp_chatwoot_unique_session" ON "zpChatwoot" ("sessionId");

-- Create trigger to automatically update updatedAt
CREATE TRIGGER update_zp_chatwoot_updated_at
    BEFORE UPDATE ON "zpChatwoot"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Add comments for documentation
COMMENT ON TABLE "zpChatwoot" IS 'Chatwoot integration configuration - one per session';
COMMENT ON COLUMN "zpChatwoot"."id" IS 'Unique configuration identifier';
COMMENT ON COLUMN "zpChatwoot"."sessionId" IS 'Reference to WhatsApp session (one-to-one)';
COMMENT ON COLUMN "zpChatwoot"."url" IS 'Chatwoot instance URL';
COMMENT ON COLUMN "zpChatwoot"."token" IS 'Chatwoot user token';
COMMENT ON COLUMN "zpChatwoot"."accountId" IS 'Chatwoot account ID';
COMMENT ON COLUMN "zpChatwoot"."inboxId" IS 'Optional Chatwoot inbox ID';
COMMENT ON COLUMN "zpChatwoot"."enabled" IS 'Whether configuration is enabled';

-- Advanced configuration comments
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
