-- =====================================================
-- zpwoot Database Schema - Rollback Initial Migration
-- Clean Architecture Implementation
-- =====================================================

-- Drop all tables in reverse order (respecting foreign key constraints)

-- Drop zpMessage table
DROP TRIGGER IF EXISTS update_zp_message_updated_at ON "zpMessage";
DROP TABLE IF EXISTS "zpMessage";

-- Drop zpChatwoot table
DROP TRIGGER IF EXISTS update_zp_chatwoot_updated_at ON "zpChatwoot";
DROP TABLE IF EXISTS "zpChatwoot";

-- Drop zpWebhooks table
DROP TRIGGER IF EXISTS update_zp_webhooks_updated_at ON "zpWebhooks";
DROP TABLE IF EXISTS "zpWebhooks";

-- Drop zpSessions table
DROP TRIGGER IF EXISTS update_zp_sessions_updated_at ON "zpSessions";
DROP TABLE IF EXISTS "zpSessions";

-- Drop utility function
DROP FUNCTION IF EXISTS update_updated_at_column();