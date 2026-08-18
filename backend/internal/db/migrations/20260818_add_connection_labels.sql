-- Up
ALTER TABLE connected_accounts ADD COLUMN IF NOT EXISTS account_name TEXT;
ALTER TABLE connected_accounts ADD COLUMN IF NOT EXISTS account_number_last4 TEXT;
ALTER TABLE connected_accounts ADD COLUMN IF NOT EXISTS connection_label TEXT;

-- Down (execute manually only after clients stop consuming these fields)
-- ALTER TABLE connected_accounts DROP COLUMN IF EXISTS connection_label;
-- ALTER TABLE connected_accounts DROP COLUMN IF EXISTS account_number_last4;
-- ALTER TABLE connected_accounts DROP COLUMN IF EXISTS account_name;
