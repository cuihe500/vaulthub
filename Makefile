# 项目配置
BINARY_NAME=vaulthub
BUILD_DIR=build
CMD_DIR=cmd/vaulthub

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
	./$(BUILD_DIR)/$(BINARY_NAME) serve

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

# 帮助
.PHONY: help
help:
	@echo "VaultHub Makefile Commands:"
	@echo ""
	@echo "  make build       - Build the binary"
	@echo "  make build-prod  - Build for production (Linux/amd64)"
	@echo "  make run         - Build and run the application"
	@echo "  make clean       - Remove build artifacts"
	@echo "  make test        - Run tests"
	@echo "  make coverage    - Generate test coverage report"
	@echo "  make fmt         - Format code"
	@echo "  make lint        - Run linter"
	@echo "  make deps        - Install dependencies"
	@echo "  make version     - Show version information"
	@echo "  make help        - Show this help message"
