package main

import (
	"fmt"
	"github.com/rivo/tview"
	"testBarn/internal/api"
)

func main() {
	app := tview.NewApplication()

	rootNode := tview.NewTreeNode("Начало работы").
		AddChild(tview.NewTreeNode("Просмотр").
			AddChild(tview.NewTreeNode("Открыть тестовый набор").
				AddChild(tview.NewTreeNode("Открыть тест-кейс").SetReference("open_test_case")).
				AddChild(tview.NewTreeNode("Открыть тестовый прогон").SetReference("open_test_run")))).
		AddChild(tview.NewTreeNode("Создание чего-либо").
			AddChild(tview.NewTreeNode("Создать тестовый набор").
				AddChild(tview.NewTreeNode("Создать тест-кейс").SetReference("create_test_case"))).
			AddChild(tview.NewTreeNode("Создать тестовый план").SetReference("create_test_plan")))

	tree := tview.NewTreeView().
		SetRoot(rootNode).
		SetCurrentNode(rootNode)

	tree.SetSelectedFunc(func(node *tview.TreeNode) {
		reference := node.GetReference()
		if reference != nil {
			action := reference.(string)
			switch action {
			case "open_test_case":
				fmt.Println("Открыть тест-кейс")
				api.GetTestCaseCLI("1") // Пример вызова функции из API
			case "open_test_run":
				fmt.Println("Открыть тестовый прогон")
				api.OpenTestRun()
			case "create_test_case":
				fmt.Println("Создать тест-кейс")
				//api.CreateTestCase(http.ResponseWriter, *http.Request)
			case "create_test_plan":
				fmt.Println("Создать тестовый план")
				api.CreateTestPlan()
			}
		}
	})

	if err := app.SetRoot(tree, true).Run(); err != nil {
		panic(err)
	}
}
