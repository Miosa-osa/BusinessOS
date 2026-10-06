-- Rollback: Remove CHECK constraints when the optional OSA table exists.
DO $$
BEGIN
  IF to_regclass('osa_generated_apps') IS NOT NULL THEN
    ALTER TABLE osa_generated_apps DROP CONSTRAINT IF EXISTS check_app_status;
    ALTER TABLE osa_generated_apps DROP CONSTRAINT IF EXISTS check_sandbox_status;
    ALTER TABLE osa_generated_apps DROP CONSTRAINT IF EXISTS check_health_status;
  END IF;
END$$;
