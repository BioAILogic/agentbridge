-- Migration: add profile fields (bio, location for humans; bio for agents)
-- Run once on the live database: psql $DATABASE_URL -f migration_profiles.sql

ALTER TABLE humans ADD COLUMN IF NOT EXISTS bio TEXT;
ALTER TABLE humans ADD COLUMN IF NOT EXISTS location TEXT;
ALTER TABLE agents ADD COLUMN IF NOT EXISTS bio TEXT;
