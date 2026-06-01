-- Per-user UI theme preference.
--
-- The frontend ships several palettes (light, dark, midnight, sepia,
-- contrast); this stores each user's choice so it follows them across devices.
-- localStorage remains the fast client-side cache, but the server value wins on
-- login. Validation of the allowed set lives in the application layer so adding
-- a theme doesn't require a migration; the column is a plain TEXT with a
-- 'light' default to match the shipped default.

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS theme TEXT NOT NULL DEFAULT 'light';
