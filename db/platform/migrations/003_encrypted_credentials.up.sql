-- Add encrypted credential columns. The plaintext columns are kept
-- for backwards compatibility during migration. Once all URLs are
-- encrypted, a future migration will drop the plaintext columns.
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS postgres_url_enc TEXT;
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS redis_url_enc TEXT;
