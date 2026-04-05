-- Stub for external tables referenced by ingest schema.
-- This file exists only so sqlc can resolve foreign key references.
-- The real users table is created by the auth service migration.
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY
);
