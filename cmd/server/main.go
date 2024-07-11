package main

import (
	"log"
	"net/http"
	"testBarn/internal/api"
	"testBarn/internal/config"
	"testBarn/internal/db"

	"github.com/gorilla/mux"
)

func main() {
	config.InitConfig()
	db.InitDB()
	defer db.DBPool.Close()

	r := mux.NewRouter()
	r.HandleFunc("/testcases", api.CreateTestCase).Methods("POST")
	r.HandleFunc("/testcases/{id:[0-9]+}", api.GetTestCase).Methods("GET")
	r.HandleFunc("/testcases", api.GetAllTestCases).Methods("GET")

	http.Handle("/", r)
	log.Println("Server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
