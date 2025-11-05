# 项目配置
BINARY_NAME=vaulthub
BUILD_DIR=build
CMD_DIR=cmd/vaulthub

# 配置文件路径（可选，默认使用应用内置的 configs/config.toml）
CONFIG ?=

# 构建配置参数
ifneq ($(CONFIG),)
    CONFIG_FLAG=--config $(CONFIG)
else
    CONFIG_FLAG=
endif

# 版本信息
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S_UTC' 2>/dev/null || echo "unknown")
GO_VERSION=$(shell go version | awk '{print $$3}')

# ldflags 注入版本信息
LDFLAGS=-ldflags "\
	-X 'github.com/cuihe500/vaulthub/pkg/version.Version=$(VERSION)' \
	-X 'github.com/cuihe500/vaulthub/pkg/version.GitCommit=$(GIT_COMMIT)' \
	-X 'github.com/cuihe500/vaulthub/pkg/version.BuildTime=$(BUILD_TIME)'"

# 默认目标
.PHONY: all
all: build

# 构建
.PHONY: build
build:
	@echo "Building $(BINARY_NAME) version $(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./$(CMD_DIR)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# 构建（生产环境）
.PHONY: build-prod
build-prod:
	@echo "Building $(BINARY_NAME) for production..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./$(CMD_DIR)
	@echo "Production build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# 运行
.PHONY: run
run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME) serve $(CONFIG_FLAG)

# 清理
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@echo "Clean complete"

# 测试
.PHONY: test
test:
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...
	@echo "Tests complete"

# 测试覆盖率
.PHONY: coverage
coverage: test
	@echo "Generating coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# 格式化代码
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...
	@echo "Format complete"

# 代码检查
.PHONY: lint
lint:
	@echo "Running linter..."
	golangci-lint run ./...
	@echo "Lint complete"

# 安装依赖
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy
	@echo "Dependencies installed"

# 版本信息
.PHONY: version
version:
	@echo "Version:    $(VERSION)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Go Version: $(GO_VERSION)"

# 数据库迁移
.PHONY: migrate-up
migrate-up: build
	@echo "Applying all pending migrations..."
	./$(BUILD_DIR)/$(BINARY_NAME) migrate up $(CONFIG_FLAG)
	@echo "Migration complete"

.PHONY: migrate-down
migrate-down: build
	@echo "Rolling back last migration..."
	./$(BUILD_DIR)/$(BINARY_NAME) migrate down $(CONFIG_FLAG)
	@echo "Rollback complete"

.PHONY: migrate-version
migrate-version: build
	@echo "Getting current migration version..."
	./$(BUILD_DIR)/$(BINARY_NAME) migrate version $(CONFIG_FLAG)

.PHONY: migrate-steps
migrate-steps: build
	@if [ -z "$(STEPS)" ]; then \
		echo "Error: STEPS parameter is required"; \
		echo "Usage: make migrate-steps STEPS=N (positive for up, negative for down)"; \
		exit 1; \
	fi
	@echo "Migrating $(STEPS) steps..."
	./$(BUILD_DIR)/$(BINARY_NAME) migrate steps -n $(STEPS) $(CONFIG_FLAG)
	@echo "Migration complete"

.PHONY: migrate-force
migrate-force: build
	@if [ -z "$(VERSION)" ]; then \
		echo "Error: VERSION parameter is required"; \
		echo "Usage: make migrate-force VERSION=N"; \
		exit 1; \
	fi
	@echo "WARNING: Force setting migration version to $(VERSION)"
	@echo "This may cause data inconsistency. Continue? [y/N] " && read ans && [ $${ans:-N} = y ]
	./$(BUILD_DIR)/$(BINARY_NAME) migrate force -v $(VERSION) $(CONFIG_FLAG)
	@echo "Version forced to $(VERSION)"

.PHONY: migrate-reset
migrate-reset: build
	@echo "WARNING: This will reset the database and destroy all data!"
	@echo "Continue? [y/N] " && read ans && [ $${ans:-N} = y ]
	@echo "Rolling back all migrations..."
	./$(BUILD_DIR)/$(BINARY_NAME) migrate steps -n -9999 $(CONFIG_FLAG) || true
	@echo "Applying all migrations..."
	./$(BUILD_DIR)/$(BINARY_NAME) migrate up $(CONFIG_FLAG)
	@echo "Database reset complete"

# 帮助
.PHONY: help
help:
	@echo "VaultHub Makefile Commands:"
	@echo ""
	@echo "Build & Run:"
	@echo "  make build       - Build the binary"
	@echo "  make build-prod  - Build for production (Linux/amd64)"
	@echo "  make run         - Build and run the application"
	@echo "  make clean       - Remove build artifacts"
	@echo ""
	@echo "Testing & Quality:"
	@echo "  make test        - Run tests"
	@echo "  make coverage    - Generate test coverage report"
	@echo "  make fmt         - Format code"
	@echo "  make lint        - Run linter"
	@echo ""
	@echo "Database Migration:"
	@echo "  make migrate-up      - Apply all pending migrations"
	@echo "  make migrate-down    - Rollback last migration"
	@echo "  make migrate-version - Show current migration version"
	@echo "  make migrate-steps STEPS=N - Migrate N steps (positive=up, negative=down)"
	@echo "  make migrate-force VERSION=N - Force set migration version (use with caution)"
	@echo "  make migrate-reset   - Reset database (WARNING: destroys all data)"
	@echo ""
	@echo "Others:"
	@echo "  make deps        - Install dependencies"
	@echo "  make version     - Show version information"
	@echo "  make help        - Show this help message"
	@echo ""
	@echo "Global Parameters:"
	@echo "  CONFIG=path/to/config.toml - Specify custom config file path"
	@echo "  Example: make run CONFIG=/etc/vaulthub/production.toml"
