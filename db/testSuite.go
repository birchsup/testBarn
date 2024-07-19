package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"
)

type TestSuite struct {
	ID          int        `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	CreatedAt   time.Time  `json:"created_at"`
	TestCases   []TestCase `json:"test_cases"`
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

func GetTestSuiteByID(suiteID string) (TestSuite, error) {
	var testSuite TestSuite
	testSuite.TestCases = []TestCase{} // Инициализация пустого среза

	query := `
		SELECT 
			ts.id, ts.name, ts.description, ts.created_at,
			COALESCE(tc.id, 0), COALESCE(tc.test, '{}')
		FROM 
			test_suites ts
		LEFT JOIN 
			test_suite_cases tsc ON ts.id = tsc.suite_id
		LEFT JOIN 
			test_cases tc ON tsc.case_id = tc.id
		WHERE 
			ts.id = $1`

	rows, err := DBPool.Query(context.Background(), query, suiteID)
	if err != nil {
		return testSuite, err
	}
	defer rows.Close()

	found := false
	for rows.Next() {
		found = true
		var testCase TestCase
		var testCaseID sql.NullInt64
		var testCaseTest json.RawMessage

		if err := rows.Scan(&testSuite.ID, &testSuite.Name, &testSuite.Description, &testSuite.CreatedAt, &testCaseID, &testCaseTest); err != nil {
			return testSuite, err
		}

		if testCaseID.Valid {
			testCase.ID = int64(int(testCaseID.Int64))
			testCase.Test = testCaseTest
			testSuite.TestCases = append(testSuite.TestCases, testCase)
		}
	}

	if err := rows.Err(); err != nil {
		return testSuite, err
	}

	if !found {
		return testSuite, sql.ErrNoRows
	}

	return testSuite, nil
}

// UpdateTestSuite updates the details of an existing test suite
func UpdateTestSuite(suiteID int, name string, description string) (TestSuite, error) {
	query := `UPDATE test_suites SET name = $1, description = $2 WHERE id = $3 RETURNING id, name, description`
	var testSuite TestSuite
	err := DBPool.QueryRow(context.Background(), query, name, description, suiteID).Scan(&testSuite.ID, &testSuite.Name, &testSuite.Description)
	if err != nil {
		return TestSuite{}, err
	}
	return testSuite, nil
}

// DeleteTestSuite deletes a test suite and its associated test cases
func DeleteTestSuite(suiteID int) error {
	query := `DELETE FROM test_suites WHERE id = $1`
	_, err := DBPool.Exec(context.Background(), query, suiteID)
	if err != nil {
		return err
	}
	return nil
}

func RemoveTestCaseFromSuite(suiteID int, caseID int) error {
	query := `DELETE FROM test_suite_cases WHERE suite_id = $1 AND case_id = $2`
	_, err := DBPool.Exec(context.Background(), query, suiteID, caseID)
	if err != nil {
		return err
	}
	return nil
}
