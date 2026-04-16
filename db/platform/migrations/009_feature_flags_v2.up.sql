-- Feature flags v2: gradual rollout + audit trail
ALTER TABLE feature_flags
  ADD COLUMN IF NOT EXISTS rollout_pct   INT NOT NULL DEFAULT 100,
  ADD COLUMN IF NOT EXISTS updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  ADD COLUMN IF NOT EXISTS updated_by    TEXT;

COMMENT ON COLUMN feature_flags.rollout_pct IS
  'Percentage of users who see this flag (0-100). Evaluation: hash(flag_id + user_id) mod 100 < rollout_pct (deterministic per user).';
