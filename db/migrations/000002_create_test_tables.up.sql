
CREATE TABLE test_suites (
                             id SERIAL PRIMARY KEY,
                             name VARCHAR(255) NOT NULL,
                             description TEXT,
                             created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE test_suite_cases (
                                  suite_id INTEGER REFERENCES test_suites(id),
                                  case_id INTEGER REFERENCES test_cases(id),
                                  PRIMARY KEY (suite_id, case_id)
);

CREATE TABLE test_runs (
                           id SERIAL PRIMARY KEY,
                           suite_id INTEGER REFERENCES test_suites(id),
                           run_details JSONB NOT NULL,
                           created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE test_run_cases (
                                run_id INTEGER REFERENCES test_runs(id),
                                case_id INTEGER REFERENCES test_cases(id),
                                PRIMARY KEY (run_id, case_id)
);


