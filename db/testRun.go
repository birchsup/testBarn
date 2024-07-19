package db

import (
	"encoding/json"
	"time"
)

type TestRun struct {
	ID         int             `json:"id"`
	SuiteID    int             `json:"suite_id"`
	RunDetails json.RawMessage `json:"run_details"`
	CreatedAt  time.Time       `json:"created_at"`
}

//func CreateTestRun(suiteID int, runDetails json.RawMessage) ([]TestRun, error) {
//	query := `INSERT INTO test_runs (suite_id, run_details) VALUES ($1, $2) RETURNING id, suite_id, run_details, created_at`
//
//	rows, err := DBPool.Query(context.Background(), query, suiteID, runDetails)
//	if err != nil {
//		return nil, err
//	}
//	defer rows.Close()
//
//	var testRuns []TestRun
//
//	for rows.Next() {
//		var testRun TestRun
//		if err := rows.Scan(&testRun.ID, &testRun.SuiteID, &testRun.RunDetails, &testRun.CreatedAt); err != nil {
//			return nil, err
//		}
//		testRuns = append(testRuns, testRun)
//	}
//
//	if err := rows.Err(); err != nil {
//		return nil, err
//	}
//
//	return testRuns, nil
//}
