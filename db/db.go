package db

import (
	"context"
	"encoding/json"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
	"os"
)

var DBPool *pgxpool.Pool

func InitDB() {
	var err error
	DBPool, err = pgxpool.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
}

type TestCase struct {
	ID   int64           `json:"id"`
	Test json.RawMessage `json:"test"`
}

func CreateTestCaseInDB(testCase TestCase) (int64, error) {
	var id int64
	err := DBPool.QueryRow(context.Background(), "INSERT INTO test_cases (test) VALUES ($1) RETURNING id", testCase.Test).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func GetTestCaseFromDB(id string) (TestCase, error) {
	var testCase TestCase
	err := DBPool.QueryRow(context.Background(), "SELECT id, test FROM test_cases WHERE id=$1", id).Scan(&testCase.ID, &testCase.Test)
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
