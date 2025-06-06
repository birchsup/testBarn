package db

import (
	"context"
	"database/sql"
	"encoding/json"
	_ "github.com/jackc/pgx/v4/pgxpool"
	_ "log"
	_ "os"
)

type TestCase struct {
	ID        int64           `json:"id"`
	Test      json.RawMessage `json:"test"`
	SuiteID   sql.NullInt64   `json:"suite_id"`
	SuiteName sql.NullString  `json:"suite_name"`
	Tags      []string        `json:"tags"`
}

func CreateTestCaseInDB(testCase TestCase) (int64, error) {
	tx, err := DBPool.Begin(context.Background())
	if err != nil {
		return 0, err
	}
	// Создаем тест-кейс
	var id int64
	err = tx.QueryRow(context.Background(),
		"INSERT INTO test_cases (test) VALUES ($1) RETURNING id",
		testCase.Test).Scan(&id)
	if err != nil {
		return 0, err
	}

	// Добавляем теги
	tagsJSON, err := json.Marshal(testCase.Tags)
	if err != nil {
		return 0, err
	}

	_, err = tx.Exec(context.Background(),
		"INSERT INTO test_case_tags (case_id, tags) VALUES ($1, $2)",
		id, tagsJSON)
	if err != nil {
		return 0, err
	}

	return id, tx.Commit(context.Background())
}

func GetTestCaseFromDB(id string) (TestCase, error) {
	var testCase TestCase
	query := `
		SELECT tc.id, tc.test, tsc.suite_id, ts.name
		FROM test_cases tc
		LEFT JOIN test_suite_cases tsc ON tc.id = tsc.case_id
		LEFT JOIN test_suites ts ON tsc.suite_id = ts.id
		WHERE tc.id=$1`
	err := DBPool.QueryRow(context.Background(), query, id).Scan(&testCase.ID, &testCase.Test, &testCase.SuiteID, &testCase.SuiteName)
	if err != nil {
		return testCase, err
	}
	return testCase, nil
}

func GetAllTestCases() ([]TestCase, error) {
	rows, err := DBPool.Query(context.Background(), "SELECT id, test FROM test_cases")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var testCases []TestCase
	for rows.Next() {
		var testCase TestCase
		if err := rows.Scan(&testCase.ID, &testCase.Test); err != nil {
			return nil, err
		}
		testCases = append(testCases, testCase)
	}

	// Check for any errors encountered during iteration
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return testCases, nil
}

func UpdateTestCaseInDB(id int64, updatedTest json.RawMessage) error {
	_, err := DBPool.Exec(context.Background(), "UPDATE test_cases SET test=$1 WHERE id=$2", updatedTest, id)
	return err
}

func DeleteTestCaseInDB(id int64) error {
	// delete test from  test_suite_cases
	_, err := DBPool.Exec(context.Background(), "DELETE FROM test_suite_cases WHERE case_id=$1", id)
	if err != nil {
		return err
	}

	_, err = DBPool.Exec(context.Background(), "DELETE FROM test_cases WHERE id=$1", id)
	return err
}
