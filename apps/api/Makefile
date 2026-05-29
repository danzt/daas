.PHONY: build run test lint migrate-up migrate-down tidy

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

## migrate-up: apply all pending migrations
## Usage: make migrate-up DATABASE_URL=postgres://...
migrate-up:
	migrate -path ./migrations -database "$(DATABASE_URL)" up

## migrate-down: roll back the last migration
## Usage: make migrate-down DATABASE_URL=postgres://...
migrate-down:
	migrate -path ./migrations -database "$(DATABASE_URL)" down 1
