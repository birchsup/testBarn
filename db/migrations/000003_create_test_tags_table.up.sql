CREATE TABLE test_case_tags (
                                case_id BIGINT REFERENCES test_cases(id),
                                tags JSONB,
                                PRIMARY KEY (case_id)
);

CREATE TABLE test_run_suites (
                                 run_id INT NOT NULL,
                                 suite_id INT NOT NULL,
                                 PRIMARY KEY (run_id, suite_id),
                                 FOREIGN KEY (run_id) REFERENCES test_runs(id) ON DELETE CASCADE,
                                 FOREIGN KEY (suite_id) REFERENCES test_suites(id) ON DELETE CASCADE
);