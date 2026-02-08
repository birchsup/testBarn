ALTER TABLE test_run_cases
    ADD COLUMN status TEXT NOT NULL DEFAULT 'not_run',
    ADD COLUMN comment TEXT,
    ADD COLUMN executed_at TIMESTAMP,
    ADD COLUMN executed_by VARCHAR(255),
    ADD CONSTRAINT test_run_cases_status_check CHECK (status IN ('passed', 'failed', 'blocked', 'skipped', 'not_run'));

CREATE INDEX idx_test_run_cases_run_id ON test_run_cases(run_id);
CREATE INDEX idx_test_run_cases_status ON test_run_cases(status);
CREATE INDEX idx_test_runs_created_at ON test_runs(created_at);
