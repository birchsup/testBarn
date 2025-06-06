package main

import (
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"testBarn/config"
	"testBarn/db"
	"testBarn/internal/api"
)

func main() {
	config.InitConfig()
	db.InitDB()
	defer db.DBPool.Close()

	r := mux.NewRouter()
	//testCases
	r.HandleFunc("/testcases", api.CreateTestCase).Methods("POST")
	r.HandleFunc("/testcase", api.GetTestCaseHandler).Methods("GET")
	r.HandleFunc("/testcases", api.GetAllTestCases).Methods("GET")
	r.HandleFunc("/test-case/update", api.UpdateTestCaseHandler).Methods("PUT")
	r.HandleFunc("/test-case/delete", api.DeleteTestCaseHandler).Methods("DELETE")

	//test suites
	r.HandleFunc("/test-suites", api.GetAllTestSuitesHandler).Methods("GET")
	r.HandleFunc("/add-test-suite", api.CreateTestSuiteHandler).Methods("POST")
	r.HandleFunc("/test-suites/add-cases", api.AddTestCasesToSuiteHandler).Methods("POST")
	r.HandleFunc("/test-suite", api.GetTestSuiteByIDHandler).Methods("GET")
	r.HandleFunc("/test-suite/update", api.UpdateTestSuiteHandler).Methods("PUT")
	r.HandleFunc("/test-suite/delete", api.DeleteTestSuiteHandler).Methods("DELETE")
	r.HandleFunc("/test-suite/remove-case", api.RemoveTestCaseFromSuiteHandler).Methods("DELETE")
	// test tags
	r.HandleFunc("/api/test-cases/tags", api.GetTestCasesByTagsHandler).Methods("GET")
	r.HandleFunc("/api/test-case/tags", api.AddTagsToTestCaseHandler).Methods("PUT")
	r.HandleFunc("/api/test-case/tags/remove", api.RemoveTagsFromTestCaseHandler).Methods("DELETE")
	r.HandleFunc("/api/test-case/tags/list", api.GetTagsForTestCaseHandler).Methods("GET")

	// Test Reports (Playwright)
	r.HandleFunc("/test-reports", api.CreateTestReportHandler).Methods("POST")   // загрузка отчета
	r.HandleFunc("/test-reports", api.GetAllTestReportsHandler).Methods("GET")   // список всех отчетов
	r.HandleFunc("/test-reports", api.GetTestReportHandler).Methods("GET")       // получение отчета по ID
	r.HandleFunc("/test-reports", api.DeleteTestReportHandler).Methods("DELETE") // удаление отчета

	// Test Runs
	r.HandleFunc("/test-runs", api.CreateTestRunHandler).Methods("POST")
	r.HandleFunc("/test-runs", api.GetAllTestRunsHandler).Methods("GET")
	r.HandleFunc("/test-runs/cases", api.GetAllTestCasesFromRunHandler).Methods("GET")
	r.HandleFunc("/test-runs/case", api.GetTestCaseDetailsHandler).Methods("GET")
	r.HandleFunc("/test-runs/suite", api.AddTestSuiteToRunHandler).Methods("POST")
	r.HandleFunc("/test-runs/case", api.AddTestCaseToRunHandler).Methods("POST")
	r.HandleFunc("/test-runs/case/status", api.UpdateTestCaseStatusHandler).Methods("PUT")

	// CORS
	corsHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
		handlers.AllowedHeaders([]string{"ngrok-skip-browser-warning", "truego"}),
	)(r)

	log.Println("Server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", corsHandler))
}
