.PHONY: help build start test tools setup dev migrate-up migrate-down migration-create lint format

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
ifeq (, $(shell which reflex))
	@echo "reflex not found, installing tools..."
	$(MAKE) tools
endif
	reflex -r '^cmd/(.*?).*|.env|pkg/(.*?).*|internal/(.*?).*$$' -s -- go run cmd/server/main.go

migrate-up:
ifeq (, $(shell which migrate))
	@echo "migrate not found, installing tools..."
	$(MAKE) tools
endif
	@. .env && migrate -path $$MIGRATIONS_PATH -database $$DATABASE_URL up

migrate-down:
ifeq (, $(shell which migrate))
	@echo "migrate not found, installing tools..."
	$(MAKE) tools
endif
	@. .env && migrate -path $$MIGRATIONS_PATH -database $$DATABASE_URL down -all

migration-create:
ifeq (, $(shell which migrate))
	@echo "migrate not found, installing tools..."
	$(MAKE) tools
endif
	migrate create -ext sql -dir pkg/database/sql/migrations -seq $(NAME)

lint:
ifeq (, $(shell which golangci-lint))
	@echo "golangci-lint not found, installing tools..."
	$(MAKE) tools
endif
	golangci-lint run --fix

format:
ifeq (, $(shell which golangci-lint))
	@echo "golangci-lint not found, installing tools..."
	$(MAKE) tools
endif
	golangci-lint fmt
