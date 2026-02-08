DROP INDEX IF EXISTS idx_test_runs_created_at;
DROP INDEX IF EXISTS idx_test_run_cases_status;
DROP INDEX IF EXISTS idx_test_run_cases_run_id;

ALTER TABLE test_run_cases
    DROP CONSTRAINT IF EXISTS test_run_cases_status_check,
    DROP COLUMN IF EXISTS executed_by,
    DROP COLUMN IF EXISTS executed_at,
    DROP COLUMN IF EXISTS comment,
    DROP COLUMN IF EXISTS status;
