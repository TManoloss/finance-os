-- backend/internal/db/migrations/20260511_add_institution_metadata.sql
ALTER TABLE connected_accounts ADD COLUMN IF NOT EXISTS institution_logo TEXT;
ALTER TABLE connected_accounts ADD COLUMN IF NOT EXISTS institution_color TEXT;
