DROP INDEX IF EXISTS idx_accounts_user_id;
ALTER TABLE accounts
DROP COLUMN user_id;
