-- Up
ALTER TABLE connected_accounts DROP CONSTRAINT IF EXISTS connected_accounts_pluggy_account_id_key;
CREATE UNIQUE INDEX IF NOT EXISTS idx_connected_accounts_user_pluggy_account
    ON connected_accounts(user_id, pluggy_account_id);
ALTER TABLE transactions DROP CONSTRAINT IF EXISTS transactions_pluggy_transaction_id_key;
CREATE UNIQUE INDEX IF NOT EXISTS idx_transactions_account_pluggy_transaction
    ON transactions(account_id, pluggy_transaction_id);

-- Down (execute manually only after confirming there are no cross-user duplicates)
-- DROP INDEX IF EXISTS idx_connected_accounts_user_pluggy_account;
-- ALTER TABLE connected_accounts ADD CONSTRAINT connected_accounts_pluggy_account_id_key UNIQUE (pluggy_account_id);
-- DROP INDEX IF EXISTS idx_transactions_account_pluggy_transaction;
-- ALTER TABLE transactions ADD CONSTRAINT transactions_pluggy_transaction_id_key UNIQUE (pluggy_transaction_id);
