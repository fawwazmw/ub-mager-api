-- Migration: 005_timescaledb_hypertables

CREATE TABLE IF NOT EXISTS driver_location_history (
    driver_id UUID NOT NULL,
    lat DOUBLE PRECISION NOT NULL,
    lng DOUBLE PRECISION NOT NULL,
    speed DOUBLE PRECISION DEFAULT 0,
    heading DOUBLE PRECISION DEFAULT 0,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

SELECT create_hypertable('driver_location_history', 'recorded_at', if_not_exists => TRUE);

CREATE INDEX IF NOT EXISTS idx_driver_loc_history_driver ON driver_location_history(driver_id, recorded_at DESC);
