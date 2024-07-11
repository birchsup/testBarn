package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func GetTestCaseCLI(testCaseID string) {
	url := fmt.Sprintf("http://127.0.0.1:8081/testcases/%s", testCaseID)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Ошибка при отправке запроса:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Ошибка при получении тест-кейса:", resp.Status)
		return
	}

	var testCase map[string]interface{}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Ошибка при чтении ответа:", err)
		return
	}
	if err := json.Unmarshal(body, &testCase); err != nil {
		fmt.Println("Ошибка при декодировании ответа:", err)
		return
	}

	fmt.Printf("Получен тест-кейс: %+v\n", testCase)
}
