package db

import (
	"context"
	"encoding/json"
	"log"
	"time"
)

type TestCaseRun struct {
	ID        int64      `json:"id"`
	Name      string     `json:"name"`
	Details   string     `json:"details"`
	Test      string     `json:"test"`
	Status    *string    `json:"status,omitempty"` // nil -> null в JSON
	Comment   *string    `json:"comment,omitempty"`
	SuiteID   *int64     `json:"suite_id,omitempty"`
	SuiteName *string    `json:"suite_name,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
}

func GetAllTestCasesFromRun(runID int64) ([]TestCaseRun, error) {
	query := `
		SELECT DISTINCT 
			tc.id, 
			COALESCE(tc.test ->> 'name', '') AS test_case_name, 
			trc.status, 
			trc.comment, 
			trs.suite_id
		FROM test_cases tc
		LEFT JOIN test_suite_cases tsc ON tc.id = tsc.case_id
		LEFT JOIN test_run_suites trs ON tsc.suite_id = trs.suite_id
		LEFT JOIN test_run_cases trc ON tc.id = trc.case_id AND (trs.run_id = trc.run_id OR trc.run_id = $1)
		WHERE trs.run_id = $1 OR trc.run_id = $1;
	`

	rows, err := DBPool.Query(context.Background(), query, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var testCasesRun []TestCaseRun
	for rows.Next() {
		var tc TestCaseRun
		if err := rows.Scan(&tc.ID, &tc.Name, &tc.Status, &tc.Comment, &tc.SuiteID); err != nil {
			return nil, err
		}

		if tc.Name == "" {
			log.Printf("Skipping test case %d: missing name", tc.ID)
			continue
		}

		testCasesRun = append(testCasesRun, tc)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return testCasesRun, nil
}

func GetTestCaseDetails(runID, caseID int64) (TestCaseRun, error) {
	query := `
		SELECT tc.id, tc.test  ->> 'name' AS test_case_name, trc.status, trc.comment, trs.suite_id, ts.name, tr.created_at
		FROM test_cases tc
		LEFT JOIN test_run_cases trc ON tc.id = trc.case_id
		LEFT JOIN test_run_suites trs ON tc.id = trs.suite_id
		LEFT JOIN test_suites ts ON trs.suite_id = ts.id
		LEFT JOIN test_runs tr ON trc.run_id = tr.id
		WHERE trc.run_id = $1 AND tc.id = $2;
	`

	var tcr TestCaseRun
	err := DBPool.QueryRow(context.Background(), query, runID, caseID).Scan(
		&tcr.ID, &tcr.Test, &tcr.Status, &tcr.Comment, &tcr.SuiteID, &tcr.SuiteName, &tcr.CreatedAt,
	)
	if err != nil {
		return TestCaseRun{}, err
	}
	return tcr, nil
}

//func AddTestSuiteToRun(runID, suiteID int64) error {
//	_, err := DBPool.Exec(context.Background(), "INSERT INTO test_run_suites (run_id, suite_id) VALUES ($1, $2)", runID, suiteID)
//	return err
//}

func AddTestSuiteToRun(runID, suiteID int64) error {
	ctx := context.Background()
	_, err := DBPool.Exec(ctx, `
		WITH inserted_suite AS (
			INSERT INTO test_run_suites (run_id, suite_id)
			VALUES ($1, $2)
			RETURNING run_id, suite_id
		)
		INSERT INTO test_run_cases (run_id, case_id)
		SELECT inserted_suite.run_id, tsc.case_id
		FROM inserted_suite
		JOIN test_suite_cases tsc ON inserted_suite.suite_id = tsc.suite_id
		ON CONFLICT (run_id, case_id) DO NOTHING;
	`, runID, suiteID)

	return err
}

func AddAllTestCasesFromSuiteToRun(runID int64) error {
	query := `
		INSERT INTO test_run_cases (run_id, case_id)
		SELECT trs.run_id, tsc.case_id
		FROM test_run_suites trs
		JOIN test_suite_cases tsc ON trs.suite_id = tsc.suite_id
		WHERE trs.run_id = $1
		ON CONFLICT (run_id, case_id) DO NOTHING;
	`
	_, err := DBPool.Exec(context.Background(), query, runID)
	return err
}

func UpdateTestCaseStatus(runID, caseID int64, status, comment string) error {
	query := `
		UPDATE test_run_cases
		SET status = $1, comment = $2
		WHERE run_id = $3 AND case_id = $4;
	`

	_, err := DBPool.Exec(context.Background(), query, status, comment, runID, caseID)
	if err != nil {
		log.Printf("Error updating status of test case %d in run %d: %v", caseID, runID, err)
		return err
	}

	log.Printf("Successfully updated test case %d in run %d with status: %s", caseID, runID, status)
	return nil
}

func CreateTestRun(details json.RawMessage) (int64, error) {
	query := `
		INSERT INTO test_runs (run_details)
		VALUES ($1)
		RETURNING id;
	`

	var runID int64
	err := DBPool.QueryRow(context.Background(), query, details).Scan(&runID)
	if err != nil {
		return 0, err
	}
	return runID, nil
}

func GetAllTestRuns() ([]TestCaseRun, error) {
	query := `
	SELECT id, COALESCE(run_details->>'details', '') AS details, created_at 
	FROM test_runs
	ORDER BY created_at DESC;
	`

	rows, err := DBPool.Query(context.Background(), query)
	if err != nil {
		log.Printf("Error executing query: %v", err)
		return nil, err
	}
	defer rows.Close()

	var testRuns []TestCaseRun
	for rows.Next() {
		var tr TestCaseRun
		if err := rows.Scan(&tr.ID, &tr.Details, &tr.CreatedAt); err != nil {
			log.Printf("Error scanning row: %v", err)
			return nil, err
		}
		testRuns = append(testRuns, tr)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating rows: %v", err)
		return nil, err
	}

	return testRuns, nil
}
