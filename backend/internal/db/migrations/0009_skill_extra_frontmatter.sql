-- Skill extra frontmatter
--
-- Some imported SKILL.md files carry YAML frontmatter keys other than
-- name and description (e.g. allowed-tools, license, custom metadata).
-- We preserve them verbatim as raw YAML text so round-trip through the
-- marketplace is lossless: store on import, re-emit on materialization.

ALTER TABLE skills
    ADD COLUMN IF NOT EXISTS extra_frontmatter TEXT NOT NULL DEFAULT '';

ALTER TABLE skill_versions
    ADD COLUMN IF NOT EXISTS extra_frontmatter TEXT NOT NULL DEFAULT '';
