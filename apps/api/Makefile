.PHONY: build run test lint migrate-up migrate-down migrate-create tidy sqlc-generate

# Binary output path
BIN := bin/api

# Go command
GOCMD := go

## build: compile the API binary
build:
	$(GOCMD) build -o $(BIN) ./cmd/server/

## run: build and start the server
run: build
	./$(BIN)

## test: run all tests
test:
	$(GOCMD) test ./...

## lint: run golangci-lint (requires golangci-lint installed)
lint:
	golangci-lint run ./...

## tidy: synchronise go.mod and go.sum
tidy:
	$(GOCMD) mod tidy

MIGRATE_DIR=migrations
DATABASE_URL?=$(shell grep DATABASE_URL .env 2>/dev/null | cut -d= -f2)

## migrate-up: apply all pending migrations
## Usage: make migrate-up DATABASE_URL=postgres://...
migrate-up:
	migrate -path $(MIGRATE_DIR) -database "$(DATABASE_URL)" up

## migrate-down: roll back the last migration
## Usage: make migrate-down DATABASE_URL=postgres://...
migrate-down:
	migrate -path $(MIGRATE_DIR) -database "$(DATABASE_URL)" down 1

## migrate-create: create a new migration
## Usage: make migrate-create name=create_users
migrate-create:
	migrate create -ext sql -dir $(MIGRATE_DIR) -seq $(name)

## sqlc-generate: generate Go code from SQL queries
sqlc-generate:
	sqlc generate
