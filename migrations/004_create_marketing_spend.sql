-- Optional manual migration (startup also runs CREATE TABLE IF NOT EXISTS in database.go).
CREATE TABLE IF NOT EXISTS marketing_spend (
	id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::TEXT,
	platform TEXT NOT NULL,
	campaign_name TEXT NOT NULL,
	amount_spent NUMERIC(12,2) NOT NULL,
	start_date DATE NOT NULL,
	end_date DATE,
	orders_generated INTEGER NOT NULL DEFAULT 0,
	revenue_generated NUMERIC(12,2) NOT NULL DEFAULT 0,
	notes TEXT,
	created_by TEXT NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_marketing_spend_created_by ON marketing_spend (created_by);
CREATE INDEX IF NOT EXISTS idx_marketing_spend_start_date ON marketing_spend (start_date);
CREATE INDEX IF NOT EXISTS idx_marketing_spend_platform ON marketing_spend (platform);
