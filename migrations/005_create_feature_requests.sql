-- Optional manual migration (startup also runs CREATE TABLE IF NOT EXISTS in database.go).
CREATE TABLE IF NOT EXISTS feature_requests (
	id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::TEXT,
	title TEXT NOT NULL,
	description TEXT NOT NULL,
	category TEXT,
	created_by TEXT NOT NULL,
	status TEXT NOT NULL DEFAULT 'pending',
	created_at TIMESTAMP NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
	CONSTRAINT feature_requests_status_check CHECK (
		status IN ('pending', 'reviewing', 'planned', 'completed', 'rejected')
	)
);

CREATE INDEX IF NOT EXISTS idx_feature_requests_created_by ON feature_requests (created_by);
CREATE INDEX IF NOT EXISTS idx_feature_requests_status ON feature_requests (status);
CREATE INDEX IF NOT EXISTS idx_feature_requests_category ON feature_requests (category);
CREATE INDEX IF NOT EXISTS idx_feature_requests_created_at ON feature_requests (created_at DESC);
