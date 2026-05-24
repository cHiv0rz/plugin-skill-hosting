-- Admin role.
--
-- User-management endpoints (list, approve, reject, delete, promote, demote)
-- are gated to is_admin = TRUE. The default for new rows is FALSE; the very
-- first registered user is bootstrapped to admin both in this migration (for
-- existing deployments) and in the application layer (for fresh installs and
-- the OIDC create path). Admins can promote/demote other admins via the
-- /users page; the first admin is "elected" purely by creation order.

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS is_admin BOOLEAN NOT NULL DEFAULT FALSE;

-- Existing deployments: anoint the earliest-created user as admin so the
-- /users page remains operable after this migration. Only fires when no
-- admin exists yet, so re-running the migration is a no-op.
UPDATE users SET is_admin = TRUE
WHERE id = (
    SELECT id FROM users
    WHERE NOT EXISTS (SELECT 1 FROM users WHERE is_admin)
    ORDER BY created_at ASC, id ASC
    LIMIT 1
);
