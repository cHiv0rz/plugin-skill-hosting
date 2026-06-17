-- Allow the 'move' action in skill version history.
--
-- Moving a skill from one plugin to another records a version entry with
-- action='move'. The original CHECK constraint from 0004 only permitted
-- create/update/delete/restore/revert, so the insert failed (SQLSTATE 23514).
-- Drop and re-add the constraint with 'move' included; the name is the
-- Postgres-default for the inline CHECK in 0004.

ALTER TABLE skill_versions DROP CONSTRAINT IF EXISTS skill_versions_action_check;
ALTER TABLE skill_versions ADD CONSTRAINT skill_versions_action_check
    CHECK (action IN ('create','update','delete','restore','revert','move'));
