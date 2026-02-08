package integrations_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testBarn/db"
	"testBarn/internal/api"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/gorilla/mux"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/stretchr/testify/assert"
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
	conn, err := sql.Open("pgx", dbURL)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer func() {
		_ = conn.Close()
	}()

	driver, err := postgres.WithInstance(conn, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %w", err)
	}

	migrationsPath, err := migrationsPath()
	if err != nil {
		return err
	}

	log.Println("running migrations from", migrationsPath)
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
	conn, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer func() {
		_ = conn.Close()
	}()

	rows, err := conn.Query("SELECT table_name FROM information_schema.tables WHERE table_schema='public'")
	if err != nil {
		log.Fatalf("Failed to query tables: %v", err)
	}
	defer func() {
		_ = rows.Close()
	}()

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

func migrationsPath() (string, error) {
	// Portable resolution order:
	// 1) TEST_MIGRATIONS_PATH env (absolute/relative path)
	// 2) discover project root from current working directory
	// 3) discover project root from this test file location
	if fromEnv := strings.TrimSpace(os.Getenv("TEST_MIGRATIONS_PATH")); fromEnv != "" {
		absPath, err := filepath.Abs(fromEnv)
		if err != nil {
			return "", fmt.Errorf("failed to resolve TEST_MIGRATIONS_PATH=%q: %w", fromEnv, err)
		}
		if stat, err := os.Stat(absPath); err != nil || !stat.IsDir() {
			return "", fmt.Errorf("TEST_MIGRATIONS_PATH does not point to an existing directory: %s", absPath)
		}
		return "file://" + filepath.ToSlash(absPath), nil
	}

	if wd, err := os.Getwd(); err == nil {
		if root, ok := findProjectRoot(wd); ok {
			return "file://" + filepath.ToSlash(filepath.Join(root, "db", "migrations")), nil
		}
	}

	_, filename, _, ok := runtime.Caller(0)
	if ok {
		if root, ok := findProjectRoot(filepath.Dir(filename)); ok {
			return "file://" + filepath.ToSlash(filepath.Join(root, "db", "migrations")), nil
		}
	}

	return "", fmt.Errorf("failed to locate db/migrations; set TEST_MIGRATIONS_PATH explicitly")
}

func findProjectRoot(start string) (string, bool) {
	curr := filepath.Clean(start)
	for {
		migrationsDir := filepath.Join(curr, "db", "migrations")
		if stat, err := os.Stat(migrationsDir); err == nil && stat.IsDir() {
			return curr, true
		}

		parent := filepath.Dir(curr)
		if parent == curr {
			return "", false
		}
		curr = parent
	}
}

func TestMain(m *testing.M) {
	useExternalDB := os.Getenv("TEST_USE_EXTERNAL_DB") == "1"
	var dbURL string
	var cleanup func()

	if useExternalDB {
		dbURL = os.Getenv("DATABASE_URL")
		if dbURL == "" {
			log.Fatal("DATABASE_URL is required when TEST_USE_EXTERNAL_DB=1")
		}
	} else {
		postgresC, url, err := startPostgresContainer()
		if err != nil {
			log.Fatalf("Failed to start container: %v", err)
		}
		dbURL = url
		cleanup = func() {
			_ = postgresC.Terminate(context.Background())
		}
	}

	if err := os.Setenv("DATABASE_URL", dbURL); err != nil {
		log.Fatalf("Failed to set DATABASE_URL: %v", err)
	}

	db.InitDB()
	defer db.DBPool.Close()

	if err := runMigrations(dbURL); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	code := m.Run()

	if cleanup != nil {
		cleanup()
	}
	os.Exit(code)
}

type TestCase struct {
	ID   int64           `json:"id"`
	Test json.RawMessage `json:"test"`
}

func TestCreateAndGetTestCase(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/testcases", api.CreateTestCase).Methods("POST")
	r.HandleFunc("/testcases/{id}", api.GetTestCaseHandler).Methods("GET")

	server := &http.Server{
		Addr:    ":8081",
		Handler: r,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not listen on :8081: %v\n", err)
		}
	}()
	defer func() {
		_ = server.Close()
	}()

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
	defer func() {
		_ = resp.Body.Close()
	}()
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
	defer func() {
		_ = getResp.Body.Close()
	}()
	assert.Equal(t, http.StatusOK, getResp.StatusCode)

	var fetchedTestCase TestCase
	err = json.NewDecoder(getResp.Body).Decode(&fetchedTestCase)
	if err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}
	assert.Equal(t, createdTestCase.ID, fetchedTestCase.ID)
	assert.JSONEq(t, string(createdTestCase.Test), string(fetchedTestCase.Test))
}

func TestGetTestCaseContractErrors(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/testcases/{id}", api.GetTestCaseHandler).Methods("GET")

	t.Run("invalid id format", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/testcases/not-a-number", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("missing id in path", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/testcases", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("non-existing id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/testcases/99999999", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("zero id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/testcases/0", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("negative id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/testcases/-10", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

// Test function to check if test_cases table was created
func TestTableTestCases(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	conn, err := sql.Open("pgx", dbURL)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() {
		_ = conn.Close()
	}()

	tables := []string{
		"test_cases",
		"test_suites",
		"test_suite_cases",
		"test_runs",
		"test_run_cases",
	}

	for _, table := range tables {
		query := `
			SELECT EXISTS (
				SELECT FROM information_schema.tables 
				WHERE table_schema = 'public'
				AND table_name = $1
			);
		`
		var exists bool
		err := conn.QueryRow(query, table).Scan(&exists)
		if err != nil {
			t.Fatalf("Failed to check if table %s exists: %v", table, err)
		}

		assert.True(t, exists, "Table %s should exist after migrations", table)
	}
}
