-- Migration: 006_trigram_search
-- Enable pg_trgm for fast ILIKE pattern matching

CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Users: search by name, phone, email
CREATE INDEX IF NOT EXISTS idx_users_name_trgm ON users USING gin (full_name gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_users_phone_trgm ON users USING gin (phone gin_trgm_ops);

-- Rides: search by address
CREATE INDEX IF NOT EXISTS idx_rides_pickup_trgm ON rides USING gin (pickup_address gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_rides_dropoff_trgm ON rides USING gin (dropoff_address gin_trgm_ops);

-- Tasks: search by title and description
CREATE INDEX IF NOT EXISTS idx_tasks_title_trgm ON tasks USING gin (title gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_tasks_desc_trgm ON tasks USING gin (description gin_trgm_ops);

-- Driver profiles: search by license plate
CREATE INDEX IF NOT EXISTS idx_drivers_plate_trgm ON driver_profiles USING gin (license_plate gin_trgm_ops);
