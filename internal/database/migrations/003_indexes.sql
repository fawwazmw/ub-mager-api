-- Migration: 003_indexes

CREATE INDEX IF NOT EXISTS idx_rides_passenger_id ON rides(passenger_id);
CREATE INDEX IF NOT EXISTS idx_rides_driver_id ON rides(driver_id);
CREATE INDEX IF NOT EXISTS idx_rides_status ON rides(status);
CREATE INDEX IF NOT EXISTS idx_rides_requested_at ON rides(requested_at DESC);
CREATE INDEX IF NOT EXISTS idx_rides_completed_at ON rides(completed_at DESC) WHERE completed_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_rides_status_requested ON rides(status, requested_at) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_ratings_ride_id ON ratings(ride_id);
CREATE INDEX IF NOT EXISTS idx_ratings_ratee_id ON ratings(ratee_id);
CREATE INDEX IF NOT EXISTS idx_ratings_rater_ride ON ratings(rater_id, ride_id);

CREATE INDEX IF NOT EXISTS idx_driver_profiles_online ON driver_profiles(is_online) WHERE is_online = TRUE AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_driver_profiles_verified ON driver_profiles(is_verified) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);
CREATE INDEX IF NOT EXISTS idx_driver_profiles_deleted_at ON driver_profiles(deleted_at);
CREATE INDEX IF NOT EXISTS idx_rides_deleted_at ON rides(deleted_at);
