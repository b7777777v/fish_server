# 設定 help 為預設目標，當直接執行 make 時會顯示幫助信息
.DEFAULT_GOAL := help

# 將所有目標聲明為 .PHONY，避免與同名文件衝突
.PHONY: all proto wire gen build build-game build-admin run-game run-admin test lint tidy clean help \
        docker-build run-dev docker-down docker-logs migrate-up migrate-down

## ===================================================================================
## Go & Project Variables
## ===================================================================================

# Go 相關命令
GO            ?= go
GO_BUILD      := $(GO) build
GO_RUN        := $(GO) run
GO_TEST       := $(GO) test
GO_CLEAN      := $(GO) clean
GO_MOD_TIDY   := $(GO) mod tidy

# 專案結構
CMD_DIR       := ./cmd
BIN_DIR       := ./bin

# 二進位檔名稱
GAME_SERVER_NAME  := game-server
ADMIN_SERVER_NAME := admin-server
GAME_SERVER_BIN   := $(BIN_DIR)/$(GAME_SERVER_NAME)
ADMIN_SERVER_BIN  := $(BIN_DIR)/$(ADMIN_SERVER_NAME)

# Docker 相關
DOCKER_COMPOSE_FILE := deployments/docker-compose.yml
DOCKER_COMPOSE      := docker-compose -f $(DOCKER_COMPOSE_FILE)

# 版本資訊 (可由 CI/CD 系統傳入)
VERSION       ?= $(shell git describe --tags --always --dirty)
COMMIT_HASH   ?= $(shell git rev-parse --short HEAD)
BUILD_DATE    ?= $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')

# Go LDFLAGS for embedding version info and shrinking binary size
# -s: 移除符號表
# -w: 移除 DWARF 調試信息
# -X: 設置變量值
LDFLAGS := -s -w \
           -X 'main.Version=$(VERSION)' \
           -X 'main.CommitHash=$(COMMIT_HASH)' \
           -X 'main.BuildDate=$(BUILD_DATE)'

## ===================================================================================
## Code Generation
## ===================================================================================

gen: proto wire ## Generate all necessary code

proto: ## Generate protobuf code
	@echo ">> Generating protobuf code..."
	@sh ./scripts/proto-gen.sh

wire: ## Generate dependency injection code
	@echo ">> Generating wire code..."
	@sh ./scripts/wire-gen.sh

## ===================================================================================
## Build & Run
## ===================================================================================

all: build ## Build all binaries (default)

build: gen ## Build all binaries after generating code
	@echo ">> Building all binaries..."
	@mkdir -p $(BIN_DIR)
	@$(GO_BUILD) -ldflags="$(LDFLAGS)" -o $(GAME_SERVER_BIN) $(CMD_DIR)/game
	@$(GO_BUILD) -ldflags="$(LDFLAGS)" -o $(ADMIN_SERVER_BIN) $(CMD_DIR)/admin
	@echo "✓ Binaries are ready in $(BIN_DIR)"

build-game: gen ## Build only the game server binary
	@echo ">> Building game server binary..."
	@mkdir -p $(BIN_DIR)
	@$(GO_BUILD) -ldflags="$(LDFLAGS)" -o $(GAME_SERVER_BIN) $(CMD_DIR)/game

build-admin: gen ## Build only the admin server binary
	@echo ">> Building admin server binary..."
	@mkdir -p $(BIN_DIR)
	@$(GO_BUILD) -ldflags="$(LDFLAGS)" -o $(ADMIN_SERVER_BIN) $(CMD_DIR)/admin

run-game: build-game ## Run game server locally
	@echo ">> Starting game server..."
	@$(GAME_SERVER_BIN)

run-admin: build-admin ## Run admin server locally
	@echo ">> Starting admin server..."
	@$(ADMIN_SERVER_BIN)

## ===================================================================================
## Quality & Test
## ===================================================================================

test: ## Run all tests with coverage
	@echo ">> Running tests..."
	@$(GO_TEST) -v -race -cover ./...

lint: ## Run linter (requires golangci-lint)
	@echo ">> Running linter..."
	@golangci-lint run ./...

tidy: ## Tidy go modules
	@echo ">> Tidying go modules..."
	@$(GO_MOD_TIDY)

## ===================================================================================
## Database Migration
## ===================================================================================

# !! 請將 DB_URL 替換為你的本地開發資料庫連接字串 !!
DB_URL="postgresql://user:password@localhost:5432/fish_db?sslmode=disable"

migrate-up: ## Apply all up migrations (requires migrate CLI)
	@echo ">> Applying database migrations..."
	@migrate -database "$(DB_URL)" -path ./storage/migrations up

migrate-down: ## Revert last migration (requires migrate CLI)
	@echo ">> Reverting last database migration..."
	@migrate -database "$(DB_URL)" -path ./storage/migrations down 1

## ===================================================================================
## Docker Operations
## ===================================================================================

docker-build: ## Build docker images
	@echo ">> Building docker images..."
	@$(DOCKER_COMPOSE) build

run-dev: ## Run services with docker-compose
	@echo ">> Starting services with docker-compose..."
	@$(DOCKER_COMPOSE) up --build

docker-down: ## Stop and remove docker-compose containers
	@echo ">> Stopping docker-compose services..."
	@$(DOCKER_COMPOSE) down

docker-logs: ## Follow logs from docker-compose services
	@echo ">> Following logs..."
	@$(DOCKER_COMPOSE) logs -f

## ===================================================================================
## Cleanup & Help
## ===================================================================================

clean: ## Clean up build artifacts and caches
	@echo ">> Cleaning up..."
	@rm -rf $(BIN_DIR)
	@$(GO_CLEAN) -cache -testcache -modcache

help: ## Show this help message
	@echo "Usage: make <target>"
	@echo ""
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)