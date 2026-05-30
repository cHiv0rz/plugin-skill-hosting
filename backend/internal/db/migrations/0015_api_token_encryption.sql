-- Stop storing API tokens in plaintext.
--
-- The API token gates git / MCP / marketplace, so it must be verifiable on
-- every request AND re-displayable in the UI. AES-GCM ciphertext is
-- non-deterministic and can't be looked up, so we keep two columns:
--   api_token_hash — SHA-256, deterministic, used for authentication lookups
--   api_token_enc  — AES-256-GCM ciphertext, used only to re-display the token
-- Authentication depends solely on the hash, so the encryption key can change
-- (or be unavailable) without locking anyone out — only display is affected.
--
-- The hash is backfilled here in SQL (pgcrypto digest, already enabled in
-- 0001). The ciphertext is filled in by the application at startup because it
-- needs the app key, and that step then clears the plaintext column. The
-- plaintext column's NOT NULL + UNIQUE constraints are relaxed so new accounts
-- never write a plaintext token; the empty column is dropped in a later
-- migration once every deployment has run the backfill at least once.
ALTER TABLE users ADD COLUMN IF NOT EXISTS api_token_hash TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS api_token_enc TEXT;

UPDATE users
SET api_token_hash = encode(digest(api_token, 'sha256'), 'hex')
WHERE api_token_hash IS NULL AND api_token IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS users_api_token_hash_idx ON users(api_token_hash);

ALTER TABLE users ALTER COLUMN api_token DROP NOT NULL;
DROP INDEX IF EXISTS users_api_token_idx;
