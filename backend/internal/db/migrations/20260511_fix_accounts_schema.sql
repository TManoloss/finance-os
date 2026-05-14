-- backend/internal/db/migrations/20260511_fix_accounts_schema.sql
ALTER TABLE connected_accounts DROP CONSTRAINT IF EXISTS connected_accounts_pluggy_item_id_key;
ALTER TABLE connected_accounts ADD COLUMN IF NOT EXISTS pluggy_account_id TEXT UNIQUE;
ALTER TABLE connected_accounts ADD COLUMN IF NOT EXISTS subtype TEXT;
