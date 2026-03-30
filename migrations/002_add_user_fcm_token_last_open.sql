-- Optional manual migration (startup also runs ADD COLUMN IF NOT EXISTS in database.go).
ALTER TABLE users ADD COLUMN IF NOT EXISTS fcm_token TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_open TIMESTAMP;
