package api

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"strings"
	"testBarn/db"
)

type CreateTestRunRequest struct {
	SuiteID    *int            `json:"suite_id,omitempty"`
	RunDetails json.RawMessage `json:"run_details"`
	CaseIDs    []int           `json:"test_case_ids"`
	ExecutedBy *string         `json:"executed_by,omitempty"`
}

type UpdateRunCaseStatusRequest struct {
	Status     string  `json:"status"`
	Comment    *string `json:"comment,omitempty"`
	ExecutedBy *string `json:"executed_by,omitempty"`
}

func CreateTestRunHandler(w http.ResponseWriter, r *http.Request) {
	var createReq CreateTestRunRequest
	if err := json.NewDecoder(r.Body).Decode(&createReq); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for _, id := range createReq.CaseIDs {
		if id <= 0 {
			http.Error(w, "Invalid test_case_id", http.StatusBadRequest)
			return
		}
	}
	if createReq.SuiteID != nil && *createReq.SuiteID <= 0 {
		http.Error(w, "Invalid suite_id", http.StatusBadRequest)
		return
	}

	run, err := db.CreateTestRun(db.CreateTestRunParams{
		SuiteID:    createReq.SuiteID,
		CaseIDs:    createReq.CaseIDs,
		RunDetails: createReq.RunDetails,
		ExecutedBy: createReq.ExecutedBy,
	})
	if err != nil {
		switch {
		case errors.Is(err, db.ErrTestSuiteNotFound), errors.Is(err, db.ErrTestCaseNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
		case strings.Contains(err.Error(), "at least one source"), strings.Contains(err.Error(), "resolved test case set is empty"):
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(run)
}

func GetAllTestRunsHandler(w http.ResponseWriter, r *http.Request) {
	runs, err := db.GetAllTestRuns()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(runs)
}

func GetTestRunByIDHandler(w http.ResponseWriter, r *http.Request) {
	runIDStr := mux.Vars(r)["id"]
	if runIDStr == "" {
		http.Error(w, "Missing run ID", http.StatusBadRequest)
		return
	}

	runID, err := strconv.Atoi(runIDStr)
	if err != nil || runID <= 0 {
		http.Error(w, "Invalid run ID", http.StatusBadRequest)
		return
	}

	run, err := db.GetTestRunByID(runID)
	if err != nil {
		if errors.Is(err, db.ErrTestRunNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(run)
}

func UpdateTestRunCaseStatusHandler(w http.ResponseWriter, r *http.Request) {
	runIDStr := mux.Vars(r)["runId"]
	caseIDStr := mux.Vars(r)["caseId"]
	if runIDStr == "" || caseIDStr == "" {
		http.Error(w, "Missing runId or caseId", http.StatusBadRequest)
		return
	}

	runID, err := strconv.Atoi(runIDStr)
	if err != nil || runID <= 0 {
		http.Error(w, "Invalid run ID", http.StatusBadRequest)
		return
	}

	caseID, err := strconv.Atoi(caseIDStr)
	if err != nil || caseID <= 0 {
		http.Error(w, "Invalid case ID", http.StatusBadRequest)
		return
	}

	var updateReq UpdateRunCaseStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if _, ok := db.AllowedRunCaseStatuses[updateReq.Status]; !ok {
		http.Error(w, "Invalid status", http.StatusBadRequest)
		return
	}

	if err := db.UpdateRunCaseStatus(runID, caseID, updateReq.Status, updateReq.Comment, updateReq.ExecutedBy); err != nil {
		switch {
		case errors.Is(err, db.ErrTestRunNotFound), errors.Is(err, db.ErrTestRunCaseNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
		case strings.Contains(err.Error(), "invalid status"):
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	updatedRun, err := db.GetTestRunByID(runID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedRun)
}
