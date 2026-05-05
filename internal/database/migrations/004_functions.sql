-- Migration: 004_functions

CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DO $$ BEGIN
    CREATE TRIGGER trg_users_updated_at
        BEFORE UPDATE ON users
        FOR EACH ROW EXECUTE FUNCTION update_updated_at();
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
    CREATE TRIGGER trg_driver_profiles_updated_at
        BEFORE UPDATE ON driver_profiles
        FOR EACH ROW EXECUTE FUNCTION update_updated_at();
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
    CREATE TRIGGER trg_rides_updated_at
        BEFORE UPDATE ON rides
        FOR EACH ROW EXECUTE FUNCTION update_updated_at();
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

CREATE OR REPLACE FUNCTION update_driver_rating()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE driver_profiles
    SET rating_avg = (
        SELECT COALESCE(AVG(score), 5.0)
        FROM ratings
        WHERE ratee_id = (SELECT user_id FROM driver_profiles WHERE id = NEW.ratee_id)
    )
    WHERE user_id = NEW.ratee_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DO $$ BEGIN
    CREATE TRIGGER trg_update_driver_rating
        AFTER INSERT ON ratings
        FOR EACH ROW EXECUTE FUNCTION update_driver_rating();
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;
