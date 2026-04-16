DROP TABLE IF EXISTS erp_quality_action_tasks;
DROP TABLE IF EXISTS erp_quality_action_plans;
ALTER TABLE erp_nonconformities DROP COLUMN IF EXISTS cost_impact;
DROP TABLE IF EXISTS erp_nc_origins;
