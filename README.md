# DevHelp Backend

A stateless developer toolbox API built with Go + Gin.

## Tools

| Tool | Endpoint |
|---|---|
| HTTP Request Inspector | `POST /api/v1/http-inspector` |
| cURL Converter | `POST /api/v1/curl-converter` |
| JWT Toolkit — Generate | `POST /api/v1/jwt/generate` |
| JWT Toolkit — Decode | `POST /api/v1/jwt/decode` |
| JSON Generator | `POST /api/v1/json-generator` |
| Health | `GET /api/v1/health` |

## Quick Start

```bash
# Clone and enter the backend directory
cd backend

# Copy environment config
cp .env.example .env

# Run in development mode
APP_ENV=development go run ./cmd/main.go

# Or use live-reload (requires Air: go install github.com/air-verse/air@latest)
air
```

## Docker

```bash
# Build and start production container
docker-compose up -d devhelp

# View logs
docker-compose logs -f devhelp
```

## Make Targets

```
make run          # Run with go run (development)
make dev          # Run with Air (live-reload)
make build        # Compile production binary to bin/devhelp
make test         # Run tests with race detector
make lint         # Run golangci-lint
make swagger      # Generate Swagger docs
make docker-build # Build Docker image
make docker-up    # Start via Docker Compose
make docker-down  # Stop containers
make clean        # Remove build artifacts
```

## Architecture

```
backend/
├── cmd/              # main.go — wires all dependencies
├── configs/          # YAML config per environment
├── docs/             # Swagger-generated docs (gitignored until generated)
├── internal/
│   ├── config/       # Viper-based config loader
│   ├── handlers/     # Gin HTTP handlers (thin — delegate to services)
│   ├── logger/       # Zap + lumberjack setup
│   ├── middleware/   # Recovery, RequestID, Logger, CORS, RateLimiter, Timeout
│   ├── routes/       # Route registration
│   ├── services/     # Business logic per tool
│   │   ├── dto/          # Shared request/response types
│   │   ├── httpinspector/
│   │   ├── curlconverter/
│   │   ├── jwttool/
│   │   └── jsongenerator/
│   └── utils/        # Shared helpers (response writers, curl parser)
├── tests/            # Integration + unit tests
└── logs/             # Daily log files (development only)
```

**Adding a new tool:**
1. Create `internal/services/yourtool/service.go` implementing a `Service` interface
2. Create `internal/services/dto/yourtool.go` with request/response DTOs
3. Create `internal/handlers/yourtool.go`
4. Register the route group in `internal/routes/routes.go`

No other files need modification.

## Rate Limiting

100 requests/minute per IP. Returns `HTTP 429` with a `Retry-After: 60` header when exceeded.

## Configuration

| Variable | Default | Description |
|---|---|---|
| `APP_ENV` | `development` | Loads `configs/<APP_ENV>.yaml` |

## Generating Swagger Docs

```bash
go install github.com/swaggo/swag/cmd/swag@latest
make swagger
# Docs served at http://localhost:8080/swagger/index.html
```
