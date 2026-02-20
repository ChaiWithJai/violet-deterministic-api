# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Is

Violet Deterministic API (VDA) — a deterministic, API-first Go service replacing Violet Rails runtime for SaaS building and operations. Core guarantees: deterministic decision ranking, immutable policy/data versioning, idempotent mutations, tenant-scoped auth, and replay-safe contracts.

## Build & Run Commands

```bash
# Docker-based (primary workflow)
make up        # Start full stack: postgres, redis, gorse, api (port 4020)
make down      # Tear down
make build     # Rebuild API container
make logs      # Follow API logs
make cli       # Cross-compile CLI binary to bin/vda

# Local Go
go build -o bin/vda ./cmd/api       # API server
go build -o bin/vda ./cmd/vda       # CLI tool
go test ./internal/...              # Run all internal tests
go test ./internal/decision/...     # Run specific package tests
go test -run TestDecideDeterministic ./internal/decision/  # Single test
go fmt ./...                        # Format
go vet ./...                        # Static analysis
```

## Stack Topology (docker-compose.demo.yml)

| Service    | Port  | Purpose                        |
|------------|-------|--------------------------------|
| PostgreSQL | 5436  | Primary storage                |
| Redis      | 6382  | Gorse cache backend            |
| Gorse      | 19086/19088 | Recommendation engine    |
| API        | 4020  | VDA service                    |

Dev auth tokens: `dev-token:t_acme:dev-user` and `ops-token:t_ops:ops-user`

## Architecture

### Three-Lane Design

1. **Decision/App Control Plane** — Deterministic ranking, app CRUD, blueprint mutations, verification, deploy intents
2. **Model Routing Lane** — Multi-provider LLM (local Ollama + OpenAI-compatible "frontier"), request-level provider selection
3. **Studio Generation Lane** — Prompt-to-app code generation, live SSE events, preview rendering, bundle export, terminal access

### Package Map

- `cmd/api/` — HTTP server entrypoint
- `cmd/vda/` — CLI tool
- `internal/config/` — Env var loading with defaults
- `internal/http/` — All HTTP handlers and route registration (`server.go` wires routes)
- `internal/decision/` — Core deterministic engine (three stages: gorse → policy → rank)
- `internal/storage/` — PostgreSQL persistence (pgx v4), auto-creates tables on init
- `internal/auth/` — Bearer token authentication from `AUTH_TOKENS` env
- `internal/idempotency/` — Request deduplication via `Idempotency-Key` header
- `internal/llm/` — Ollama and frontier provider HTTP clients
- `internal/studio/` — Job lifecycle, preview generation, bundle export, runtime exec
- `internal/adapters/gorse/` — Recommendation client interface
- `internal/adapters/gorules/` — Policy evaluation client interface
- `internal/adapters/pipeline/` — Orchestration seam (reserved)

### Determinism Contract

- **Stable sort**: score descending, item_id ascending (tie-breaker)
- **Canonical hashing**: context keys, candidates, and tags sorted before SHA256
- **Immutable versions**: `policy_version` and `data_version` frozen per decision
- **Idempotency**: all mutations wrapped in `withIdempotency()` — same key returns cached response
- **Graceful degradation**: gorse/policy failures produce `dependency_status: "degraded"`, decision still completes

### Handler Pattern

Every handler follows: auth claims → idempotency key → parse body → validate → wrap in idempotency (for mutations) → business logic → JSON response. See `internal/http/handlers.go`.

### Adding New Endpoints

1. Add handler method to Server in the appropriate `*_handlers.go` file
2. Register route in `internal/http/server.go`
3. Add storage method in `internal/storage/postgres.go` if persistence needed
4. Update `api/openapi.yaml`

## Key Dependencies

- **Go 1.22** (specified in go.mod and Dockerfile)
- **pgx/v4** — only direct dependency (PostgreSQL driver)
- **No web framework** — stdlib `net/http` with manual routing

## API Contract

OpenAPI spec at `api/openapi.yaml`. Key route groups:
- `POST /v1/decisions`, `POST /v1/replay`, `POST /v1/feedback`
- `POST /v1/apps`, `GET/PATCH /v1/apps/{id}`, mutations, verify, deploy-intents
- `POST /v1/agents/{plan|clarify|act|verify|deploy}`
- `GET /v1/llm/providers`, `POST /v1/llm/infer`
- `POST /v1/studio/jobs`, SSE events, preview, bundle, terminal
- `GET /v1/health`

## Skills

See `AGENTS.md` for available execution skills. Key ones: `deterministic-go-api-implementation`, `agent-orchestration`, `ddia-api-first-audit`, `shapeup-release-r1`.
