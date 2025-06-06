package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"testBarn/db"
)

// POST /api/test-reports — загрузка отчета Playwright
func CreateTestReportHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var request db.TestReportRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Ошибка разбора JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	testReport, err := db.CreateTestReport(request.Report)
	if err != nil {
		http.Error(w, "Ошибка сохранения отчета: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(testReport)
}

// GET /api/test-reports/:id — получение отчета по ID
func GetTestReportHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Получаем `id` из URL-параметров
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Не указан ID отчета", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Некорректный ID", http.StatusBadRequest)
		return
	}

	testReport, err := db.GetTestReportByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Отчет не найден", http.StatusNotFound)
			return
		}
		http.Error(w, "Ошибка получения отчета: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(testReport)
}

// GET /api/test-reports — получение всех отчетов
func GetAllTestReportsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	testReports, err := db.GetAllTestReports()
	if err != nil {
		http.Error(w, "Ошибка получения отчетов: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(testReports)
}

// DELETE /api/test-reports/:id — удаление отчета по ID
func DeleteTestReportHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Получаем `id` из URL-параметров
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Не указан ID отчета", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Некорректный ID", http.StatusBadRequest)
		return
	}

	err = db.DeleteTestReport(id)
	if err != nil {
		http.Error(w, "Ошибка удаления отчета: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Отчет удален"})
}
