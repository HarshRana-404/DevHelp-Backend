.PHONY: run dev build test lint clean docker-build docker-up docker-down swagger tidy

## ── Local Development ────────────────────────────────────────────────────────

# Run the server in development mode (standard go run).
run:
	APP_ENV=development go run ./cmd/main.go

# Run with live-reload using Air.
dev:
	air

# Build a production binary into bin/devhelp.
build:
	CGO_ENABLED=0 go build -ldflags="-w -s" -o bin/devhelp ./cmd/main.go

## ── Testing ──────────────────────────────────────────────────────────────────

# Run all unit tests with race detector and coverage.
test:
	go test -race -cover ./tests/... ./internal/...

# Run tests with verbose output.
test-v:
	go test -race -v -cover ./tests/... ./internal/...

## ── Code Quality ─────────────────────────────────────────────────────────────

# Run golangci-lint (must be installed: brew install golangci-lint).
lint:
	golangci-lint run ./...

# Tidy Go module dependencies.
tidy:
	go mod tidy

## ── Swagger ──────────────────────────────────────────────────────────────────

# Generate Swagger docs (requires swag: go install github.com/swaggo/swag/cmd/swag@latest).
swagger:
	swag init -g cmd/main.go -o docs

## ── Docker ───────────────────────────────────────────────────────────────────

# Build the Docker image.
docker-build:
	docker build -t devhelp:latest .

# Start the production container via Docker Compose.
docker-up:
	docker-compose up -d devhelp

# Stop and remove containers.
docker-down:
	docker-compose down

# View live logs from the running container.
docker-logs:
	docker-compose logs -f devhelp

## ── Cleanup ──────────────────────────────────────────────────────────────────

clean:
	rm -rf bin/ tmp/ logs/*.log
