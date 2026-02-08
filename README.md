# testBarn

## Run backend + database in Docker

```bash
docker compose up --build
```

This starts:
- `db` (PostgreSQL)
- `migrate` (applies SQL migrations)
- `backend` (Go API on port `8080`)

API URL:
- `http://localhost:8080`

## Run tests in Docker

```bash
docker compose --profile test run --rm tests
```

The `tests` container runs `go test -v ./...` against the Docker Postgres service.

## Stop everything

```bash
docker compose down -v
```
