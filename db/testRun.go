package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v4"
	"sort"
	"time"
)

var (
	ErrTestSuiteNotFound   = errors.New("test suite not found")
	ErrTestCaseNotFound    = errors.New("test case not found")
	ErrTestRunNotFound     = errors.New("test run not found")
	ErrTestRunCaseNotFound = errors.New("test run case not found")
)

var AllowedRunCaseStatuses = map[string]struct{}{
	"passed":  {},
	"failed":  {},
	"blocked": {},
	"skipped": {},
	"not_run": {},
}

type TestRun struct {
	ID         int             `json:"id"`
	SuiteID    *int            `json:"suite_id,omitempty"`
	RunDetails json.RawMessage `json:"run_details"`
	CreatedAt  time.Time       `json:"created_at"`
}

type TestRunCase struct {
	CaseID     int64           `json:"case_id"`
	Test       json.RawMessage `json:"test"`
	Status     string          `json:"status"`
	Comment    *string         `json:"comment,omitempty"`
	ExecutedAt *time.Time      `json:"executed_at,omitempty"`
	ExecutedBy *string         `json:"executed_by,omitempty"`
}

type TestRunSummary struct {
	Passed  int `json:"passed"`
	Failed  int `json:"failed"`
	Blocked int `json:"blocked"`
	Skipped int `json:"skipped"`
	NotRun  int `json:"not_run"`
}

type TestRunDetails struct {
	TestRun
	Cases   []TestRunCase  `json:"cases"`
	Summary TestRunSummary `json:"summary"`
}

type CreateTestRunParams struct {
	SuiteID    *int            `json:"suite_id,omitempty"`
	CaseIDs    []int           `json:"test_case_ids"`
	RunDetails json.RawMessage `json:"run_details"`
	ExecutedBy *string         `json:"executed_by,omitempty"`
}

func normalizeCaseIDs(caseIDs []int) []int {
	seen := make(map[int]struct{}, len(caseIDs))
	result := make([]int, 0, len(caseIDs))
	for _, id := range caseIDs {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	sort.Ints(result)
	return result
}

func validateAllCasesExist(ctx context.Context, caseIDs []int) error {
	if len(caseIDs) == 0 {
		return nil
	}

	query := `SELECT id FROM test_cases WHERE id = ANY($1::int[])`
	rows, err := DBPool.Query(ctx, query, caseIDs)
	if err != nil {
		return err
	}
	defer rows.Close()

	existing := make(map[int]struct{}, len(caseIDs))
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return err
		}
		existing[id] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, id := range caseIDs {
		if _, ok := existing[id]; !ok {
			return ErrTestCaseNotFound
		}
	}
	return nil
}

func collectSuiteCaseIDs(ctx context.Context, suiteID int) ([]int, error) {
	var exists bool
	if err := DBPool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM test_suites WHERE id = $1)`, suiteID).Scan(&exists); err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrTestSuiteNotFound
	}

	rows, err := DBPool.Query(ctx, `SELECT case_id FROM test_suite_cases WHERE suite_id = $1`, suiteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ids, nil
}

func CreateTestRun(params CreateTestRunParams) (TestRunDetails, error) {
	ctx := context.Background()

	dedupExplicitIDs := normalizeCaseIDs(params.CaseIDs)
	if params.SuiteID == nil && len(dedupExplicitIDs) == 0 {
		return TestRunDetails{}, fmt.Errorf("at least one source of test cases is required")
	}

	if err := validateAllCasesExist(ctx, dedupExplicitIDs); err != nil {
		return TestRunDetails{}, err
	}

	allCaseIDs := make([]int, 0, len(dedupExplicitIDs)+8)
	allCaseIDs = append(allCaseIDs, dedupExplicitIDs...)
	if params.SuiteID != nil {
		suiteCaseIDs, err := collectSuiteCaseIDs(ctx, *params.SuiteID)
		if err != nil {
			return TestRunDetails{}, err
		}
		allCaseIDs = append(allCaseIDs, suiteCaseIDs...)
	}
	allCaseIDs = normalizeCaseIDs(allCaseIDs)
	if len(allCaseIDs) == 0 {
		return TestRunDetails{}, fmt.Errorf("resolved test case set is empty")
	}

	runDetails := params.RunDetails
	if len(runDetails) == 0 {
		runDetails = json.RawMessage(`{}`)
	}

	tx, err := DBPool.Begin(ctx)
	if err != nil {
		return TestRunDetails{}, err
	}
	defer tx.Rollback(ctx)

	var createdRun TestRun
	if params.SuiteID != nil {
		err = tx.QueryRow(ctx, `
			INSERT INTO test_runs (suite_id, run_details)
			VALUES ($1, $2)
			RETURNING id, suite_id, run_details, created_at
		`, *params.SuiteID, runDetails).Scan(&createdRun.ID, &createdRun.SuiteID, &createdRun.RunDetails, &createdRun.CreatedAt)
	} else {
		err = tx.QueryRow(ctx, `
			INSERT INTO test_runs (suite_id, run_details)
			VALUES (NULL, $1)
			RETURNING id, suite_id, run_details, created_at
		`, runDetails).Scan(&createdRun.ID, &createdRun.SuiteID, &createdRun.RunDetails, &createdRun.CreatedAt)
	}
	if err != nil {
		return TestRunDetails{}, err
	}

	for _, caseID := range allCaseIDs {
		if _, err := tx.Exec(ctx, `
			INSERT INTO test_run_cases (run_id, case_id, status, executed_by)
			VALUES ($1, $2, 'not_run', $3)
		`, createdRun.ID, caseID, params.ExecutedBy); err != nil {
			return TestRunDetails{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return TestRunDetails{}, err
	}

	return GetTestRunByID(createdRun.ID)
}

func GetAllTestRuns() ([]TestRun, error) {
	rows, err := DBPool.Query(context.Background(), `
		SELECT id, suite_id, run_details, created_at
		FROM test_runs
		ORDER BY created_at DESC, id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runs []TestRun
	for rows.Next() {
		var run TestRun
		if err := rows.Scan(&run.ID, &run.SuiteID, &run.RunDetails, &run.CreatedAt); err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return runs, nil
}

func GetTestRunByID(runID int) (TestRunDetails, error) {
	ctx := context.Background()

	var details TestRunDetails
	err := DBPool.QueryRow(ctx, `
		SELECT id, suite_id, run_details, created_at
		FROM test_runs
		WHERE id = $1
	`, runID).Scan(&details.ID, &details.SuiteID, &details.RunDetails, &details.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return TestRunDetails{}, ErrTestRunNotFound
		}
		return TestRunDetails{}, err
	}

	rows, err := DBPool.Query(ctx, `
		SELECT trc.case_id, tc.test, trc.status, trc.comment, trc.executed_at, trc.executed_by
		FROM test_run_cases trc
		JOIN test_cases tc ON tc.id = trc.case_id
		WHERE trc.run_id = $1
		ORDER BY trc.case_id
	`, runID)
	if err != nil {
		return TestRunDetails{}, err
	}
	defer rows.Close()

	details.Cases = make([]TestRunCase, 0)
	for rows.Next() {
		var c TestRunCase
		var comment sql.NullString
		var executedAt sql.NullTime
		var executedBy sql.NullString
		if err := rows.Scan(&c.CaseID, &c.Test, &c.Status, &comment, &executedAt, &executedBy); err != nil {
			return TestRunDetails{}, err
		}
		if comment.Valid {
			c.Comment = &comment.String
		}
		if executedAt.Valid {
			t := executedAt.Time
			c.ExecutedAt = &t
		}
		if executedBy.Valid {
			c.ExecutedBy = &executedBy.String
		}
		details.Cases = append(details.Cases, c)
		switch c.Status {
		case "passed":
			details.Summary.Passed++
		case "failed":
			details.Summary.Failed++
		case "blocked":
			details.Summary.Blocked++
		case "skipped":
			details.Summary.Skipped++
		default:
			details.Summary.NotRun++
		}
	}
	if err := rows.Err(); err != nil {
		return TestRunDetails{}, err
	}

	return details, nil
}

func UpdateRunCaseStatus(runID int, caseID int, status string, comment *string, executedBy *string) error {
	if _, ok := AllowedRunCaseStatuses[status]; !ok {
		return fmt.Errorf("invalid status")
	}

	result, err := DBPool.Exec(context.Background(), `
		UPDATE test_run_cases
		SET status = $3,
			comment = $4,
			executed_at = CASE WHEN $3 = 'not_run' THEN NULL ELSE NOW() END,
			executed_by = $5
		WHERE run_id = $1 AND case_id = $2
	`, runID, caseID, status, comment, executedBy)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		var exists bool
		if err := DBPool.QueryRow(context.Background(), `SELECT EXISTS(SELECT 1 FROM test_runs WHERE id = $1)`, runID).Scan(&exists); err != nil {
			return err
		}
		if !exists {
			return ErrTestRunNotFound
		}
		return ErrTestRunCaseNotFound
	}

	return nil
}
