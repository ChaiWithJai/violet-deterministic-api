# Local Smoke Runbook

1. `docker compose -f docker-compose.demo.yml build api`
2. `docker compose -f docker-compose.demo.yml up -d`
3. `curl -s http://localhost:4020/v1/health`
4. Post sample decision to `/v1/decisions` with `Idempotency-Key`.
5. Replay via `/v1/replay` using returned `decision_id`.

## Verification evidence (2026-02-18 UTC)

### Machine context
1. OS: macOS 26.0 (Darwin 25.0.0 arm64)
2. Timestamp: 2026-02-18T18:01:34Z

### Executed checks and outcomes
1. `docker compose -f docker-compose.demo.yml config` -> PASS (`compose_config_ok`, 98 lines rendered).
2. `docker compose -f docker-compose.demo.yml up -d` -> FAIL (`Cannot connect to the Docker daemon at unix:///Users/jaibhagat/.docker/run/docker.sock`).
3. `go test ./...` -> FAIL (`go: command not found`).
4. Live API checks (`/v1/health`, `/v1/decisions`, `/v1/replay`) -> BLOCKED because service could not be started locally.

### Implementation-scope verification (static)
1. Implemented HTTP routes are currently: `/v1/health`, `/v1/decisions`, `/v1/replay`, `/v1/feedback`.
2. `api/openapi.yaml` currently defines the same 4 routes only.
3. Planned R1 surfaces for app build/customize/verify/deploy (`/v1/apps`, `/v1/apps/{id}/verify`, deployment-intent endpoints) are not yet present in runtime routes or OpenAPI.

## Verification evidence (2026-02-18 UTC, container runtime pass)

### Runtime outcomes
1. `docker compose -f docker-compose.demo.yml build api` -> PASS.
2. `docker compose -f docker-compose.demo.yml up -d` -> PASS.
3. `GET /v1/health` -> PASS (`200` and expected `status`, `policy_version`, `data_version`).
4. `POST /v1/decisions` twice with same `Idempotency-Key` -> PASS (identical payload returned).
5. `POST /v1/replay` using returned `decision_id` -> PASS (exact payload replayed).
6. Tie-break behavior -> PASS (`score desc`, then `item_id asc` on equal scores).

### Scope and contract gaps observed in live checks
1. `POST /v1/apps` -> `404`.
2. `POST /v1/apps/{id}/verify` -> `404`.
3. `POST /v1/apps/{id}/deploy-intents/self-host` -> `404`.
4. `POST /v1/apps/{id}/deploy-intents/managed` -> `404`.
5. Replay is not durable across API restart: decision replay returned `404 decision_not_found` after `docker compose restart api`.
6. Anonymous mutating request accepted: `POST /v1/decisions` without auth returned `200`.
7. Decision path does not depend on running Gorse container: `POST /v1/decisions` still returned `200` while `gorse` service was stopped.
8. `gorse` container exited with config parse error (`toml: expected character =`), so demo topology is not healthy end-to-end.

## Verification evidence (2026-02-18 UTC, post-implementation pass)

### Environment
1. Docker compose stack rebuilt and recreated with current API image.
2. Auth token used: `Authorization: Bearer dev-token`.

### Deterministic core checks
1. `POST /v1/decisions` with identical `Idempotency-Key` returned byte-identical payloads.
2. `POST /v1/replay` returned byte-identical payload to original decision response.
3. Replay remained byte-identical after `docker compose restart api`.
4. Anonymous `POST /v1/decisions` returns `401`.

### App builder + verify + deploy checks
1. `POST /v1/apps` -> `201`.
2. `GET /v1/apps/{id}` -> `200`.
3. `PATCH /v1/apps/{id}` -> `200`.
4. `POST /v1/apps/{id}/mutations` with allowed class -> `200`.
5. `POST /v1/apps/{id}/mutations` with disallowed class -> `403`.
6. `POST /v1/apps/{id}/verify` -> `200` with `verdict=pass`.
7. `POST /v1/apps/{id}/deploy-intents/self-host` -> `202`.
8. `POST /v1/apps/{id}/deploy-intents/managed` -> `202`.
9. Cross-tenant read check (`ops-token` against `t_acme` app) -> `404`.

### Agent contract checks
1. `POST /v1/agents/plan` -> `200`.
2. `POST /v1/agents/act` -> `200`.
3. `POST /v1/agents/verify` -> `200` with `verdict=pass`.
4. `POST /v1/agents/deploy` -> `202` with `status=pending_approval`.

## Verification evidence (2026-02-18 UTC, final run)

1. Health: `{"status":"ok","policy_version":"policy-v1","data_version":"data-v1" ...}`.
2. Decision flow: `POST /v1/decisions` with same idempotency key -> `200/200`, payload equality `PASS`.
3. Replay flow: `POST /v1/replay` -> `200`, payload equality `PASS`.
4. Replay after API restart: `POST /v1/replay` -> `200`, payload equality `PASS`.
5. App flow: `POST /v1/apps` -> `201`, verify -> `200(pass)`, deploy self-host -> `202`, deploy managed -> `202`.
6. Agent flow: plan -> `200`, act -> `200`, verify -> `200`, deploy -> `202`.
7. Security checks: unauthenticated mutating call -> `401`; cross-tenant app read -> `404`.
8. UI check: `GET /ui/` serves trial console HTML/CSS/JS for manual end-to-end testing.
9. Root redirect check: `GET /` -> `307` redirect to `/ui/`.
10. Studio APIs: create job (`POST /v1/studio/jobs`), inspect job (`GET /v1/studio/jobs/{id}`), terminal (`POST /v1/studio/jobs/{id}/terminal`), console (`GET /v1/studio/jobs/{id}/console`) return success responses.
11. Studio data check: job creation returned workload (`7` items), files (`5` artifacts), terminal `ls apps` output populated, console logs non-empty.

## Verification evidence (2026-02-18 UTC, SSE streaming pass)

1. `GET /v1/studio/jobs/{id}/events` (with `token` query for browser EventSource compatibility) -> PASS (`200`, `Content-Type: text/event-stream`).
2. Initial SSE `job` event contains full job snapshot (`status`, `workload`, `files`, `terminal_logs`, `console_logs`).
3. After `POST /v1/studio/jobs/{id}/terminal`, next SSE `job` event reflects incremented `terminal_logs` and `console_logs`.
4. `/ui/` stream status shows `connected`; polling fallback remains available when stream errors.
5. `GET /v1/studio/jobs/{id}/preview?client=web` -> PASS (`200`, `Content-Type: text/html`, clickable dashboard/entities/workflows tabs).
6. `GET /v1/studio/jobs/{id}/preview?client=mobile` -> PASS (`200`, `Content-Type: text/html`, clickable bottom-nav mobile views).

## Verification evidence (2026-02-18 UTC, generated runtime preview pass)

1. `GET /v1/studio/jobs/{id}/runtime/web/app.css` -> PASS (`200`, `text/css`).
2. `GET /v1/studio/jobs/{id}/runtime/web/app.js` -> PASS (`200`, `application/javascript`) and includes generated model payload.
3. `GET /v1/studio/jobs/{id}/runtime/mobile/app.css` -> PASS (`200`, `text/css`).
4. `GET /v1/studio/jobs/{id}/runtime/mobile/app.js` -> PASS (`200`, `application/javascript`) and includes generated model payload.
5. `GET /v1/studio/jobs/{id}` now reports runtime source artifacts under `apps/<slug>/clients/web/*` and `apps/<slug>/clients/mobile/*`.

## Verification evidence (2026-02-19 UTC, multi-model adapter and tools pass)

1. `GET /v1/llm/providers` -> PASS (`200`) with `ollama` and `frontier` provider entries.
2. `POST /v1/llm/infer` with `provider=ollama` -> PASS (`200`) when local Ollama model is available; response includes `provider`, `model`, `text`, `latency_ms`.
3. `POST /v1/llm/infer` replay with same `Idempotency-Key` -> PASS (same cached payload returned).
4. `GET /v1/tools` -> PASS (`200`) with tool descriptors for `agent.*`, `llm.providers`, and `llm.infer`.
5. CLI build -> PASS (`make cli` cross-compiles for host platform via containerized Go toolchain).
6. CLI smoke -> PASS (`bin/vda tools list`, `bin/vda llm providers`) against local API with bearer token.

## Verification evidence (2026-02-19 UTC, full GLM end-to-end pass)

1. Ollama inventory -> PASS (`glm-4.7-flash:latest` visible locally).
2. `GET /v1/llm/providers` -> PASS (`ollama.reachable=true`, GLM model present).
3. `POST /v1/llm/infer` with `provider=ollama`, `model=glm-4.7-flash:latest` -> PASS (returned `ok`).
4. `POST /v1/llm/infer` replay with same `Idempotency-Key` -> PASS (`LLM_REPLAY_MATCH=true`).
5. `GET /v1/tools` -> PASS (`TOOLS_COUNT=6`), CLI parity -> PASS (`CLI_TOOLS_STATUS=200`, `CLI_LLM_STATUS=200`).
6. Decision deterministic checks -> PASS (`DEC_MATCH=true`, `REPLAY_MATCH=true`, `decision_id=dec_0054ff04324c74a9`).
7. App lifecycle checks -> PASS (`GET=200`, `PATCH=200`, allowed mutation `200`, disallowed mutation `403`, verify `200`, deploy intents `202/202`).
8. Agent lifecycle checks -> PASS (`plan=200`, `act=200`, `verify=200`, `deploy=202`).
9. Studio checks -> PASS (`job_id=job_388ca308137601fb`, `workload=7`, `files=9`, preview web/mobile `200/200`, runtime JS web/mobile `200/200`, terminal output lines `9`, console lines `6`).
10. SSE check -> PASS (at least one `event: job` observed from `/v1/studio/jobs/{id}/events` stream).

## Verification evidence (2026-02-19 UTC, frontier fix pass)

1. `GET /v1/llm/providers` -> PASS (`ollama.reachable=true`, `frontier.reachable=true`).
2. `frontier.models` includes local GLM via OpenAI-compatible endpoint (`glm-4.7-flash:latest`).
3. `POST /v1/llm/infer` with `provider=frontier`, `model=glm-4.7-flash:latest` -> PASS (`200`, returned `ok`).
4. CLI parity -> PASS (`bin/vda llm infer --provider frontier ...` returned status `200`).

## Verification evidence (2026-02-19 UTC, infer hook to studio generation pass)

### Machine context
1. Timestamp: `2026-02-19T03:45:51Z`
2. Kernel: `Darwin 25.0.0 arm64`
3. Docker: `29.1.5`, Compose: `v5.0.1`

### Mandatory platform checks
1. `docker compose -f docker-compose.demo.yml config` -> PASS.
2. `docker compose -f docker-compose.demo.yml up -d --build api` -> PASS.
3. `GET /v1/health` -> PASS (`200`).
4. Determinism checks:
   - `POST /v1/decisions` twice with same `Idempotency-Key` -> PASS (`200/200`, identical payloads).
   - `POST /v1/replay` with returned `decision_id` -> PASS (`200`, exact payload replay).

### New infer hook checks (expected bridge from model output -> app workload/code preview)
1. `POST /v1/llm/infer` with `post_hooks:["studio_generate"]` and `hook_confirmation` -> PASS (`200`).
2. Response hook payload includes `hooks[0].job_id=job_f6410b7359ff63ac`.
3. Hook summary includes `template=violet-rails-extension` and `source_system=violet-rails`.
4. `GET /v1/studio/jobs/job_f6410b7359ff63ac` -> PASS (`200`) with `workload=9`, `files=10`.
5. Generated artifacts include `apps/violet-crm-pilot/boilerplate/violet_rails_extension.md`.
6. Preview/runtime surfaces:
   - `GET /v1/studio/jobs/{id}/preview?client=web&token=dev-token` -> `200`
   - `GET /v1/studio/jobs/{id}/preview?client=mobile&token=dev-token` -> `200`
   - `GET /v1/studio/jobs/{id}/runtime/web/app.js?token=dev-token` -> `200`
   - `GET /v1/studio/jobs/{id}/runtime/mobile/app.js?token=dev-token` -> `200`
7. UI reachability:
   - `GET /ui/` -> `200`
   - `GET /` -> `307` redirect to `/ui/`

### UI click-flow validation (Playwright CLI, same runtime)
1. Opened `http://localhost:4020/ui/`, filled prompt + app name, and clicked `Run Model Call` with `Auto-generate boilerplate + previews` enabled.
2. During run, button state changed to `Running...` and infer panel displayed `status=running`.
3. Post-run snapshot shows:
   - `Job ID` populated (`job_4971c92043750eb4`)
   - stream status `connected`
   - infer payload includes hook `studio_generate` with `template=violet-rails-extension`
   - generated code list includes `apps/ui-hook-proof-app/boilerplate/violet_rails_extension.md`
   - web and mobile preview iframes rendered and interactive.
4. Snapshot artifact: `.playwright-cli/page-2026-02-19T03-47-59-830Z.yml`.

## Verification evidence (2026-02-19 UTC, real workspace generation pass)

1. `POST /v1/llm/infer` with `post_hooks:["studio_generate"]` -> PASS (`200`) and returned `job_28ba3fe100b095fd`.
2. `GET /v1/studio/jobs/{id}` -> PASS (`200`) with:
   - `workspace_path=output/studio/job_28ba3fe100b095fd`
   - inferred CRM/RBAC workflows (`capture_lead`, `approve_quote`, `manage_roles`, `grant_permissions`)
   - `files=13` including `clients/web/index.html`, `clients/mobile/index.html`, and `package.json`.
3. Generated code quality check:
   - `clients/web/app.js` contains executable runtime UI code (`Runtime Overview`) rather than seed comments.
4. Terminal real execution check:
   - `POST /v1/studio/jobs/{id}/terminal` with `{"command":"exec pwd"}` -> PASS, output `/home/appuser/output/studio/job_28ba3fe100b095fd`.
   - `POST /v1/studio/jobs/{id}/terminal` with `{"command":"exec ls -1 apps"}` -> PASS.
5. Preview parity check:
   - `GET /v1/studio/jobs/{id}/preview?client=web&token=dev-token` -> `200`
   - `GET /v1/studio/jobs/{id}/runtime/web/app.js?token=dev-token` -> `200`.
6. UI click flow rerun (`.playwright-cli/page-2026-02-19T03-59-19-476Z.yml`) confirms:
   - `Workspace` field populated in the UI
   - generated code list includes web/mobile `index.html`, `app.js`, `app.css`, and `package.json`
   - console logs include `workspace materialized` line.

## Verification evidence (2026-02-20 UTC, UI network_error hardening pass)

### Machine context
1. Timestamp: `2026-02-20T03:13:56Z`
2. Kernel: `Darwin 25.2.0 arm64`
3. Docker: `28.0.1`, Compose: `v2.33.1-desktop.1`

### Mandatory platform checks
1. `docker compose -f docker-compose.demo.yml config` -> PASS.
2. `docker compose -f docker-compose.demo.yml up -d` -> PASS (all demo services started).
3. `GET /v1/health` with bearer token -> PASS (`200`).
4. Determinism checks:
   - `POST /v1/decisions` twice with same `Idempotency-Key` -> PASS (`200/200`, identical payload bytes).
   - `POST /v1/replay` with returned `decision_id` -> PASS (`200`, identical `decision_id` and `decision_hash` to original decision payload).

### UI/API client fix verification
1. `web/src/lib/api/client.ts` now classifies failures into:
   - transport failures -> `network_error` with URL context
   - HTTP failures -> preserved status with parsed JSON or raw text details
   - empty/non-JSON payloads -> handled without mislabeling as network failure.
2. `web/src/lib/stores/auth.svelte.ts` normalizes persisted `baseUrl` (trim + trailing slash removal + default fallback).
3. URL joins now use a single helper across fetch/SSE/preview paths to avoid malformed `//v1/...` construction.
4. Frontend quality gates:
   - `npm run check` -> PASS (existing unrelated warning in `Accordion.svelte`)
   - `npm run build` -> PASS.

### Observed contract risk (follow-up)
1. `POST /v1/replay` runtime payload currently does not include `hashes_match`, while `web/src/lib/api/types.ts` still expects it in `ReplayResponse`.
