package api

import (
	"context"
	_ "database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"testBarn/db"
)

func GetTestCasesByTagsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tagsStr := r.URL.Query().Get("tags")
	if tagsStr == "" {
		http.Error(w, "Missing tags parameter", http.StatusBadRequest)
		return
	}

	tags := strings.Split(tagsStr, ",")
	testCases, err := db.GetTestCasesByTags(context.Background(), tags)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("ngrok-skip-browser-warning", "true")
	err = json.NewEncoder(w).Encode(testCases)
	if err != nil {
		return
	}
}

func AddTagsToTestCaseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	caseIDStr := r.URL.Query().Get("id")
	if caseIDStr == "" {
		http.Error(w, "Missing case ID", http.StatusBadRequest)
		return
	}

	caseID, err := strconv.Atoi(caseIDStr)
	if err != nil {
		http.Error(w, "Invalid case ID", http.StatusBadRequest)
		return
	}

	var tags []string
	err = json.NewDecoder(r.Body).Decode(&tags)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = db.AddTagsToTestCase(context.Background(), int64(caseID), tags)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func RemoveTagsFromTestCaseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	caseIDStr := r.URL.Query().Get("id")
	if caseIDStr == "" {
		http.Error(w, "Missing case ID", http.StatusBadRequest)
		return
	}

	caseID, err := strconv.Atoi(caseIDStr)
	if err != nil {
		http.Error(w, "Invalid case ID", http.StatusBadRequest)
		return
	}

	var tags []string
	err = json.NewDecoder(r.Body).Decode(&tags)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = db.RemoveTagsFromTestCase(context.Background(), int64(caseID), tags)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func GetTagsForTestCaseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	caseIDStr := r.URL.Query().Get("id")
	if caseIDStr == "" {
		http.Error(w, "Missing case ID", http.StatusBadRequest)
		return
	}

	caseID, err := strconv.Atoi(caseIDStr)
	if err != nil {
		http.Error(w, "Invalid case ID", http.StatusBadRequest)
		return
	}

	tags, err := db.GetTagsForTestCase(context.Background(), int64(caseID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("ngrok-skip-browser-warning", "true")
	err = json.NewEncoder(w).Encode(tags)
	if err != nil {
		return
	}
}
