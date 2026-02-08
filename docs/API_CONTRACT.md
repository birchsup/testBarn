# API_CONTRACT

## 1) Обзор API

### Base path
- Явный base path/version prefix отсутствует.
- Все маршруты зарегистрированы от корня (`/`) в `main.go`.

### Формат ответов
- Общего envelope (типа `{ "success": ..., "data": ..., "error": ... }`) нет.
- Успешные ответы:
  - чаще всего сырой JSON сущности/массива;
  - для части операций только HTTP-статус без тела.
- Ошибки:
  - формируются через `http.Error(...)`;
  - тело ошибки — plain text строка (не JSON), обычно с переводом строки.

### Аутентификация / заголовки
- Аутентификация в коде не реализована (middleware/auth отсутствуют).
- CORS включён глобально.
- Разрешённые методы: `GET, POST, PUT, DELETE, OPTIONS`.
- Разрешённые заголовки задаются двумя вызовами `AllowedHeaders(...)` (см. спорные моменты).
- В handlers test-cases выставляется заголовок ответа `ngrok-skip-browser-warning: true`.

## 2) Эндпоинты (полный список)

### 2.1 Active endpoints (из `main.go`)

| Method | Path | Назначение | Где в коде |
|---|---|---|---|
| POST | `/testcases` | Создать test case | `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/main.go:20`, `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/internal/api/handlers.go:11` |
| GET | `/testcase` | Получить test case по `id` (query) | `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/main.go:21`, `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/internal/api/handlers.go:30` |
| GET | `/testcases` | Получить список test cases | `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/main.go:22`, `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/internal/api/handlers.go:52` |
| PUT | `/test-case/update` | Обновить test case по `id` (query) | `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/main.go:23`, `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/internal/api/handlers.go:64` |
| DELETE | `/test-case/delete` | Удалить test case по `id` (query) | `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/main.go:24`, `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/internal/api/handlers.go:97` |
| GET | `/test-suites` | Получить список test suites | `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/main.go:30`, `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/internal/api/testSuite.go:155` |
| POST | `/test-suites` | Создать test suite | `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/main.go:31`, `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/internal/api/testSuite.go:10` |
| POST | `/test-suites/add-cases` | Привязать список test cases к suite | `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/main.go:32`, `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/internal/api/testSuite.go:28` |
| GET | `/test-suite` | Получить suite по `id` (query) вместе с `test_cases` | `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/main.go:33`, `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/internal/api/testSuite.go:45` |
| PUT | `/test-suite/update` | Обновить suite по `id` (query) | `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/main.go:34`, `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/internal/api/testSuite.go:64` |
| DELETE | `/test-suite/delete` | Удалить suite по `id` (query) | `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/main.go:35`, `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/internal/api/testSuite.go:94` |
| DELETE | `/test-suite/remove-case` | Удалить связь suite-case по `suite_id` и `case_id` | `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/main.go:36`, `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/internal/api/testSuite.go:116` |

### 2.2 Inactive / commented endpoint

| Method | Path | Назначение | Статус | Где в коде |
|---|---|---|---|---|
| POST | `/testrun` | Создать test run | Маршрут закомментирован | `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/main.go:27`, `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/internal/api/testRun.go:13` |

---

### Endpoint details

#### POST `/testcases`
- Назначение: создать запись test case.
- Параметры: path/query отсутствуют.
- Request body (фактически принимается `db.TestCase`):
  - JSON schema (упрощённо):
    - `id` (integer, optional, игнорируется при вставке)
    - `test` (object/array/any valid JSON, required по БД)
    - `suite_id` (nullable object при сериализации `sql.NullInt64`, optional)
    - `suite_name` (nullable object при сериализации `sql.NullString`, optional)
- Пример:
```json
{
  "test": {
    "name": "Login Test",
    "steps": [
      {"step": 1, "action": "Open login page", "expected_result": "Login page is displayed"}
    ]
  }
}
```
- Response:
  - Success: `200 OK`, тело — созданный `TestCase` с `id`.
  - Errors:
    - `400 Bad Request` — невалидный JSON.
    - `500 Internal Server Error` — ошибка вставки.
- Связь с БД:
  - `INSERT INTO test_cases (test)`.
  - Таблица: `test_cases`.
- Где в коде:
  - Handler: `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/internal/api/handlers.go:11`
  - DB: `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/db/db.go:29`

#### GET `/testcase?id={id}`
- Назначение: получить один test case + информацию о suite (если есть связь).
- Параметры:
  - Query: `id` (required, string; явная numeric-валидация в handler отсутствует).
- Request body: отсутствует.
- Response:
  - Success: `200 OK`, JSON `TestCase`.
  - Errors:
    - `400 Bad Request` — если `id` отсутствует.
    - `404 Not Found` — если запись не найдена (`sql.ErrNoRows`).
    - `500 Internal Server Error` — прочие ошибки БД.
- Пример success (фактическая структура из-за `sql.Null*`):
```json
{
  "id": 1,
  "test": {"name": "Login Test"},
  "suite_id": {"Int64": 0, "Valid": false},
  "suite_name": {"String": "", "Valid": false}
}
```
- Связь с БД:
  - `SELECT ... FROM test_cases LEFT JOIN test_suite_cases LEFT JOIN test_suites WHERE tc.id=$1`.
  - Таблицы: `test_cases`, `test_suite_cases`, `test_suites`.
- Где в коде:
  - Handler: `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/internal/api/handlers.go:30`
  - DB: `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/db/db.go:38`

#### GET `/testcases`
- Назначение: получить все test cases.
- Параметры: отсутствуют.
- Request body: отсутствует.
- Response:
  - Success: `200 OK`, JSON массив `[]TestCase` (фактически только `id` и `test` заполняются).
  - Errors: `500 Internal Server Error`.
- Связь с БД:
  - `SELECT id, test FROM test_cases`.
  - Таблица: `test_cases`.
- Где в коде:
  - Handler: `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/internal/api/handlers.go:52`
  - DB: `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/db/db.go:53`

#### PUT `/test-case/update?id={id}`
- Назначение: обновить поле `test` у test case.
- Параметры:
  - Query: `id` (required, integer).
- Request body:
  - ожидается JSON с полем `test`.
- Пример:
```json
{
  "test": {
    "name": "Updated Login Test"
  }
}
```
- Response:
  - Success: `200 OK`, без тела.
  - Errors:
    - `400 Bad Request` — отсутствует/некорректный `id`, невалидный JSON.
    - `500 Internal Server Error` — ошибка БД.
- Связь с БД:
  - `UPDATE test_cases SET test=$1 WHERE id=$2`.
  - Таблица: `test_cases`.
- Где в коде:
  - Handler: `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/internal/api/handlers.go:64`
  - DB: `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/db/db.go:77`

#### DELETE `/test-case/delete?id={id}`
- Назначение: удалить test case и его связи с suite.
- Параметры:
  - Query: `id` (required, integer).
- Request body: отсутствует.
- Response:
  - Success: `200 OK`, JSON `{ "message": "Test case was deleted" }`.
  - Errors:
    - `400 Bad Request` — отсутствует/некорректный `id`.
    - `500 Internal Server Error` — ошибка БД.
- Связь с БД:
  - `DELETE FROM test_suite_cases WHERE case_id=$1`.
  - `DELETE FROM test_cases WHERE id=$1`.
  - Таблицы: `test_suite_cases`, `test_cases`.
- Где в коде:
  - Handler: `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/internal/api/handlers.go:97`
  - DB: `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/db/db.go:82`

#### GET `/test-suites`
- Назначение: получить список test suites.
- Параметры: отсутствуют.
- Request body: отсутствует.
- Response:
  - Success: `200 OK`, JSON массив `[]TestSuite` (`id`, `name`, `description`, `created_at`).
  - Errors: `500 Internal Server Error`.
- Связь с БД:
  - `SELECT id, name, description, created_at FROM test_suites`.
  - Таблица: `test_suites`.
- Где в коде:
  - Handler: `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/internal/api/testSuite.go:155`
  - DB: `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/db/testSuite.go:137`

#### POST `/test-suites`
- Назначение: создать test suite.
- Параметры: отсутствуют.
- Request body (`db.TestSuiteRequest`):
  - `name` (string)
  - `description` (string)
- Пример:
```json
{
  "name": "Smoke Suite",
  "description": "Critical smoke checks"
}
```
- Response:
  - Success: `201 Created`, JSON `TestSuite`.
  - Errors:
    - `400 Bad Request` — невалидный JSON.
    - `500 Internal Server Error` — ошибка БД.
- Связь с БД:
  - `INSERT INTO test_suites (name, description) RETURNING id, name, description, created_at`.
  - Таблица: `test_suites`.
- Где в коде:
  - Handler: `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/internal/api/testSuite.go:10`
  - DB: `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/db/testSuite.go:28`

#### POST `/test-suites/add-cases`
- Назначение: назначить набор test cases конкретному suite.
- Параметры: отсутствуют.
- Request body (`db.AddTestCaseRequest`):
  - `suite_id` (integer)
  - `case_ids` (array of integer)
- Пример:
```json
{
  "suite_id": 10,
  "case_ids": [1, 2, 3]
}
```
- Response:
  - Success: `200 OK`, без тела.
  - Errors:
    - `400 Bad Request` — невалидный JSON.
    - `500 Internal Server Error` — ошибка БД/foreign key.
- Связь с БД:
  - сначала удаляются все связи этих `case_ids` с любыми suite;
  - затем создаются связи `(suite_id, case_id)`.
  - Таблица: `test_suite_cases`.
- Где в коде:
  - Handler: `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/internal/api/testSuite.go:28`
  - DB: `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/db/testSuite.go:38`

#### GET `/test-suite?id={id}`
- Назначение: получить suite и вложенный массив `test_cases`.
- Параметры:
  - Query: `id` (required).
- Request body: отсутствует.
- Response:
  - Success: `200 OK`, JSON `TestSuite`.
  - Errors:
    - `400 Bad Request` — отсутствует `id`.
    - `500 Internal Server Error` — любые ошибки, включая `sql.ErrNoRows`.
- Пример success:
```json
{
  "id": 10,
  "name": "Smoke Suite",
  "description": "Critical smoke checks",
  "created_at": "2026-02-08T16:00:00Z",
  "test_cases": [
    {"id": 1, "test": {"name": "Login Test"}}
  ]
}
```
- Связь с БД:
  - `SELECT ... FROM test_suites LEFT JOIN test_suite_cases LEFT JOIN test_cases WHERE ts.id=$1`.
  - Таблицы: `test_suites`, `test_suite_cases`, `test_cases`.
- Где в коде:
  - Handler: `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/internal/api/testSuite.go:45`
  - DB: `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/db/testSuite.go:55`

#### PUT `/test-suite/update?id={id}`
- Назначение: обновить `name`/`description` suite.
- Параметры:
  - Query: `id` (required, integer).
- Request body (`db.TestSuiteRequest`):
```json
{
  "name": "Regression Suite",
  "description": "Nightly regression"
}
```
- Response:
  - Success: `200 OK`, JSON обновлённого suite (`id`, `name`, `description`).
  - Errors:
    - `400 Bad Request` — невалидный JSON / отсутствует / нечисловой `id`.
    - `500 Internal Server Error` — ошибка БД, включая отсутствие строки (через `RETURNING`).
- Связь с БД:
  - `UPDATE test_suites SET ... WHERE id=$3 RETURNING id,name,description`.
  - Таблица: `test_suites`.
- Где в коде:
  - Handler: `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/internal/api/testSuite.go:64`
  - DB: `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/db/testSuite.go:108`

#### DELETE `/test-suite/delete?id={id}`
- Назначение: удалить suite.
- Параметры:
  - Query: `id` (required, integer).
- Request body: отсутствует.
- Response:
  - Success: `204 No Content`.
  - Errors:
    - `400 Bad Request` — отсутствует/нечисловой `id`.
    - `500 Internal Server Error` — ошибка БД.
- Связь с БД:
  - `DELETE FROM test_suites WHERE id=$1`.
  - Таблица: `test_suites`.
  - Важный нюанс: в БД нет `ON DELETE CASCADE` для `test_suite_cases`, поэтому удаление suite с привязанными cases может падать по FK.
- Где в коде:
  - Handler: `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/internal/api/testSuite.go:94`
  - DB: `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/db/testSuite.go:119`

#### DELETE `/test-suite/remove-case?suite_id={suiteId}&case_id={caseId}`
- Назначение: удалить связь suite-case.
- Параметры:
  - Query: `suite_id` (required, integer)
  - Query: `case_id` (required, integer)
- Request body: отсутствует.
- Response:
  - Success: `204 No Content`.
  - Errors:
    - `400 Bad Request` — отсутствуют/некорректны параметры.
    - `500 Internal Server Error` — ошибка БД.
- Связь с БД:
  - `DELETE FROM test_suite_cases WHERE suite_id=$1 AND case_id=$2`.
  - Таблица: `test_suite_cases`.
- Где в коде:
  - Handler: `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/internal/api/testSuite.go:116`
  - DB: `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/db/testSuite.go:128`

### 2.3 Что проверяют интеграционные тесты

| Тест | Что проверяет | Где |
|---|---|---|
| `TestCreateAndGetTestCase` | `POST /testcases` и `GET /testcases/{id}` (path-param контракт) | `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/tests_test/integrations/main_test.go:146` |
| `TestTableTestCases` | Наличие таблиц `test_cases`, `test_suites`, `test_suite_cases`, `test_runs`, `test_run_cases` после миграций | `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/tests_test/integrations/main_test.go:209` |

Примечание: `TestCreateAndGetTestCase` в текущем виде не компилируется из-за ссылки на отсутствующий handler `api.GetTestCase`.

### 2.4 Связь API и схемы БД (миграции)

Таблицы из миграций:
- `test_cases` (`id`, `test JSONB NOT NULL`) — `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/db/migrations/000001_create_test_cases_table.up.sql:1`
- `test_suites` — `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/db/migrations/000002_create_test_tables.up.sql:2`
- `test_suite_cases` (M:N suites-cases) — `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/db/migrations/000002_create_test_tables.up.sql:9`
- `test_runs` — `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/db/migrations/000002_create_test_tables.up.sql:15`
- `test_run_cases` (M:N runs-cases) — `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/db/migrations/000002_create_test_tables.up.sql:22`

Покрытие таблиц API-эндпоинтами:
- Используются активно: `test_cases`, `test_suites`, `test_suite_cases`.
- Не используются активными эндпоинтами: `test_runs`, `test_run_cases`.

## 3) Спорные моменты и несоответствия

1. Расхождение роутов приложения и интеграционного теста для получения test case.
- Приложение: `GET /testcase?id=...` (`GetTestCaseHandler`).
- Тест: `GET /testcases/{id}` + `api.GetTestCase`.
- Факт: `api.GetTestCase` отсутствует, `go test ./...` падает на compile-time ошибке.
- Ссылки: `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/main.go:21`, `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/tests_test/integrations/main_test.go:149`.

2. Два разных паттерна контракта одновременно в API.
- Есть коллекционный путь `/testcases`.
- Но single-resource операции сделаны в стиле `query + action-path` (`/testcase?id=...`, `/test-case/update?id=...`, `/test-case/delete?id=...`).
- Для suites аналогично: `/test-suites` vs `/test-suite?id=...`.
- Это делает API неоднородным и затрудняет предсказуемость клиента.

3. Неоднозначные статусы/ошибки для not found.
- `GetTestCaseHandler` корректно мапит `sql.ErrNoRows` в `404`.
- `GetTestSuiteByIDHandler` возвращает `500` даже при `sql.ErrNoRows` из DB-слоя.
- Update/Delete для cases/suites не проверяют `RowsAffected`, поэтому «удалено/обновлено» может вернуть success даже если id не существовал.

4. Несогласованный формат ошибок и успешных ответов.
- Ошибки всегда plain text через `http.Error`, не JSON.
- Часть успешных операций возвращает JSON, часть только статус (`PUT /test-case/update`, `POST /test-suites/add-cases`, `DELETE /test-suite/delete`, `DELETE /test-suite/remove-case`).
- Нет единого response envelope и единой схемы ошибок.

5. Риск некорректного JSON-контракта из-за `sql.Null*` в `TestCase`.
- `suite_id` и `suite_name` сериализуются как структурные объекты (`{"Int64":...,"Valid":...}` / `{"String":...,"Valid":...}`), а не как простые nullable-поля.
- Для публичного API это обычно нежелательный leakage внутреннего типа БД.
- Ссылка: `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/db/db.go:22`.

6. CORS-конфигурация: двойной `AllowedHeaders(...)`.
- В `main.go` два вызова подряд; поведение (перезапись/объединение) неочевидно.
- Дополнительно `"true"` указан как имя заголовка, что выглядит как ошибка конфигурации.
- Ссылка: `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/main.go:42`.

7. Частично реализованная функциональность `test_runs`.
- Таблицы `test_runs`/`test_run_cases` есть в БД и проверяются тестом.
- Route и handler/db-функция создания test run закомментированы.
- В результате БД-контракт шире, чем публичный API-контракт.
- Ссылки: `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/main.go:27`, `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/internal/api/testRun.go:13`, `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/db/testRun.go:15`.

8. Интеграционные тесты завязаны на абсолютный путь миграций.
- Используется `file:///Users/dmitrijsadovnikov/testBarn/db/migrations`.
- Тесты непереносимы между машинами/путями.
- Ссылка: `/Users/dmitrijsadovnikov/.codex/worktrees/ffa5/testBarn/tests_test/integrations/main_test.go:73`.

9. Неполный CRUD по домену `test_runs`.
- БД поддерживает сущности run/run_cases, но нет активных endpoint-ов create/read/update/delete.
- Для домена test suites нет явного endpoint-а «получить все cases конкретного suite» отдельно от `GET /test-suite?id=...` (есть только вложенное чтение в одном запросе).

## 4) Рекомендации

### Канонический контракт (v1)
1. Принять resource-oriented стиль и path params как канонический:
- `POST /testcases`
- `GET /testcases`
- `GET /testcases/{id}`
- `PUT /testcases/{id}`
- `DELETE /testcases/{id}`
- `GET /test-suites`
- `POST /test-suites`
- `GET /test-suites/{id}`
- `PUT /test-suites/{id}`
- `DELETE /test-suites/{id}`
- `POST /test-suites/{id}/cases` (bulk replace/add, явно определить семантику)
- `DELETE /test-suites/{id}/cases/{case_id}`

2. Стандартизовать ответы:
- ошибки только JSON (`code`, `message`, `details`);
- единые статусы для not found (`404`), валидации (`400`/`422`), конфликтов (`409`), сервера (`500`).

3. Привести DTO к API-модели и убрать `sql.Null*` из публичного JSON.

### Что исправить в первую очередь (стабилизация контракта)
1. Устранить рассинхрон `GET test case`: выбрать один контракт и синхронизировать `main.go`, handlers и интеграционный тест.
2. Починить сборку тестов (`api.GetTestCase` отсутствует).
3. Выравнять not-found поведение (`GetTestSuiteByID` -> `404` при `sql.ErrNoRows`).
4. Убрать абсолютный путь миграций в тестах.
5. Определить и зафиксировать поведение `DELETE/PUT` при несуществующем `id` (через `RowsAffected` + `404`).

### Что отложить в v2
1. Полный запуск домена `test_runs` (CRUD + бизнес-семантика run-cases).
2. Рефакторинг CORS и инфраструктурный hardening (после фикса основного контракта).
3. Введение версионирования API (`/api/v1`) и возможной обратной совместимости со старыми маршрутами.

## 5) Update: MVP Test Runs (Suite-based)

Ниже зафиксированы новые активные методы для MVP test runs, включая создание run из `test suite`.

### 5.1 Active endpoints (MVP test runs)

| Method | Path | Назначение |
|---|---|---|
| POST | `/test-runs` | Создать test run из `suite_id`, списка `test_case_ids` или их комбинации |
| GET | `/test-runs` | Получить список test runs |
| GET | `/test-runs/{id}` | Получить детали test run (включая cases и summary) |
| PATCH | `/test-runs/{runId}/cases/{caseId}` | Обновить статус/комментарий конкретного case в run |

### 5.2 POST `/test-runs`
- Назначение: создать новый run.
- Поддерживаемые источники кейсов:
  - `suite_id` (все кейсы из указанного test suite),
  - `test_case_ids` (явно переданный список),
  - комбинация `suite_id + test_case_ids`.
- Request body:
```json
{
  "suite_id": 10,
  "test_case_ids": [1, 2, 3],
  "run_details": {"name": "Regression run"},
  "executed_by": "qa.user"
}
```
- Правила:
  - должен быть указан хотя бы один источник (`suite_id` или `test_case_ids`);
  - если `suite_id` не существует -> `404`;
  - если хотя бы один `test_case_id` не существует -> `404`;
  - повторяющиеся `test_case_ids` дедуплицируются;
  - материализация run-case записей выполняется в `test_run_cases` со статусом `not_run`.
- Response:
  - Success: `201 Created`, возвращается run details (`id`, `suite_id`, `cases`, `summary`, `created_at`);
  - Errors:
    - `400 Bad Request` — невалидный payload/ID или отсутствует источник;
    - `404 Not Found` — `suite`/`case` не найден.

### 5.3 GET `/test-runs`
- Назначение: получить список run-ов.
- Response:
  - Success: `200 OK`, JSON-массив run-объектов.

### 5.4 GET `/test-runs/{id}`
- Назначение: получить один run с кейсами и итоговой сводкой.
- Response:
  - Success: `200 OK`, объект run:
    - `cases[]` с `case_id`, `status`, `comment`, `executed_at`, `executed_by`;
    - `summary` с полями `passed`, `failed`, `blocked`, `skipped`, `not_run`.
  - Errors:
    - `400 Bad Request` — невалидный `id`;
    - `404 Not Found` — run не найден.

### 5.5 PATCH `/test-runs/{runId}/cases/{caseId}`
- Назначение: обновить статус и комментарий кейса в run.
- Request body:
```json
{
  "status": "passed",
  "comment": "Executed successfully",
  "executed_by": "qa.user"
}
```
- Разрешённые статусы: `passed`, `failed`, `blocked`, `skipped`, `not_run`.
- Response:
  - Success: `200 OK`, возвращается обновлённый run details с пересчитанной `summary`;
  - Errors:
    - `400 Bad Request` — невалидные `runId/caseId` или `status`;
    - `404 Not Found` — run или run-case не найден.
