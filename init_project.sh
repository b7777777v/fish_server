#!/bin/bash

# init_project.sh - A script to initialize the Neptune project structure for Linux/macOS.
# Usage: ./init_project.sh <your_go_module_name>

# Exit immediately if a command exits with a non-zero status.
set -e

# Check for module name argument
MODULE_NAME=$1
if [ -z "$MODULE_NAME" ]; then
    echo "Error: Go module name is required."
    echo "Usage: $0 <your_go_module_name>"
    exit 1
fi

echo "Initializing Go module: $MODULE_NAME"
# Check if go.mod exists, if not, create it
if [ ! -f go.mod ]; then
    go mod init "$MODULE_NAME"
else
    echo "go.mod already exists, skipping initialization."
fi


echo "Creating directory structure..."
# Create main directories using brace expansion
mkdir -p api/proto/v1 \
         cmd/{game,admin,migrator} \
         configs \
         deployments \
         internal/app/{admin,game} \
         internal/biz/{game,player,wallet} \
         internal/data/{postgres,redis} \
         internal/conf \
         internal/pkg/{logger,token} \
         pkg/pb/v1 \
         scripts \
         storage/migrations

echo "Creating placeholder files..."
# Use `touch` to create multiple files at once
touch api/proto/v1/{common.proto,game.proto} \
      configs/config.yaml \
      deployments/{Dockerfile.game,Dockerfile.admin,docker-compose.yml} \
      cmd/game/{main.go,wire.go} \
      cmd/admin/{main.go,wire.go} \
      cmd/migrator/main.go \
      internal/app/{wire.go,game/server.go} \
      internal/biz/{wire.go,game/game.go} \
      internal/data/{wire.go,data.go} \
      internal/conf/{conf.go,wire.go} \
      internal/pkg/logger/logger.go \
      internal/pkg/token/token.go \
      storage/migrations/{000001_create_initial_tables.up.sql,000001_create_initial_tables.down.sql} \
      scripts/{proto-gen.sh,wire-gen.sh,migrate.sh,run-dev.sh}

# Create READMEs and Makefile with content
echo "# Configuration guide for developers" > configs/README.md

cat > Makefile << EOL
.PHONY: all proto wire build run-dev test clean help

# Go variables
GO            ?= go
GO_BUILD      := \$(GO) build
GO_RUN        := \$(GO) run
GO_TEST       := \$(GO) test
GO_CLEAN      := \$(GO) clean

# Directories
CMD_DIR := ./cmd

# Binaries
GAME_SERVER_BIN := game-server
ADMIN_SERVER_BIN := admin-server

all: build

# Generate code from protobuf
proto:
	@echo ">> generating protobuf code"
	@sh ./scripts/proto-gen.sh

# Generate dependency injection code
wire:
	@echo ">> generating wire code"
	@sh ./scripts/wire-gen.sh

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
	@rm -f \$(GAME_SERVER_BIN) \$(ADMIN_SERVER_BIN)
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
EOL

# Make scripts executable
chmod +x scripts/*.sh

echo "Project structure initialized successfully."
echo "Next steps:"
echo "1. Run 'go mod tidy' and 'go get' to install initial dependencies."
echo "2. Fill in the placeholder files with your logic."
echo "3. Customize the scripts in the ./scripts directory."
echo "4. Use 'make help' to see available commands."