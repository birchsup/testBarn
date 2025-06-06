package db

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/net/context"
	"time"
)

type TestReport struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	Report    map[string]any `gorm:"type:jsonb" json:"report"`
}
type TestReportRequest struct {
	Report json.RawMessage `json:"report"`
}

func CreateTestReport(report json.RawMessage) (TestReport, error) {
	// Очистка отчёта: удаление stdout, stderr, attachments
	cleanedReport, err := cleanTestReport(report)
	if err != nil {
		return TestReport{}, errors.New("ошибка при обработке JSON-отчета: " + err.Error())
	}

	// Логируем очищенный отчёт перед сохранением
	fmt.Println("📝 Очищенный отчёт:", string(cleanedReport))

	query := `INSERT INTO test_reports (report) VALUES ($1::jsonb) RETURNING id, created_at, report`
	var testReport TestReport
	err = DBPool.QueryRow(context.Background(), query, cleanedReport).Scan(&testReport.ID, &testReport.CreatedAt, &testReport.Report)
	if err != nil {
		return TestReport{}, fmt.Errorf("ошибка сохранения в БД: %w", err)
	}

	return testReport, nil
}

// cleanTestReport - Очищает JSON-отчёт от ненужных данных
func cleanTestReport(report json.RawMessage) (json.RawMessage, error) {
	var parsedReport map[string]interface{}

	// Попробуем распарсить JSON
	err := json.Unmarshal(report, &parsedReport)
	if err != nil {
		return nil, fmt.Errorf("ошибка разбора JSON: %w", err)
	}

	// Логируем структуру отчёта после парсинга
	fmt.Println("📜 Распознанный JSON:", parsedReport)

	// Проверяем наличие ключа "suites"
	suitesRaw, exists := parsedReport["suites"]
	if !exists {
		return nil, errors.New("отсутствует поле 'suites' в отчёте")
	}

	// Проверяем, является ли "suites" массивом
	suitesArray, ok := suitesRaw.([]interface{})
	if !ok {
		return nil, errors.New("поле 'suites' имеет неправильный формат")
	}

	// Логируем количество сьютов
	fmt.Printf("🔍 Найдено %d suites\n", len(suitesArray))

	// Обход всех suite
	for _, suite := range suitesArray {
		suiteMap, ok := suite.(map[string]interface{})
		if !ok {
			continue
		}

		// Проверяем specs
		specsRaw, specExists := suiteMap["specs"]
		if !specExists {
			continue
		}

		specs, ok := specsRaw.([]interface{})
		if !ok {
			continue
		}

		for _, spec := range specs {
			specMap, ok := spec.(map[string]interface{})
			if !ok {
				continue
			}

			// Проверяем tests
			testsRaw, testExists := specMap["tests"]
			if !testExists {
				continue
			}

			tests, ok := testsRaw.([]interface{})
			if !ok {
				continue
			}

			for _, test := range tests {
				testMap, ok := test.(map[string]interface{})
				if !ok {
					continue
				}

				// Удаляем ненужные поля
				delete(testMap, "stdout")
				delete(testMap, "stderr")
				delete(testMap, "attachments")
			}
		}
	}

	// Преобразуем обратно в JSON
	cleanedJSON, err := json.Marshal(parsedReport)
	if err != nil {
		return nil, fmt.Errorf("ошибка маршализации JSON: %w", err)
	}

	// Логируем размер очищенного JSON
	fmt.Printf("✅ Очищенный JSON, размер: %d байт\n", len(cleanedJSON))

	return cleanedJSON, nil
}
func GetTestReportByID(reportID int) (TestReport, error) {
	var testReport TestReport

	query := `SELECT id, created_at, report FROM test_reports WHERE id = $1`
	err := DBPool.QueryRow(context.Background(), query, reportID).Scan(&testReport.ID, &testReport.CreatedAt, &testReport.Report)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return TestReport{}, err
		}
		return TestReport{}, err
	}

	return testReport, nil
}
func GetAllTestReports() ([]TestReport, error) {
	query := `SELECT id, created_at, report FROM test_reports ORDER BY created_at DESC`
	rows, err := DBPool.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var testReports []TestReport
	for rows.Next() {
		var testReport TestReport
		err := rows.Scan(&testReport.ID, &testReport.CreatedAt, &testReport.Report)
		if err != nil {
			return nil, err
		}
		testReports = append(testReports, testReport)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return testReports, nil
}

func DeleteTestReport(reportID int) error {
	query := `DELETE FROM test_reports WHERE id = $1`
	_, err := DBPool.Exec(context.Background(), query, reportID)
	if err != nil {
		return err
	}
	return nil
}
