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
	r.HandleFunc("/testcases", api.CreateTestCase).Methods("POST")
	r.HandleFunc("/testcases/{id:[0-9]+}", api.GetTestCase).Methods("GET")
	r.HandleFunc("/testcases", api.GetAllTestCases).Methods("GET")
	r.HandleFunc("/testrun", api.CreateTestRunHandler).Methods("POST")
	r.HandleFunc("/test-suites", api.CreateTestSuiteHandler).Methods("POST")
	r.HandleFunc("/test-suites/add-cases", api.AddTestCasesToSuiteHandler).Methods("POST")

	// Настройка CORS
	corsHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
	)(r)

	log.Println("Server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", corsHandler))
}
