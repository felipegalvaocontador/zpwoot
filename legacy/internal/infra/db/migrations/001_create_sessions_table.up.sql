-- Create sessions table
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

-- Create trigger to automatically update updatedAt
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW."updatedAt" = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_zp_sessions_updated_at
    BEFORE UPDATE ON "zpSessions"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Add comments for documentation
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
