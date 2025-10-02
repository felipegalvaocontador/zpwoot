-- Create webhooks table
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

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS "idx_zp_webhooks_session_id" ON "zpWebhooks" ("sessionId");
CREATE INDEX IF NOT EXISTS "idx_zp_webhooks_enabled" ON "zpWebhooks" ("enabled");
CREATE INDEX IF NOT EXISTS "idx_zp_webhooks_created_at" ON "zpWebhooks" ("createdAt");

-- Create trigger to automatically update updatedAt
CREATE TRIGGER update_zp_webhooks_updated_at
    BEFORE UPDATE ON "zpWebhooks"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Add comments for documentation
COMMENT ON TABLE "zpWebhooks" IS 'Webhook configurations for sessions';
COMMENT ON COLUMN "zpWebhooks"."id" IS 'Unique webhook identifier';
COMMENT ON COLUMN "zpWebhooks"."sessionId" IS 'Associated session ID (NULL for global webhooks)';
COMMENT ON COLUMN "zpWebhooks"."url" IS 'Webhook endpoint URL';
COMMENT ON COLUMN "zpWebhooks"."secret" IS 'Optional webhook secret for verification';
COMMENT ON COLUMN "zpWebhooks"."events" IS 'Array of subscribed event types';
COMMENT ON COLUMN "zpWebhooks"."enabled" IS 'Whether webhook is enabled by user';
COMMENT ON COLUMN "zpWebhooks"."createdAt" IS 'Webhook creation timestamp';
COMMENT ON COLUMN "zpWebhooks"."updatedAt" IS 'Last update timestamp';
