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

type TestSuiteResponse struct {
	ID int `json:"id"`
}

type TestRunCaseResponse struct {
	CaseID     int64      `json:"case_id"`
	Status     string     `json:"status"`
	Comment    *string    `json:"comment,omitempty"`
	ExecutedAt *time.Time `json:"executed_at,omitempty"`
}

type TestRunSummaryResponse struct {
	Passed  int `json:"passed"`
	Failed  int `json:"failed"`
	Blocked int `json:"blocked"`
	Skipped int `json:"skipped"`
	NotRun  int `json:"not_run"`
}

type TestRunResponse struct {
	ID        int                    `json:"id"`
	SuiteID   *int                   `json:"suite_id,omitempty"`
	Cases     []TestRunCaseResponse  `json:"cases,omitempty"`
	Summary   TestRunSummaryResponse `json:"summary"`
	CreatedAt time.Time              `json:"created_at"`
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

func TestCreateListGetAndPatchTestRuns(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/testcases", api.CreateTestCase).Methods("POST")
	r.HandleFunc("/test-suites", api.CreateTestSuiteHandler).Methods("POST")
	r.HandleFunc("/test-suites/add-cases", api.AddTestCasesToSuiteHandler).Methods("POST")
	r.HandleFunc("/test-runs", api.CreateTestRunHandler).Methods("POST")
	r.HandleFunc("/test-runs", api.GetAllTestRunsHandler).Methods("GET")
	r.HandleFunc("/test-runs/{id}", api.GetTestRunByIDHandler).Methods("GET")
	r.HandleFunc("/test-runs/{runId}/cases/{caseId}", api.UpdateTestRunCaseStatusHandler).Methods("PATCH")

	server := httptest.NewServer(r)
	defer server.Close()

	createCase := func(name string) int64 {
		payload := map[string]interface{}{
			"test": map[string]interface{}{
				"name": name,
			},
		}
		body, _ := json.Marshal(payload)
		resp, err := http.Post(server.URL+"/testcases", "application/json", bytes.NewBuffer(body))
		if err != nil {
			t.Fatalf("Failed to create test case: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var created TestCase
		if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
			t.Fatalf("Failed to decode created test case: %v", err)
		}
		return created.ID
	}

	caseFromSuite := createCase("Case from suite")
	caseExtra := createCase("Case extra")

	createSuiteBody, _ := json.Marshal(map[string]string{
		"name":        "Run source suite",
		"description": "Suite for test-runs integration test",
	})
	suiteResp, err := http.Post(server.URL+"/test-suites", "application/json", bytes.NewBuffer(createSuiteBody))
	if err != nil {
		t.Fatalf("Failed to create suite: %v", err)
	}
	defer func() { _ = suiteResp.Body.Close() }()
	assert.Equal(t, http.StatusCreated, suiteResp.StatusCode)

	var createdSuite TestSuiteResponse
	if err := json.NewDecoder(suiteResp.Body).Decode(&createdSuite); err != nil {
		t.Fatalf("Failed to decode suite: %v", err)
	}

	addCasesBody, _ := json.Marshal(map[string]interface{}{
		"suite_id": createdSuite.ID,
		"case_ids": []int64{caseFromSuite},
	})
	addReq, _ := http.NewRequest(http.MethodPost, server.URL+"/test-suites/add-cases", bytes.NewBuffer(addCasesBody))
	addReq.Header.Set("Content-Type", "application/json")
	addResp, err := http.DefaultClient.Do(addReq)
	if err != nil {
		t.Fatalf("Failed to add cases to suite: %v", err)
	}
	defer func() { _ = addResp.Body.Close() }()
	assert.Equal(t, http.StatusOK, addResp.StatusCode)

	createRunBody, _ := json.Marshal(map[string]interface{}{
		"suite_id":      createdSuite.ID,
		"test_case_ids": []int64{caseFromSuite, caseExtra, caseExtra},
		"run_details": map[string]interface{}{
			"name": "Regression run",
		},
		"executed_by": "qa.bot",
	})
	createRunResp, err := http.Post(server.URL+"/test-runs", "application/json", bytes.NewBuffer(createRunBody))
	if err != nil {
		t.Fatalf("Failed to create test run: %v", err)
	}
	defer func() { _ = createRunResp.Body.Close() }()
	assert.Equal(t, http.StatusCreated, createRunResp.StatusCode)

	var createdRun TestRunResponse
	if err := json.NewDecoder(createRunResp.Body).Decode(&createdRun); err != nil {
		t.Fatalf("Failed to decode created run: %v", err)
	}
	assert.NotZero(t, createdRun.ID)
	assert.Len(t, createdRun.Cases, 2, "suite + explicit case IDs should be deduplicated")
	assert.Equal(t, 2, createdRun.Summary.NotRun)

	listRunsResp, err := http.Get(server.URL + "/test-runs")
	if err != nil {
		t.Fatalf("Failed to list test runs: %v", err)
	}
	defer func() { _ = listRunsResp.Body.Close() }()
	assert.Equal(t, http.StatusOK, listRunsResp.StatusCode)

	var runs []TestRunResponse
	if err := json.NewDecoder(listRunsResp.Body).Decode(&runs); err != nil {
		t.Fatalf("Failed to decode runs list: %v", err)
	}
	assert.NotEmpty(t, runs)

	getRunResp, err := http.Get(fmt.Sprintf("%s/test-runs/%d", server.URL, createdRun.ID))
	if err != nil {
		t.Fatalf("Failed to get run by id: %v", err)
	}
	defer func() { _ = getRunResp.Body.Close() }()
	assert.Equal(t, http.StatusOK, getRunResp.StatusCode)

	var fetchedRun TestRunResponse
	if err := json.NewDecoder(getRunResp.Body).Decode(&fetchedRun); err != nil {
		t.Fatalf("Failed to decode fetched run: %v", err)
	}
	assert.Len(t, fetchedRun.Cases, 2)
	assert.Equal(t, 2, fetchedRun.Summary.NotRun)

	targetCaseID := fetchedRun.Cases[0].CaseID
	patchBody, _ := json.Marshal(map[string]interface{}{
		"status":      "passed",
		"comment":     "Executed successfully",
		"executed_by": "qa.user",
	})
	patchReq, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("%s/test-runs/%d/cases/%d", server.URL, createdRun.ID, targetCaseID), bytes.NewBuffer(patchBody))
	patchReq.Header.Set("Content-Type", "application/json")
	patchResp, err := http.DefaultClient.Do(patchReq)
	if err != nil {
		t.Fatalf("Failed to patch run case status: %v", err)
	}
	defer func() { _ = patchResp.Body.Close() }()
	assert.Equal(t, http.StatusOK, patchResp.StatusCode)

	var patchedRun TestRunResponse
	if err := json.NewDecoder(patchResp.Body).Decode(&patchedRun); err != nil {
		t.Fatalf("Failed to decode patched run response: %v", err)
	}
	assert.Equal(t, 1, patchedRun.Summary.Passed)
	assert.Equal(t, 1, patchedRun.Summary.NotRun)
}

func TestTestRunContractErrors(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/test-runs", api.CreateTestRunHandler).Methods("POST")
	r.HandleFunc("/test-runs/{id}", api.GetTestRunByIDHandler).Methods("GET")
	r.HandleFunc("/test-runs/{runId}/cases/{caseId}", api.UpdateTestRunCaseStatusHandler).Methods("PATCH")

	server := httptest.NewServer(r)
	defer server.Close()

	t.Run("create run requires source", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"run_details": map[string]interface{}{"name": "bad run"},
		})
		resp, err := http.Post(server.URL+"/test-runs", "application/json", bytes.NewBuffer(body))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("create run with invalid case id", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"test_case_ids": []int{99999999},
		})
		resp, err := http.Post(server.URL+"/test-runs", "application/json", bytes.NewBuffer(body))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("get run with invalid id", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/test-runs/not-a-number")
		if err != nil {
			t.Fatalf("Failed to get run: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("get non-existing run", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/test-runs/9999999")
		if err != nil {
			t.Fatalf("Failed to get run: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("patch with invalid status", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"status": "unknown",
		})
		req, _ := http.NewRequest(http.MethodPatch, server.URL+"/test-runs/1/cases/1", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to patch run case: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("patch non-existing run", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"status": "passed",
		})
		req, _ := http.NewRequest(http.MethodPatch, server.URL+"/test-runs/9999999/cases/1", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to patch run case: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
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
