# Copilot Instructions for stldevs

This repository contains the backend and data aggregation services for the St. Louis developers profile site.

## High Level Details

- **Project Type**: Go backend application.
- **Language**: Go (1.25+).
- **Database**: PostgreSQL.
- **Database Access**: Uses `sqlc` to generate type-safe Go code from SQL.
- **Web Framework**: Standard library `net/http` with custom routing/middleware in `web/`.
- **Architecture**:
    - `cmd/`: Application entry points.
    - `web/`: HTTP server and handlers.
    - `db/`: Database interaction and generated code.
    - `aggregator/`: Logic for gathering data (e.g., from GitHub).
    - `migrations/`: Database schema migrations defined in Go.

## Build and Validate

### Prerequisites
- Go
- Docker (for running PostgreSQL)
- `sqlc` (if modifying SQL queries)

### Build
To build all binaries:
```bash
go build ./...
```

### Test
To run tests, use the provided script which spins up a temporary Postgres container:
```bash
./test.sh
```
Or manually if you have a database running:
```bash
go test ./...
```

### Run
To run the main web server:
1.  Ensure a Postgres database is running. You can use `pg.sh` to start a persistent dev container:
    ```bash
    ./pg.sh
    ```
2.  Ensure `config.json` exists in the root.
3.  Run the server:
    ```bash
    go run cmd/stldevs/stldevs.go
    ```

### Database Changes
If you modify files in `db/sql/queries/` or `db/sql/schema.sql`:
1.  Run `sqlc generate` to update the Go code in `db/sqlc/`.
2.  Ensure you add any necessary schema changes to `migrations/` if they need to be applied to existing databases.

## Project Layout

- **Entry Points**:
    - `cmd/stldevs/`: Main web server.
    - `cmd/gather/`: Data gathering tool.
    - `cmd/find/`: Utility to find users.
- **Configuration**:
    - `config.json`: Application configuration (secrets, DB connection).
    - `sqlc.yaml`: Configuration for `sqlc`.
- **Database**:
    - `db/sql/schema.sql`: Database schema definitions.
    - `db/sql/queries/`: SQL queries used by `sqlc`.
    - `db/sqlc/`: Generated Go code for database interaction. Do not edit manually.
    - `migrations/`: Go-based database migrations.
- **Web**:
    - `web/`: Contains handlers, routing, and server logic.
- **Scripts**:
    - `pg.sh`: Starts a development Postgres container.
    - `test.sh`: Runs tests in an isolated environment.

## Common Tasks

- **Adding a new API endpoint**:
    1.  Define the handler in `web/`.
    2.  Register the route in `web/server.go` (or where routes are defined).
    3.  If DB access is needed, add a query to `db/sql/queries/` and run `sqlc generate`.
- **Modifying the Database Schema**:
    1.  Update `db/sql/schema.sql`.
    2.  Add a migration in `migrations/`.
    3.  Run `sqlc generate`.
