package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/lib/pq"
)

func contains(s []string, e string) bool {
	m := make(map[string]bool)
	for _, v := range s {
		m[v] = true
	}
	return m[e]
}

// GetTestCasesByTags возвращает тест-кейсы по списку тегов
func GetTestCasesByTags(ctx context.Context, tags []string) ([]TestCase, error) {
	query := `
        SELECT DISTINCT tc.id, tc.test, tsc.suite_id, ts.name
        FROM test_cases tc
        LEFT JOIN test_case_tags tct ON tc.id = tct.case_id
        LEFT JOIN test_suite_cases tsc ON tc.id = tsc.case_id
        LEFT JOIN test_suites ts ON tsc.suite_id = ts.id
        WHERE tct.tags ?| $1
    `

	rows, err := DBPool.Query(ctx, query, pq.Array(tags))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var testCases []TestCase
	for rows.Next() {
		var testCase TestCase
		err = rows.Scan(&testCase.ID, &testCase.Test, &testCase.SuiteID, &testCase.SuiteName)
		if err != nil {
			return nil, err
		}
		testCases = append(testCases, testCase)
	}

	return testCases, nil
}

// AddTagsToTestCase добавляет теги к существующему тест-кейсу
func AddTagsToTestCase(ctx context.Context, caseID int64, tags []string) error {
	tx, err := DBPool.Begin(ctx)
	if err != nil {
		return err
	}

	// Получаем существующие теги
	var existingTags json.RawMessage
	err = tx.QueryRow(ctx, "SELECT tags FROM test_case_tags WHERE case_id = $1", caseID).Scan(&existingTags)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	// Объединяем существующие и новые теги
	var currentTags []string
	if existingTags != nil {
		err = json.Unmarshal(existingTags, &currentTags)
		if err != nil {
			return err
		}
	}

	// Удаляем дубликаты
	tagsSet := make(map[string]bool, len(currentTags)+len(tags))
	for _, tag := range currentTags {
		tagsSet[tag] = true
	}
	for _, tag := range tags {
		tagsSet[tag] = true
	}

	// Преобразуем обратно в массив
	newTags := make([]string, 0, len(tagsSet))
	for tag := range tagsSet {
		newTags = append(newTags, tag)
	}

	// Обновляем теги
	tagsJSON, err := json.Marshal(newTags)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, "INSERT INTO test_case_tags (case_id, tags) VALUES ($1, $2) ON CONFLICT (case_id) DO UPDATE SET tags = $2", caseID, tagsJSON)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// RemoveTagsFromTestCase удаляет теги из тест-кейса
func RemoveTagsFromTestCase(ctx context.Context, caseID int64, tags []string) error {
	tx, err := DBPool.Begin(ctx)
	if err != nil {
		return err
	}

	// Получаем существующие теги
	var existingTags json.RawMessage
	err = tx.QueryRow(ctx, "SELECT tags FROM test_case_tags WHERE case_id = $1", caseID).Scan(&existingTags)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	// Парсим существующие теги
	var currentTags []string
	if existingTags != nil {
		err = json.Unmarshal(existingTags, &currentTags)
		if err != nil {
			return err
		}
	}

	// Удаляем указанные теги
	var newTags []string
	for _, tag := range currentTags {
		if !contains(tags, tag) {
			newTags = append(newTags, tag)
		}
	}

	// Обновляем теги
	tagsJSON, err := json.Marshal(newTags)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, "INSERT INTO test_case_tags (case_id, tags) VALUES ($1, $2) ON CONFLICT (case_id) DO UPDATE SET tags = $2", caseID, tagsJSON)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// GetTagsForTestCase возвращает все теги для тест-кейса
func GetTagsForTestCase(ctx context.Context, caseID int64) ([]string, error) {
	var tags json.RawMessage
	err := DBPool.QueryRow(ctx, "SELECT tags FROM test_case_tags WHERE case_id = $1", caseID).Scan(&tags)
	if err != nil {
		if err == sql.ErrNoRows {
			return []string{}, nil
		}
		return nil, err
	}

	var tagList []string
	err = json.Unmarshal(tags, &tagList)
	if err != nil {
		return nil, err
	}

	return tagList, nil
}
