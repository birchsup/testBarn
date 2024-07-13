package api

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
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
	json.NewEncoder(w).Encode(testCase)
}

func GetAllTestCases(w http.ResponseWriter, r *http.Request) {
	testCases, err := db.GetAllTestCases()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(testCases)
}
