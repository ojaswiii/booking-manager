-- Rollback tickets table
DROP TRIGGER IF EXISTS update_tickets_updated_at ON tickets;
DROP INDEX IF EXISTS idx_tickets_event_id;
DROP INDEX IF EXISTS idx_tickets_status;
DROP INDEX IF EXISTS idx_tickets_event_status;
DROP TABLE IF EXISTS tickets;
