package api

import (
	"encoding/json"
)

type CreateTestRunRequest struct {
	SuiteID    int             `json:"suite_id"`
	RunDetails json.RawMessage `json:"run_details"`
	CaseIDs    []int           `json:"case_ids"`
}

//func CreateTestRunHandler(w http.ResponseWriter, r *http.Request) {
//	var createReq CreateTestRunRequest
//
//	// Декодируем JSON запрос в структуру CreateTestRunRequest
//	if err := json.NewDecoder(r.Body).Decode(&createReq); err != nil {
//		http.Error(w, err.Error(), http.StatusBadRequest)
//		return
//	}
//
//	// Вызываем функцию создания тестового запуска
//	testRuns, err := db.CreateTestRun(createReq.SuiteID, createReq.RunDetails)
//	if err != nil {
//		http.Error(w, err.Error(), http.StatusInternalServerError)
//		return
//	}
//
//	// Устанавливаем заголовок ответа и кодируем результат в JSON
//	w.Header().Set("Content-Type", "application/json")
//	json.NewEncoder(w).Encode(testRuns)
//}
