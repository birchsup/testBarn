package api

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"testBarn/db"
)

func CreateTestCase(w http.ResponseWriter, r *http.Request) {
	var testCase db.TestCase
	if err := json.NewDecoder(r.Body).Decode(&testCase); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := db.CreateTestCaseInDB(testCase)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	testCase.ID = id
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("ngrok-skip-browser-warning", "true")
	json.NewEncoder(w).Encode(testCase)
}

func GetTestCase(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	testCase, err := db.GetTestCaseFromDB(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("ngrok-skip-browser-warning", "true")
	json.NewEncoder(w).Encode(testCase)
}

func GetAllTestCases(w http.ResponseWriter, r *http.Request) {
	testCases, err := db.GetAllTestCases()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("ngrok-skip-browser-warning", "true")
	json.NewEncoder(w).Encode(testCases)
}

func UpdateTestCaseHandler(w http.ResponseWriter, r *http.Request) {
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

	var updatedTest db.TestCase
	err = json.NewDecoder(r.Body).Decode(&updatedTest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = db.UpdateTestCaseInDB(int64(caseID), updatedTest.Test)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
func DeleteTestCaseHandler(w http.ResponseWriter, r *http.Request) {
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

	err = db.DeleteTestCaseInDB(int64(caseID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(map[string]string{"message": "Test case was deleted"})
	if err != nil {
		return
	}
}
