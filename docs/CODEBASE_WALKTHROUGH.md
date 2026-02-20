# Violet Deterministic API: Code Walkthrough

This walkthrough is optimized for execution, not theory.
Use it to answer:

1. Where do I implement a change?
2. Where do I debug when behavior drifts?
3. Which files define the source of truth for each capability?

## 1. High-Level Mental Model

The service has three major lanes:

1. Deterministic decision/app control plane (`/v1/decisions`, `/v1/apps`, `/v1/agents/*`)
2. Model routing lane (`/v1/llm/*`)
3. Studio generation lane (`/v1/studio/*`, UI, previews, bundle, run targets)

Everything is tenant-scoped via bearer token claims and idempotent mutations.

## 2. Boot and Wiring (Start Here)

Read these first in order:

1. `cmd/api/main.go`
2. `internal/config/config.go`
3. `internal/http/server.go`

What happens:

1. Config is loaded from env.
2. `Server` is constructed with:
   - Postgres-backed `storage.Store`
   - deterministic `decision.Engine`
   - auth parser
   - studio service with durable persistence (`studio.WithPersistence(store)`)
   - LLM service (Ollama + frontier)
3. All routes are registered in one place (`internal/http/server.go`).

If an endpoint is missing or wrong, verify it is registered there first.

## 3. Primary Code Domains

### 3.1 HTTP Layer (Request Parsing + Response Contracts)

Files:

1. `internal/http/handlers.go` (health, decisions, apps, verify, deploy-intent, agents)
2. `internal/http/llm_handlers.go` (provider listing + infer + studio hook)
3. `internal/http/studio_handlers.go` (Studio jobs, run, verification, preview, bundle, terminal)
4. `internal/http/tools_handlers.go` (tool catalog/CLI hints)

Rule of thumb:

1. If a bug is about status codes, auth errors, idempotency headers, request shape, or endpoint behavior, start in HTTP handlers.

### 3.2 Deterministic Core

Files:

1. `internal/decision/engine.go`
2. `internal/decision/types.go`

Responsibilities:

1. Candidate ranking with stable tie-breaking.
2. Canonical hashing across request/context/candidates/stages.
3. Policy and recommendation adapter integration.

If replay/hash behavior is surprising, debug `hashDecision`, `normalizeContext`, and `normalizeCandidates`.

### 3.3 Durable Storage

File:

1. `internal/storage/postgres.go`

Responsibilities:

1. Schema initialization.
2. Idempotency read/write/cleanup.
3. Decision payload persistence for replay.
4. App/mutation/verify/deploy persistence.
5. Studio job persistence (`studio_jobs`).

If anything disappears across restart, verify corresponding `Save*` / `Get*` methods exist and are called.

### 3.4 Auth and Tenant Scope

File:

1. `internal/auth/auth.go`

Mechanics:

1. Token map comes from `AUTH_TOKENS` env.
2. Format: `token:tenant_id:subject` (subject optional).
3. All critical handlers check claims and enforce tenant match.

If one tenant can read another tenant's object, start here plus the specific handler's tenant checks.

### 3.5 LLM Provider Routing

File:

1. `internal/llm/service.go`

Responsibilities:

1. Provider discovery and health.
2. Ollama inference (`/api/generate`).
3. Frontier/OpenAI-compatible inference (`/chat/completions`).
4. Error normalization (`frontier_auth_required`, `ollama_unreachable`, etc).

If model calls fail, confirm base URLs, API keys, and timeouts in config.

### 3.6 Studio Generation System

Core files:

1. `internal/studio/service.go` (job lifecycle, materialization, terminal, run targets)
2. `internal/studio/rfc_contracts.go` (artifact manifest, verification report, JTBD coverage)
3. `internal/studio/preview.go` (web/mobile preview rendering + runtime assets)
4. `internal/studio/backend_artifacts.go` (generated backend service scaffold)
5. `internal/studio/runtime_exec.go` (real API runtime smoke checks)
6. `internal/studio/bundle.go` (bundle export tar.gz)

If Studio output is "present but not useful", focus here.

### 3.7 UI and CLI Surfaces

UI files:

1. `internal/http/ui/index.html`
2. `internal/http/ui/app.js`
3. `internal/http/ui/styles.css`
4. `internal/http/ui_embed.go`

CLI files:

1. `cmd/vda/main.go`
2. `cmd/vda/studio.go`

Use UI for interactive operator loop. Use CLI for automation and repeatable launch/testing flows.

### 3.8 Contract Source of Truth

File:

1. `api/openapi.yaml`

If handler behavior changed, update OpenAPI in same PR.

## 4. Task-Oriented "Where to Implement"

### Add a new API endpoint

1. Register route in `internal/http/server.go`.
2. Implement handler in the right `internal/http/*_handlers.go` file.
3. Add persistence in `internal/storage/postgres.go` if needed.
4. Update `api/openapi.yaml`.
5. Add/extend tests.

### Add a new app mutation class

1. Add class behavior in `applyMutation` (`internal/http/handlers.go`).
2. Ensure policy allows/blocks appropriately (`internal/adapters/gorules/local_client.go`).
3. Verify mutation snapshot persistence (`SaveMutation`).
4. Add tests around mutation and verify paths.

### Change Studio generated artifacts

1. Change artifact composition in `buildFiles` (`internal/studio/service.go`).
2. Extend backend outputs in `internal/studio/backend_artifacts.go`.
3. Update verification/JTBD checks in `internal/studio/rfc_contracts.go`.
4. Validate preview/runtime if web/mobile assets changed (`internal/studio/preview.go`).

### Change run-target behavior

1. Static artifact checks: `runTargetChecks` (`internal/studio/rfc_contracts.go`).
2. Executable checks: `runGeneratedAPIRuntimeChecks` (`internal/studio/runtime_exec.go`).
3. Run orchestration entrypoint: `RunTarget` (`internal/studio/service.go`).

### Change LLM integration or provider behavior

1. Provider registry/infer logic: `internal/llm/service.go`.
2. HTTP mapping + hook behavior: `internal/http/llm_handlers.go`.
3. UI model controls: `internal/http/ui/app.js`.

### Change one-command local launch

1. Main command parser: `cmd/vda/main.go`.
2. Launch workflow implementation: `cmd/vda/studio.go`.
3. Tool catalog docs: `internal/http/tools_handlers.go`.

## 5. Fast Debug Playbook (Symptom -> First Files)

### API returns `missing_idempotency_key`

1. Caller did not send header on mutating route.
2. Check `idempotencyKey` in `internal/http/server.go`.
3. Confirm client uses `Idempotency-Key` (UI/CLI or external caller).

### Replay fails or deterministic hash seems wrong

1. `internal/decision/engine.go` hash and normalization functions.
2. Verify request candidate order/tags and context keys are canonicalized.
3. Check persisted payload in `internal/storage/postgres.go` `SaveDecision/GetDecisionPayload`.

### App verify fails preflight

1. `executeVerify` in `internal/http/handlers.go`.
2. Confirm app blueprint has `plan` and `region`.
3. Trace mutation history (`SaveMutation`) if these fields were expected to be set.

### LLM provider unreachable or auth errors

1. `internal/llm/service.go` (`list*Models`, `infer*`, `frontierHeaders`).
2. Env in `internal/config/config.go`:
   - `OLLAMA_BASE_URL`
   - `FRONTIER_BASE_URL`
   - `FRONTIER_API_KEY`
3. `GET /v1/llm/providers` output for reachability and error details.

### Studio job exists but artifacts missing after restart

1. Ensure `SaveStudioJob/GetStudioJob` are called (`internal/studio/service.go` + storage methods).
2. Confirm `studio_jobs` table exists (`internal/storage/postgres.go` schema init).
3. Validate workspace rematerialization path (`ensureWorkspace` in `internal/studio/service.go`).

### `Run API` fails in release gate

1. `runGeneratedAPIRuntimeChecks` in `internal/studio/runtime_exec.go`.
2. Common causes:
   - generated `services/api` folder missing
   - go toolchain unavailable in runtime container
   - generated server fails to boot
3. Check evidence returned in run checks (`api_runtime_*`).

### Preview loads but looks stale

1. Preview cache-busting query in UI (`internal/http/ui/app.js` `refreshPreview`).
2. Runtime asset serving (`handleStudioRuntimeAsset` in `internal/http/studio_handlers.go`).
3. Generated runtime asset lookup (`lookupGeneratedRuntimeAsset` in `internal/studio/preview.go`).

### Terminal `exec` command rejected unexpectedly

1. Allowlist and token blocking: `parseExecCommand` (`internal/studio/service.go`).
2. Remember current protections:
   - no shell metacharacters
   - no absolute paths
   - no parent traversal

### Bundle download or launch command fails

1. Bundle builder: `internal/studio/bundle.go`.
2. Bundle handler: `handleStudioBundle` in `internal/http/studio_handlers.go`.
3. CLI extraction/launch: `cmd/vda/studio.go`.

## 6. Data and Generated Output Locations

Runtime-generated outputs:

1. Studio workspace root (default): `output/studio/<job_id>`
2. Launch extraction root: `output/launch/<job+timestamp>/...`

Historical sample artifacts in repo (for reference only):

1. `internal/studio/output/studio/...`

Treat `output/` as operational artifacts, not source of truth.

## 7. Test and Validation Landmarks

Primary tests:

1. `internal/decision/engine_test.go`
2. `internal/auth/auth_test.go`
3. `internal/llm/service_test.go`
4. `internal/studio/preview_test.go`
5. `cmd/vda/studio_test.go`

Suggested local validation sequence:

1. `docker compose -f docker-compose.demo.yml up -d --build api`
2. `curl -s http://localhost:4020/v1/health | jq`
3. Create Studio job, then run:
   - `POST /v1/studio/jobs/{id}/run` target `all`
4. Verify UI at `http://localhost:4020/ui/`
5. Verify CLI launch:
   - `bin/vda studio launch --job-id <id>`

## 8. Common Change Strategy

When you add a capability, ship in this order:

1. Contract: OpenAPI + request/response shapes.
2. Handler: endpoint behavior and idempotency.
3. Domain/service logic.
4. Persistence.
5. Verification/JTBD checks.
6. UI/CLI surfaces.
7. Tests + field report note.

This prevents "UI looks done but backend contract is missing" regressions.

## 9. Key Docs to Keep Open While Working

1. `docs/rfc/RFC-0001-fullstack-violet-rails-output.md`
2. `docs/field-reports/FR-2026-02-19-rfc-0001-ralph-junior.md`
3. `docs/handoff/HANDOFF_RELEASE_R1.md`
4. `docs/shapeup/release-r1/README.md`

These define release intent and current gap history.
