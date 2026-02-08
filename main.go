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
	r.HandleFunc("/testcases/{id}", api.GetTestCaseHandler).Methods("GET")
	r.HandleFunc("/testcases", api.GetAllTestCases).Methods("GET")
	r.HandleFunc("/test-case/update", api.UpdateTestCaseHandler).Methods("PUT")
	r.HandleFunc("/test-case/delete", api.DeleteTestCaseHandler).Methods("DELETE")

	//test Runs
	// Feature flag decision: test run API is intentionally not part of the public contract yet.
	// Keep DB schema in place, but do not register routes until handlers and validation are complete.
	//r.HandleFunc("/testrun", api.CreateTestRunHandler).Methods("POST")

	//test suites
	r.HandleFunc("/test-suites", api.GetAllTestSuitesHandler).Methods("GET")
	r.HandleFunc("/test-suites", api.CreateTestSuiteHandler).Methods("POST")
	r.HandleFunc("/test-suites/add-cases", api.AddTestCasesToSuiteHandler).Methods("POST")
	r.HandleFunc("/test-suite", api.GetTestSuiteByIDHandler).Methods("GET")
	r.HandleFunc("/test-suite/update", api.UpdateTestSuiteHandler).Methods("PUT")
	r.HandleFunc("/test-suite/delete", api.DeleteTestSuiteHandler).Methods("DELETE")
	r.HandleFunc("/test-suite/remove-case", api.RemoveTestCaseFromSuiteHandler).Methods("DELETE")

	// CORS
	corsHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		// AllowedHeaders appends values. Keep a single explicit list to avoid accidental invalid header names.
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization", "Ngrok-Skip-Browser-Warning"}),
	)(r)

	log.Println("Server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", corsHandler))
}
