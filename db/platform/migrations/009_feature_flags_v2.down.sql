ALTER TABLE feature_flags
  DROP COLUMN IF EXISTS rollout_pct,
  DROP COLUMN IF EXISTS updated_at,
  DROP COLUMN IF EXISTS updated_by;
