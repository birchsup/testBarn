package db

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/jackc/pgx/v4/pgxpool"
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
