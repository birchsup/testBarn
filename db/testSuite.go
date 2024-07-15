package db

import (
	"context"
	"time"
)

type TestSuite struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type TestSuiteRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type AddTestCaseRequest struct {
	SuiteID int   `json:"suite_id"`
	CaseIDs []int `json:"case_ids"`
}

func CreateTestSuite(name string, description string) (TestSuite, error) {
	query := `INSERT INTO test_suites (name, description) VALUES ($1, $2) RETURNING id, name, description, created_at`
	var testSuite TestSuite
	err := DBPool.QueryRow(context.Background(), query, name, description).Scan(&testSuite.ID, &testSuite.Name, &testSuite.Description, &testSuite.CreatedAt)
	if err != nil {
		return TestSuite{}, err
	}
	return testSuite, nil
}

func AddTestCasesToSuite(suiteID int, caseIDs []int) error {
	for _, caseID := range caseIDs {
		_, err := DBPool.Exec(context.Background(), `INSERT INTO test_suite_cases (suite_id, case_id) VALUES ($1, $2)`, suiteID, caseID)
		if err != nil {
			return err
		}
	}
	return nil
}
