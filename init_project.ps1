# init_project.ps1 - A script to initialize the Neptune project structure for Windows.
# Usage: ./init_project.ps1 <your_go_module_name>

param (
    [string]$ModuleName
)

# Set error action to stop on first error
$ErrorActionPreference = "Stop"

if ([string]::IsNullOrEmpty($ModuleName)) {
    Write-Host "Error: Go module name is required." -ForegroundColor Red
    Write-Host "Usage: ./init_project.ps1 <your_go_module_name>"
    exit 1
}

Write-Host "Initializing Go module: $ModuleName"
# Check if go.mod exists, if not, create it
if (-not (Test-Path -Path "go.mod")) {
    go mod init $ModuleName
} else {
    Write-Host "go.mod already exists, skipping initialization." -ForegroundColor Yellow
}


Write-Host "Creating directory structure..."
# Define all directories to be created
$dirs = @(
    "api/proto/v1",
    "cmd/game", "cmd/admin", "cmd/migrator",
    "configs",
    "deployments",
    "internal/app/admin", "internal/app/game",
    "internal/biz/game", "internal/biz/player", "internal/biz/wallet",
    "internal/data/postgres", "internal/data/redis",
    "internal/conf",
    "internal/pkg/logger", "internal/pkg/token",
    "pkg/pb/v1",
    "scripts",
    "storage/migrations"
)

# Loop and create directories
foreach ($dir in $dirs) {
    if (-not (Test-Path -Path $dir)) {
        New-Item -ItemType Directory -Path $dir | Out-Null
    }
}

Write-Host "Creating placeholder files..."
# Define all files to be created
$files = @(
    "api/proto/v1/common.proto", "api/proto/v1/game.proto",
    "configs/config.yaml",
    "deployments/Dockerfile.game", "deployments/Dockerfile.admin", "deployments/docker-compose.yml",
    "cmd/game/main.go", "cmd/game/wire.go",
    "cmd/admin/main.go", "cmd/admin/wire.go",
    "cmd/migrator/main.go",
    "internal/app/wire.go", "internal/app/game/server.go",
    "internal/biz/wire.go", "internal/biz/game/game.go",
    "internal/data/wire.go", "internal/data/data.go",
    "internal/conf/conf.go", "internal/conf/wire.go",
    "internal/pkg/logger/logger.go", "internal/pkg/token/token.go",
    "storage/migrations/000001_create_initial_tables.up.sql", "storage/migrations/000001_create_initial_tables.down.sql",
    "scripts/proto-gen.sh", "scripts/wire-gen.sh", "scripts/migrate.sh", "scripts/run-dev.sh"
)

# Loop and create empty files
foreach ($file in $files) {
    if (-not (Test-Path -Path $file)) {
        New-Item -ItemType File -Path $file | Out-Null
    }
}


# Create READMEs and Makefile with content
Set-Content -Path "configs/README.md" -Value "# Configuration guide for developers"

# Use a Here-String for the Makefile content
$makefileContent = @"
.PHONY: all proto wire build run-dev test clean help

# Go variables
# Using 'go' directly should work if it's in the PATH
GO            ?= go
GO_BUILD      := \$(GO) build
GO_RUN        := \$(GO) run
GO_TEST       := \$(GO) test
GO_CLEAN      := \$(GO) clean

# Directories
CMD_DIR := ./cmd

# Binaries
GAME_SERVER_BIN := game-server.exe
ADMIN_SERVER_BIN := admin-server.exe

all: build

# Generate code from protobuf
proto:
	@echo ">> generating protobuf code"
	@bash ./scripts/proto-gen.sh

# Generate dependency injection code
wire:
	@echo ">> generating wire code"
	@bash ./scripts/wire-gen.sh

# Build binaries
build: proto wire
	@echo ">> building binaries"
	@\$(GO_BUILD) -o \$(GAME_SERVER_BIN) \$(CMD_DIR)/game
	@\$(GO_BUILD) -o \$(ADMIN_SERVER_BIN) \$(CMD_DIR)/admin

# Run services locally
run-dev:
	@echo ">> starting services with docker-compose"
	@docker-compose -f deployments/docker-compose.yml up --build

# Run tests
test:
	@echo ">> running tests"
	@\$(GO_TEST) -v ./...

# Clean build artifacts
clean:
	@echo ">> cleaning up"
	@if exist \$(GAME_SERVER_BIN) (del \$(GAME_SERVER_BIN))
	@if exist \$(ADMIN_SERVER_BIN) (del \$(ADMIN_SERVER_BIN))
	@\$(GO_CLEAN)

# Show help
help:
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  proto      Generate protobuf code"
	@echo "  wire       Generate wire dependency injection code"
	@echo "  build      Build all binaries"
	@echo "  run-dev    Run all services for local development using Docker Compose"
	@echo "  test       Run all tests"
	@echo "  clean      Clean up build artifacts"
	@echo ""
"@

Set-Content -Path "Makefile" -Value $makefileContent

# Note: The .sh scripts in the scripts/ directory will require a Bash-compatible shell on Windows,
# like Git Bash, which is highly recommended for Go development on Windows.

Write-Host "Project structure initialized successfully." -ForegroundColor Green
Write-Host "Next steps:"
Write-Host "1. Run 'go mod tidy' and 'go get' to install initial dependencies."
Write-Host "2. Fill in the placeholder files with your logic."
Write-Host "3. Customize the scripts in the ./scripts directory."
Write-Host "4. For 'make' command on Windows, you might need to install 'make' via Chocolatey or Scoop."
Write-Host "5. Use 'make help' to see available commands."