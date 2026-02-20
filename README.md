# Violet Deterministic API

Deterministic API-first service to replace Violet Rails runtime for SaaS building and operations.

## Goals

1. Deterministic decision APIs for humans and AI agents (including OpenClaw-like loops).
2. Safe customization via rule/policy versioning (no runtime `eval`).
3. Self-host and managed-service deployment paths.
4. Explicit replay, idempotency, and auditability.

## API Surface

1. `GET /v1/health`
2. `POST /v1/decisions`
3. `POST /v1/feedback`
4. `POST /v1/replay`
5. `POST /v1/apps`
6. `GET /v1/apps/{id}`
7. `PATCH /v1/apps/{id}`
8. `POST /v1/apps/{id}/mutations`
9. `POST /v1/apps/{id}/verify`
10. `POST /v1/apps/{id}/deploy-intents/self-host`
11. `POST /v1/apps/{id}/deploy-intents/managed`
12. `POST /v1/agents/plan`
13. `POST /v1/agents/clarify`
14. `POST /v1/agents/act`
15. `POST /v1/agents/verify`
16. `POST /v1/agents/deploy`
17. `GET /v1/llm/providers`
18. `POST /v1/llm/infer`
19. `GET /v1/tools`
20. `POST /v1/studio/jobs`
21. `GET /v1/studio/jobs/{id}`
22. `GET /v1/studio/jobs/{id}/events` (SSE)
23. `GET /v1/studio/jobs/{id}/preview` (web/mobile preview HTML)
24. `GET /v1/studio/jobs/{id}/runtime/{client}/{asset}` (generated runtime JS/CSS)
25. `POST /v1/studio/jobs/{id}/terminal`
26. `GET /v1/studio/jobs/{id}/console`
27. `GET /v1/studio/jobs/{id}/artifacts`
28. `POST /v1/studio/jobs/{id}/run`
29. `GET /v1/studio/jobs/{id}/verification`
30. `GET /v1/studio/jobs/{id}/jtbd`
31. `GET /v1/studio/jobs/{id}/bundle` (download generated workspace tarball)

## Determinism Contract

1. Stable sorting: score desc, item_id asc.
2. Immutable `policy_version` and `data_version` included in each decision.
3. Decision hash computed from canonical request + versions.
4. Idempotent decision creation via `Idempotency-Key`.
5. Replay endpoint returns exact stored decision payload by `decision_id`.

## Runtime Topology (demo)

1. `api` (this service)
2. `postgres` (durable state; optional in scaffold)
3. `redis` (idempotency cache; optional in scaffold)
4. `gorse` (`gorse-in-one`)

See `docker-compose.demo.yml`.

## Quick start (scaffold)

```bash
# Build API image
DOCKER_BUILDKIT=1 docker compose -f docker-compose.demo.yml build api

# Start all services
docker compose -f docker-compose.demo.yml up -d

# Probe health
curl -s http://localhost:4020/v1/health | jq

# Open trial UI
open http://localhost:4020/ui/

# Use the seeded demo token
AUTH="Authorization: Bearer dev-token"

# Create a blueprint
curl -s -X POST http://localhost:4020/v1/apps \
  -H "$AUTH" \
  -H "Idempotency-Key: app-create-1" \
  -H "Content-Type: application/json" \
  -d '{"name":"My SaaS","blueprint":{"plan":"starter","region":"us-east-1"}}' | jq

# List model providers (local + frontier)
curl -s http://localhost:4020/v1/llm/providers \
  -H "$AUTH" | jq

# One model call (local-first through Ollama)
curl -s -X POST http://localhost:4020/v1/llm/infer \
  -H "$AUTH" \
  -H "Idempotency-Key: llm-infer-1" \
  -H "Content-Type: application/json" \
  -d '{"provider":"ollama","model":"glm-4.7-flash:latest","prompt":"Draft API tool contracts for tenant-scoped billing."}' | jq

# OpenAI-compatible provider lane (defaults to local Ollama /v1 in demo compose)
curl -s -X POST http://localhost:4020/v1/llm/infer \
  -H "$AUTH" \
  -H "Idempotency-Key: llm-infer-2" \
  -H "Content-Type: application/json" \
  -d '{"provider":"frontier","model":"glm-4.7-flash:latest","prompt":"Return only: ok"}' | jq
```

## Release R1 planning package

1. Handoff doc: `/Users/jaibhagat/code/violet-deterministic-api/docs/handoff/HANDOFF_RELEASE_R1.md`
2. Shape Up docs: `/Users/jaibhagat/code/violet-deterministic-api/docs/shapeup/release-r1/`
3. Board JSON: `/Users/jaibhagat/code/violet-deterministic-api/planning/release-r1/board.json`
4. Tickets JSON: `/Users/jaibhagat/code/violet-deterministic-api/planning/release-r1/tickets.json`
5. Milestones and risks: `/Users/jaibhagat/code/violet-deterministic-api/planning/release-r1/`
6. Fullstack output RFC: `/Users/jaibhagat/code/violet-deterministic-api/docs/rfc/RFC-0001-fullstack-violet-rails-output.md`
7. RFC field report: `/Users/jaibhagat/code/violet-deterministic-api/docs/field-reports/FR-2026-02-19-rfc-0001-ralph-junior.md`

## Seeded skills and orchestration

1. Skills index: `/Users/jaibhagat/code/violet-deterministic-api/skills/README.md`
2. Orchestration sequence: `/Users/jaibhagat/code/violet-deterministic-api/skills/ORCHESTRATION.md`
3. Agent skill manifest: `/Users/jaibhagat/code/violet-deterministic-api/AGENTS.md`
4. Code walkthrough: `/Users/jaibhagat/code/violet-deterministic-api/docs/CODEBASE_WALKTHROUGH.md`
5. End-user harness PRD: `/Users/jaibhagat/code/violet-deterministic-api/docs/prd/PRD_END_USER_WEB_CLIENT_HARNESS.md`

## Repository structure

- `cmd/api/` entrypoint
- `internal/decision/` deterministic engine + canonical hashing + stage trace
- `internal/storage/` durable postgres persistence for replay/idempotency/apps
- `internal/auth/` bearer-token auth and tenant claims
- `internal/adapters/` integration seams for Gorse, GoRules, and pipeline orchestration
- `docs/` DDIA plan, Shape Up package, ADRs, runbooks
- `planning/` release board/tickets/milestones/risk artifacts
- `skills/` reusable execution skills for this repo
- `api/openapi.yaml` contract seed

## Current status

R1 implementation now includes durable replay/idempotency, tenant-scoped auth, app builder control APIs, verify and deployment-intent APIs, and agent orchestration endpoints. Remaining release management work should continue through the Shape Up board artifacts.

## Trial UI

Use `http://localhost:4020/ui/` for an interactive test console that exercises:
1. Prompt capture and structured confirmation alignment.
2. Chat-style clarification loop (`/v1/agents/clarify`) that maps answers back into the confirmation schema.
3. Build job generation (`/v1/studio/jobs`) with workload preview.
4. Generated code artifact browser and file inspection.
5. Clickable live preview panes for web and mobile clients (`/v1/studio/jobs/{id}/preview?client=web|mobile`) backed by generated runtime assets.
6. Live SSE updates (`/v1/studio/jobs/{id}/events`) for status, workload, code, terminal, and console.
7. Pseudo terminal (`/v1/studio/jobs/{id}/terminal`) and console log stream (`/v1/studio/jobs/{id}/console`).
8. Real-user flow: prompt -> clarify -> confirm form -> generation with inspectable artifacts.
9. Multi-model router panel for local Ollama and frontier providers (`/v1/llm/providers`, `/v1/llm/infer`).
10. `Run Model Call` can auto-trigger `studio_generate` post-hook to create a build job and immediately hydrate workload, code artifacts, and web/mobile previews in the same UI session.
11. Generated jobs are materialized to an actual workspace path (`output/studio/<job_id>` in runtime), and terminal supports `exec <command>` for real shell execution in that workspace.
12. Release gate panel provides run targets (`web`, `mobile`, `api`, `verify`, `all`) and renders artifacts/verification/JTBD evidence live.
13. Download bundle link exports generated workspace as `.tar.gz` including files + manifest for local execution/handoff.
14. `Run API` executes generated backend smoke checks (`go test`, boot service, `/health`, `/v1/tools`, workflow execute) in the studio runner environment.
15. End-user harness UI is available at `http://localhost:4020/ui/harness.html` for a prompt-first product flow.

## CLI Interface

Build and use the local CLI to call tool endpoints and model providers:

```bash
# Build with local Go toolchain
go build -o bin/vda ./cmd/vda

# Or build inside container (works even if Go is not installed locally)
make cli

# Tools catalog
bin/vda tools list --token dev-token

# List providers (checks local Ollama + configured frontier)
bin/vda llm providers --token dev-token

# Run one inference call
bin/vda llm infer --token dev-token --provider ollama --model glm-4.7-flash:latest --prompt "Generate tenant-safe mutation plan"

# One-click local launch for a generated Studio job
bin/vda studio launch --token dev-token --job-id <job_id>
```

## Multi-Model Agent Runtime Notes

See `/Users/jaibhagat/code/violet-deterministic-api/docs/AGENT_RUNTIME_MULTIMODEL.md` for the LangGraph + local Ollama + frontier architecture and rollout pattern.

Demo default: `frontier` points to local OpenAI-compatible endpoint (`http://host.docker.internal:11434/v1`) so both `ollama` and `frontier` lanes are testable locally without extra credentials. Set `FRONTIER_BASE_URL` and `FRONTIER_API_KEY` for hosted SOTA providers.
