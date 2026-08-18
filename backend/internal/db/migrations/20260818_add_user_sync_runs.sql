-- Up
ALTER TABLE sync_logs ADD COLUMN IF NOT EXISTS user_id UUID REFERENCES users(id) ON DELETE CASCADE;
ALTER TABLE sync_logs ADD COLUMN IF NOT EXISTS item_id TEXT;
ALTER TABLE sync_logs ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'completed';
CREATE INDEX IF NOT EXISTS idx_sync_logs_user_started_at ON sync_logs(user_id, started_at DESC);

-- Down (execute manually only after clients stop consulting manual sync runs)
-- DROP INDEX IF EXISTS idx_sync_logs_user_started_at;
-- ALTER TABLE sync_logs DROP COLUMN IF EXISTS status;
-- ALTER TABLE sync_logs DROP COLUMN IF EXISTS item_id;
-- ALTER TABLE sync_logs DROP COLUMN IF EXISTS user_id;
