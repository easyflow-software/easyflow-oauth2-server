.PHONY: help build start test tools setup dev migrate-up migrate-down migration-create lint format generate

# Function to check if a command exists and install tools if not
# Usage: $(call check-command,command-name)
define check-command
	@if ! command -v $(1) >/dev/null 2>&1; then \
		echo "$(1) not found, installing tools..."; \
		$(MAKE) tools; \
	fi
endef

help:
	@echo "Makefile commands:"
	@echo "  build             Build the server binary"
	@echo "  start             Build the binary and start the server"
	@echo "  test              Run tests"
	@echo "  tools             Install development tools"
	@echo "  setup             Setup the project (install tools and download dependencies)"
	@echo "  dev               Start the server in development mode with live reloading"
	@echo "  migrate-up        Apply database migrations"
	@echo "  migrate-down      Rollback all database migrations"
	@echo "  migration-create  Create a new database migration (usage: make migration-create NAME=<migration_name>)"
	@echo "  lint              Run linter on the codebase"
	@echo "  format            Format the codebase"
	@echo "  generate          Generate code (sqlc, mockery, swagger)"

build:
	go build -o bin/server cmd/server/main.go
	chmod +x bin/server

start: build
	./bin/server

test:
	go test ./...

tools:
	./scripts/install.sh

setup: tools
	go mod download

dev:
	$(call check-command,reflex)
	reflex -r '^cmd/(.*?).*|.env|pkg/(.*?).*|internal/(.*?).*$$' -s -- go run cmd/server/main.go

migrate-up:
	$(call check-command,migrate)
	@. .env && migrate -path $$MIGRATIONS_PATH -database $$DATABASE_URL up

migrate-down:
	$(call check-command,migrate)
	@. .env && migrate -path $$MIGRATIONS_PATH -database $$DATABASE_URL down -all

migration-create:
	$(call check-command,migrate)
	migrate create -ext sql -dir pkg/database/sql/migrations -seq $(NAME)

lint:
	$(call check-command,golangci-lint)
	golangci-lint run --fix

format:
	$(call check-command,golangci-lint)
	golangci-lint fmt

generate:
	$(call check-command,sqlc)
	$(call check-command,mockery)
	$(call check-command,swag)
	sqlc generate
	mockery
	swag init -g cmd/server/main.go -o internal/server/docs --parseDependency --parseInternal
