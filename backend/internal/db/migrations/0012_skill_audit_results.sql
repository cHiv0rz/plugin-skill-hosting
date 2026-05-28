-- Scheduled security-audit results. One row per skill per audit run; the
-- latest row per skill (by audited_at) is the current verdict. History is
-- retained so risk trends over time stay visible.

CREATE TABLE IF NOT EXISTS skill_audit_results (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    skill_id    UUID NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
    audited_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    model       TEXT NOT NULL DEFAULT '',
    -- risk_score is 0-100. 0 means "no concern", 100 means "definitely malicious".
    risk_score  INT NOT NULL DEFAULT 0,
    -- risk_level is a coarse bucket derived from risk_score: low|medium|high|critical.
    risk_level  TEXT NOT NULL DEFAULT 'low',
    -- categories is a JSON array of short threat tags, e.g. ["data-exfiltration"].
    categories  JSONB NOT NULL DEFAULT '[]'::jsonb,
    summary     TEXT NOT NULL DEFAULT '',
    -- findings is a JSON array of {category, severity, detail} objects.
    findings    JSONB NOT NULL DEFAULT '[]'::jsonb,
    -- raw holds the model's unparsed reply for debugging prompt/score drift.
    raw         TEXT NOT NULL DEFAULT '',
    -- error is non-empty when the audit call itself failed (API error, parse
    -- failure, skill could not be analyzed). Such rows have risk_score 0 and
    -- are excluded from threshold alerting.
    error       TEXT NOT NULL DEFAULT '',
    -- alerted records whether this result already triggered a notification, so
    -- re-runs don't re-alert on an unchanged verdict within the same run set.
    alerted     BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS idx_skill_audit_results_skill
    ON skill_audit_results (skill_id, audited_at DESC);

CREATE INDEX IF NOT EXISTS idx_skill_audit_results_audited_at
    ON skill_audit_results (audited_at DESC);
