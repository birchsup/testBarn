package db

import (
	"encoding/json"
	"time"
)

type TestRun struct {
	ID         int             `json:"id"`
	SuiteID    int             `json:"suite_id"`
	RunDetails json.RawMessage `json:"run_details"`
	CreatedAt  time.Time       `json:"created_at"`
}

// TestRun and related DB schema are kept for a deferred test_runs feature.
// Decision (2026-02-08): tables remain migrated, but write/read API is intentionally disabled.
