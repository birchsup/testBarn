package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"testBarn/db"
)

type CreateTestRunRequest struct {
	RunDetails json.RawMessage `json:"run_details"`
}

func GetAllTestCasesFromRunHandler(w http.ResponseWriter, r *http.Request) {
	runIDStr := r.URL.Query().Get("run_id")
	if runIDStr == "" {
		http.Error(w, "Missing run ID", http.StatusBadRequest)
		return
	}

	runID, err := strconv.ParseInt(runIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid run ID", http.StatusBadRequest)
		return
	}

	log.Printf("Fetching test cases for run_id: %d", runID)
	testCasesRun, err := db.GetAllTestCasesFromRun(runID)
	if err != nil {
		log.Printf("Error fetching test cases for run_id %d: %v", runID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Новый массив, чтобы хранить только валидные кейсы
	validTestCases := []db.TestCaseRun{}

	for _, testCase := range testCasesRun {
		log.Printf("Test Case ID: %d, Name: %s, Suite ID: %v, Status: %s",
			testCase.ID, testCase.Name, testCase.SuiteID, testCase.Status)

		// Проверяем, что name не пустой (он же не JSON теперь)
		if testCase.Name == "" {
			log.Printf("Skipping test case %d: missing name", testCase.ID)
			continue
		}

		validTestCases = append(validTestCases, testCase)
	}

	log.Printf("Valid test cases for run_id %d: %d cases found", runID, len(validTestCases))

	// Отправляем результат как JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(validTestCases); err != nil {
		log.Printf("Error encoding response JSON: %v", err)
		http.Error(w, "Error encoding response JSON", http.StatusInternalServerError)
	}
}

func GetTestCaseDetailsHandler(w http.ResponseWriter, r *http.Request) {
	runIDStr := r.URL.Query().Get("run_id")
	caseIDStr := r.URL.Query().Get("case_id")

	if runIDStr == "" || caseIDStr == "" {
		http.Error(w, "Missing run ID or case ID", http.StatusBadRequest)
		return
	}

	runID, err := strconv.ParseInt(runIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid run ID", http.StatusBadRequest)
		return
	}

	caseID, err := strconv.ParseInt(caseIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid case ID", http.StatusBadRequest)
		return
	}

	testCase, err := db.GetTestCaseDetails(runID, caseID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Test case not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(testCase)
}

func AddTestSuiteToRunHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	runIDStr := r.URL.Query().Get("run_id")
	suiteIDStr := r.URL.Query().Get("suite_id")

	if runIDStr == "" || suiteIDStr == "" {
		http.Error(w, "Missing run ID or suite ID", http.StatusBadRequest)
		return
	}

	runID, err := strconv.ParseInt(runIDStr, 10, 64)
	suiteID, err := strconv.ParseInt(suiteIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	err = db.AddTestSuiteToRun(runID, suiteID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Test suite added to run"})
}

func AddTestCaseToRunHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	runIDStr := r.URL.Query().Get("run_id")
	caseIDStr := r.URL.Query().Get("case_id")

	if runIDStr == "" || caseIDStr == "" {
		http.Error(w, "Missing run ID or case ID", http.StatusBadRequest)
		return
	}

	runID, err := strconv.ParseInt(runIDStr, 10, 64)
	caseID, err := strconv.ParseInt(caseIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	err = db.AddTestSuiteToRun(runID, caseID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Test case added to run"})
}

func UpdateTestCaseStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	runIDStr := r.URL.Query().Get("run_id")
	caseIDStr := r.URL.Query().Get("case_id")

	if runIDStr == "" || caseIDStr == "" {
		http.Error(w, "Missing run ID or case ID", http.StatusBadRequest)
		return
	}

	runID, err := strconv.ParseInt(runIDStr, 10, 64)
	caseID, err := strconv.ParseInt(caseIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// Читаем тело запроса
	var requestBody struct {
		Status  string `json:"status"`
		Comment string `json:"comment"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		log.Printf("Invalid request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("Updating test case %d in run %d with status: %s", caseID, runID, requestBody.Status)

	// Обновляем статус в БД
	err = db.UpdateTestCaseStatus(runID, caseID, requestBody.Status, requestBody.Comment)
	if err != nil {
		http.Error(w, "Failed to update test case status", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Test case status updated successfully"})
}

func CreateTestRunHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateTestRunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	runID, err := db.CreateTestRun(req.RunDetails)
	if err != nil {
		log.Printf("Error creating test run: %v", err)
		http.Error(w, "Failed to create test run", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"run_id": runID})
}

func GetAllTestRunsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Fetching all test runs")
	testRuns, err := db.GetAllTestRuns()
	if err != nil {
		log.Printf("Error fetching test runs: %v", err)
		http.Error(w, "Failed to fetch test runs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(testRuns); err != nil {
		log.Printf("Error encoding response JSON: %v", err)
		http.Error(w, "Error encoding response JSON", http.StatusInternalServerError)
	}
}
