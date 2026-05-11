-- Migration: Add Pluggy keys to users table
-- Date: 2026-05-11

ALTER TABLE users ADD COLUMN pluggy_client_id TEXT;
ALTER TABLE users ADD COLUMN pluggy_client_secret_encrypted TEXT;
