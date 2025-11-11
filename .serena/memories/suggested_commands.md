# Suggested Commands
- `make run` / `make build` / `make build-prod`: compile and run the Gin service (output binary in build/).
- `make test`, `make coverage`: run Go tests with race + cover, optionally emit HTML coverage.
- `make fmt`, `make lint`: format Go sources and run golangci-lint.
- `make migrate-up`, `make migrate-down`, `make migrate-steps STEPS=N`, `make migrate-force VERSION=N`, `make migrate-reset`: manage golang-migrate migrations through the app binary.
- `make swag`: regenerate Swagger docs (also syncs to web/api-docs/); required after API updates.
- `make deps`, `make clean`, `make version`, `make help`: dependency install, clean build/, view version metadata, list all targets.
- Common Linux tooling: `ls`, `find`/`rg --files`, `rg` for search, `sed`/`awk` for inspection, `date` for real-time timestamps, `git status/diff` for repo state.
- Entry binary: `./build/vaulthub serve --config configs/config.toml` (or other subcommands like `migrate`).