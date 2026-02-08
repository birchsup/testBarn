package api

import (
	"encoding/json"
	"net/http"
)

type CreateTestRunRequest struct {
	SuiteID    int             `json:"suite_id"`
	RunDetails json.RawMessage `json:"run_details"`
	CaseIDs    []int           `json:"case_ids"`
}

// CreateTestRunHandler intentionally returns 501 while the feature is deferred.
// Decision (2026-02-08): keep DB schema for test_runs/test_run_cases, but keep API closed.
func CreateTestRunHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "test_runs API is not enabled", http.StatusNotImplemented)
}
