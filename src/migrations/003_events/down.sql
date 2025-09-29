-- Rollback events table
DROP TRIGGER IF EXISTS update_events_updated_at ON events;
DROP INDEX IF EXISTS idx_events_date;
DROP INDEX IF EXISTS idx_events_artist;
DROP TABLE IF EXISTS events;
