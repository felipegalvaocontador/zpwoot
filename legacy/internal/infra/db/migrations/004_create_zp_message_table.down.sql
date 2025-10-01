-- Drop zpMessage table and related objects
DROP TRIGGER IF EXISTS update_zp_message_updated_at ON "zpMessage";
DROP INDEX IF EXISTS "idx_zp_message_unique_zp";
DROP INDEX IF EXISTS "idx_zp_message_cw_conversation_status";
DROP INDEX IF EXISTS "idx_zp_message_session_chat";
DROP INDEX IF EXISTS "idx_zp_message_created_at";
DROP INDEX IF EXISTS "idx_zp_message_zp_from_me";
DROP INDEX IF EXISTS "idx_zp_message_zp_type";
DROP INDEX IF EXISTS "idx_zp_message_timestamp";
DROP INDEX IF EXISTS "idx_zp_message_sync_status";
DROP INDEX IF EXISTS "idx_zp_message_cw_conversation_id";
DROP INDEX IF EXISTS "idx_zp_message_cw_message_id";
DROP INDEX IF EXISTS "idx_zp_message_zp_chat";
DROP INDEX IF EXISTS "idx_zp_message_zp_message_id";
DROP INDEX IF EXISTS "idx_zp_message_session_id";
DROP TABLE IF EXISTS "zpMessage";
