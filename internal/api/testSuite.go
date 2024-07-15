package api

import (
	"encoding/json"
	"net/http"
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
