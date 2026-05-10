-- Skill audit + history + soft-delete
--
-- skills now tracks who created/last-edited/deleted, and `deleted_at` makes
-- deletes recoverable. The unique constraint on (plugin_id, name) is replaced
-- by a partial index so a soft-deleted name does not block a new skill with
-- the same name (and so a restored skill collides with an active one).
--
-- skill_versions captures every state transition (create/update/delete/restore)
-- so any prior body+description can be reverted to.

ALTER TABLE skills ADD COLUMN IF NOT EXISTS created_by UUID REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE skills ADD COLUMN IF NOT EXISTS updated_by UUID REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE skills ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
ALTER TABLE skills ADD COLUMN IF NOT EXISTS deleted_by UUID REFERENCES users(id) ON DELETE SET NULL;

-- Backfill creator/editor for any pre-existing rows: fall back to plugin owner.
UPDATE skills s
SET created_by = COALESCE(s.created_by, p.owner_id),
    updated_by = COALESCE(s.updated_by, p.owner_id)
FROM plugins p
WHERE p.id = s.plugin_id;

-- Replace the strict unique constraint with a partial index so soft-deleted
-- rows do not occupy the (plugin_id, name) namespace.
ALTER TABLE skills DROP CONSTRAINT IF EXISTS skills_plugin_id_name_key;
CREATE UNIQUE INDEX IF NOT EXISTS skills_plugin_name_active_idx
    ON skills(plugin_id, name)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS skills_deleted_at_idx ON skills(deleted_at);

CREATE TABLE IF NOT EXISTS skill_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    skill_id UUID NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
    version INT NOT NULL,
    action TEXT NOT NULL CHECK (action IN ('create','update','delete','restore','revert')),
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    body TEXT NOT NULL DEFAULT '',
    edited_by UUID REFERENCES users(id) ON DELETE SET NULL,
    edited_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(skill_id, version)
);

CREATE INDEX IF NOT EXISTS skill_versions_skill_idx ON skill_versions(skill_id, version);

-- Seed a "create" version for any skills that exist before this migration so
-- the history view never shows an empty timeline for legacy data.
INSERT INTO skill_versions (skill_id, version, action, name, description, body, edited_by, edited_at)
SELECT s.id, 1, 'create', s.name, s.description, s.body, s.created_by, s.created_at
FROM skills s
WHERE NOT EXISTS (SELECT 1 FROM skill_versions v WHERE v.skill_id = s.id);
