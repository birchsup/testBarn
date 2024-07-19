package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"testBarn/db"
)

func CreateTestSuiteHandler(w http.ResponseWriter, r *http.Request) {
	var request db.TestSuiteRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	testSuite, err := db.CreateTestSuite(request.Name, request.Description)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(testSuite)
}

func AddTestCasesToSuiteHandler(w http.ResponseWriter, r *http.Request) {
	var request db.AddTestCaseRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = db.AddTestCasesToSuite(request.SuiteID, request.CaseIDs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func GetTestSuiteByIDHandler(w http.ResponseWriter, r *http.Request) {
	suiteID := r.URL.Query().Get("id")
	if suiteID == "" {
		http.Error(w, "Missing suite ID", http.StatusBadRequest)
		return
	}

	testSuite, err := db.GetTestSuiteByID(suiteID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(testSuite)
	if err != nil {
		return
	}
}
func UpdateTestSuiteHandler(w http.ResponseWriter, r *http.Request) {
	var request db.TestSuiteRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	suiteIDStr := r.URL.Query().Get("id")
	if suiteIDStr == "" {
		http.Error(w, "Missing suite ID", http.StatusBadRequest)
		return
	}

	suiteID, err := strconv.Atoi(suiteIDStr)
	if err != nil {
		http.Error(w, "Invalid suite ID", http.StatusBadRequest)
		return
	}

	testSuite, err := db.UpdateTestSuite(suiteID, request.Name, request.Description)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(testSuite)
}

func DeleteTestSuiteHandler(w http.ResponseWriter, r *http.Request) {
	suiteIDStr := r.URL.Query().Get("id")
	if suiteIDStr == "" {
		http.Error(w, "Missing suite ID", http.StatusBadRequest)
		return
	}

	suiteID, err := strconv.Atoi(suiteIDStr)
	if err != nil {
		http.Error(w, "Invalid suite ID", http.StatusBadRequest)
		return
	}

	err = db.DeleteTestSuite(suiteID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func RemoveTestCaseFromSuiteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	suiteIDStr := r.URL.Query().Get("suite_id")
	if suiteIDStr == "" {
		http.Error(w, "Missing suite ID", http.StatusBadRequest)
		return
	}

	caseIDStr := r.URL.Query().Get("case_id")
	if caseIDStr == "" {
		http.Error(w, "Missing case ID", http.StatusBadRequest)
		return
	}

	suiteID, err := strconv.Atoi(suiteIDStr)
	if err != nil {
		http.Error(w, "Invalid suite ID", http.StatusBadRequest)
		return
	}

	caseID, err := strconv.Atoi(caseIDStr)
	if err != nil {
		http.Error(w, "Invalid case ID", http.StatusBadRequest)
		return
	}

	err = db.RemoveTestCaseFromSuite(suiteID, caseID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
