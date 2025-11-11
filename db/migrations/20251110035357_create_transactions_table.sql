-- migrate:up
CREATE TABLE IF NOT EXISTS transactions (
	id BIGINT PRIMARY KEY,
	user_id BIGINT NOT NULL,
	amount NUMERIC(18,2) NOT NULL,
	type TEXT NOT NULL CHECK (type IN ('credit','debit')),
	datetime TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_transactions_user_id ON transactions (user_id);
CREATE INDEX IF NOT EXISTS idx_transactions_datetime ON transactions (datetime);
CREATE INDEX IF NOT EXISTS idx_transactions_type ON transactions (type);

-- migrate:down
DROP INDEX IF EXISTS idx_transactions_type;
DROP INDEX IF EXISTS idx_transactions_datetime;
DROP INDEX IF EXISTS idx_transactions_user_id;
DROP TABLE IF EXISTS transactions;
