-- Migration: 001_extensions_and_enums
-- Enable required PostgreSQL extensions and create enum types

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "postgis";
CREATE EXTENSION IF NOT EXISTS "timescaledb" CASCADE;

DO $$ BEGIN
    CREATE TYPE user_role AS ENUM ('PASSENGER', 'DRIVER', 'ADMIN');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
    CREATE TYPE vehicle_type AS ENUM ('MOTORCYCLE', 'CAR', 'CAR_XL');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
    CREATE TYPE ride_status AS ENUM (
        'SEARCHING', 'MATCHED', 'DRIVER_EN_ROUTE',
        'ARRIVED_AT_PICKUP', 'IN_PROGRESS', 'COMPLETED', 'CANCELLED'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
    CREATE TYPE payment_method AS ENUM ('CASH', 'EWALLET');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;
