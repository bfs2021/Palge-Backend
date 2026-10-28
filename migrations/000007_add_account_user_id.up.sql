ALTER TABLE accounts ADD COLUMN user_id bigint REFERENCES users ON DELETE CASCADE;
CREATE INDEX idx_accounts_user_id ON accounts (user_id);