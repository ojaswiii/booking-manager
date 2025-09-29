-- Rollback bookings table
DROP TRIGGER IF EXISTS update_bookings_updated_at ON bookings;
DROP INDEX IF EXISTS idx_bookings_user_id;
DROP INDEX IF EXISTS idx_bookings_event_id;
DROP INDEX IF EXISTS idx_bookings_status;
DROP INDEX IF EXISTS idx_bookings_expires_at;
DROP INDEX IF EXISTS idx_bookings_user_status;
DROP TABLE IF EXISTS bookings;
