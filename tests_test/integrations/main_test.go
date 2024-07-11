package integrations_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"log"
	"net/http"
	"os"
	"testBarn/internal/api"
	"testBarn/internal/db"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func startPostgresContainer() (testcontainers.Container, string, error) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:15",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_PASSWORD": "password",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(5 * time.Minute),
	}

	postgresC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, "", err
	}

	host, err := postgresC.Host(ctx)
	if err != nil {
		return nil, "", err
	}

	port, err := postgresC.MappedPort(ctx, "5432")
	if err != nil {
		return nil, "", err
	}

	dbURL := fmt.Sprintf("postgres://postgres:password@%s:%s/testdb?sslmode=disable", host, port.Port())
	return postgresC, dbURL, nil
}

func runMigrations(dbURL string) error {
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %w", err)
	}

	migrationsPath := fmt.Sprintf("file:///Users/dmitrijsadovnikov/testBarn/db/migrations")

	println("running migrations from", migrationsPath)
	m, err := migrate.NewWithDatabaseInstance(
		migrationsPath,
		"postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	log.Println("Starting migrations...")

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Migrations ran successfully")
	return nil
}

func logTables(dbURL string) {
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT table_name FROM information_schema.tables WHERE table_schema='public'")
	if err != nil {
		log.Fatalf("Failed to query tables: %v", err)
	}
	defer rows.Close()

	log.Println("Tables in the database:")
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			log.Fatalf("Failed to scan table name: %v", err)
		}
		log.Println(tableName)
	}
	if err := rows.Err(); err != nil {
		log.Fatalf("Rows error: %v", err)
	}
}

func TestMain(m *testing.M) {
	postgresC, dbURL, err := startPostgresContainer()
	if err != nil {
		log.Fatalf("Failed to start container: %v", err)
	}
	defer postgresC.Terminate(context.Background())

	os.Setenv("DATABASE_URL", dbURL)

	db.InitDB()
	defer db.DBPool.Close()

	if err := runMigrations(dbURL); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	logTables(dbURL)

	code := m.Run()
	os.Exit(code)
}

type TestCase struct {
	ID   int64           `json:"id"`
	Test json.RawMessage `json:"test"`
}

func TestCreateAndGetTestCase(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/testcases", api.CreateTestCase).Methods("POST")
	r.HandleFunc("/testcases/{id:[0-9]+}", api.GetTestCase).Methods("GET")

	server := &http.Server{
		Addr:    ":8081",
		Handler: r,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not listen on :8081: %v\n", err)
		}
	}()
	defer server.Close()

	time.Sleep(2 * time.Second) // Дайте серверу время для запуска

	testCase := map[string]interface{}{
		"test": map[string]interface{}{
			"name": "Login Test",
			"steps": []map[string]interface{}{
				{"step": 1, "action": "Open login page", "expected_result": "Login page is displayed"},
				{"step": 2, "action": "Enter username", "expected_result": "Username is entered"},
				{"step": 3, "action": "Enter password", "expected_result": "Password is entered"},
				{"step": 4, "action": "Click login button", "expected_result": "User is logged in"},
			},
			"created_by": "QA Engineer",
			"created_at": "2024-07-06",
		},
	}

	testCaseBytes, _ := json.Marshal(testCase)
	resp, err := http.Post("http://localhost:8081/testcases", "application/json", bytes.NewBuffer(testCaseBytes))
	if err != nil {
		t.Fatalf("Failed to create test case: %v", err)
	}
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var createdTestCase TestCase
	err = json.NewDecoder(resp.Body).Decode(&createdTestCase)
	if err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}
	assert.NotZero(t, createdTestCase.ID)

	getResp, err := http.Get("http://localhost:8081/testcases/" + fmt.Sprint(createdTestCase.ID))
	if err != nil {
		t.Fatalf("Failed to get test case: %v", err)
	}
	assert.Equal(t, http.StatusOK, getResp.StatusCode)

	var fetchedTestCase TestCase
	err = json.NewDecoder(getResp.Body).Decode(&fetchedTestCase)
	if err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}
	assert.Equal(t, createdTestCase.ID, fetchedTestCase.ID)
	assert.JSONEq(t, string(createdTestCase.Test), string(fetchedTestCase.Test))
}
