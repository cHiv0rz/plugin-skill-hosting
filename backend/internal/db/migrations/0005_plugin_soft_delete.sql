-- Plugin soft-delete
--
-- Mirrors the skill soft-delete from 0004: a deleted plugin is hidden from
-- listings, the marketplace feed, and the git endpoint, but the row stays in
-- the database so the owner can restore it later.
--
-- The unique constraint on plugins(name) is replaced by a partial index so a
-- soft-deleted name does not block a new plugin (or a restored plugin) from
-- using the same name -- restore will still fail at this index if an active
-- plugin with that name exists.

ALTER TABLE plugins ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
ALTER TABLE plugins ADD COLUMN IF NOT EXISTS deleted_by UUID REFERENCES users(id) ON DELETE SET NULL;

ALTER TABLE plugins DROP CONSTRAINT IF EXISTS plugins_name_key;
CREATE UNIQUE INDEX IF NOT EXISTS plugins_name_active_idx
    ON plugins(name)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS plugins_deleted_at_idx ON plugins(deleted_at);
