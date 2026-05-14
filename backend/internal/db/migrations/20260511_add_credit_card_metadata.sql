-- backend/internal/db/migrations/20260511_add_credit_card_metadata.sql
ALTER TABLE connected_accounts ADD COLUMN IF NOT EXISTS close_day INT DEFAULT 1;
ALTER TABLE connected_accounts ADD COLUMN IF NOT EXISTS due_day INT DEFAULT 10;
