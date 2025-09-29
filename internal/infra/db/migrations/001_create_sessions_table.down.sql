-- Drop sessions table and related objects
DROP TRIGGER IF EXISTS update_zp_sessions_updated_at ON "zpSessions";
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP TABLE IF EXISTS "zpSessions";
