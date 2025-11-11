# Task Completion Checklist
- Ensure code follows layered architecture, uses Manager-provided resources, and updates shared enums/errors/log messages as needed.
- Run `make fmt`, `make lint`, and `make test` (plus targeted `make migrate-*` or `make swag` when schema/API changes occur). Remove any temporary test files afterwards.
- For API changes: regenerate Swagger via `make swag`, update `api-test.http`, and confirm handlers return Base responses (HTTP 200, code field conveys errors). Verify middlewares/permissions via ChainBuilder and Scope middleware notes.
- For schema/model changes: add matching up/down SQL in `internal/database/migrations` before touching models/services.
- Check configs (static vs dynamic) remain in the correct location; avoid introducing new config paths unless approved.
- Before handing off, make sure logs/comments are Chinese, no fmt.Printf usages, no stray build artifacts outside `build/`, and that time handling uses UTC storage with Asia/Shanghai presentation (use `date` command for real time).