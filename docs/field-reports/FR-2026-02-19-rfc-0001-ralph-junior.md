# Field Report: RFC-0001 Reality Check (Ralph Junior Cycle)

**Date:** 2026-02-19  
**Repository:** `/Users/jaibhagat/code/violet-deterministic-api`  
**Mode:** `ralph-mode` execution loop (implement + verify + document)

## Mission

Test how close current system behavior is to RFC-0001 ("Prompt -> real fullstack Violet Rails output"), identify footguns, and document concrete limits.

## Environment

1. Stack: `docker compose -f docker-compose.demo.yml up -d`
2. API base: `http://localhost:4020`
3. Auth: `dev-token`, `ops-token`
4. Model: `ollama`, `glm-4.7-flash:latest`

## Experiment Matrix

1. Deep generation via `POST /v1/llm/infer` + `post_hooks:["studio_generate"]`
2. Job artifact inspection via `GET /v1/studio/jobs/{id}`
3. Preview/runtime checks for web/mobile
4. Terminal execution checks (`exec pwd`, `exec ls`, runtime toolchain probes)
5. RFC endpoint existence checks (`/artifacts`, `/run`, `/verification`, `/jtbd`)
6. Restart durability check for Studio jobs
7. Tenant isolation check for Studio jobs

## Key Results

1. Generation hook works and produces jobs.
Evidence: infer `200`; `job_a=job_3d6b0c33aa3cc699`.

2. Idempotency replay works for infer with same key.
Evidence: `idempotent_replay_match=true`.

3. Same prompt with different idempotency keys creates different job IDs.
Evidence: `job_a != job_b` (`job_3d6b0c33aa3cc699` vs `job_5152680d0ef3ddb2`).

4. Generated artifacts are materially better than earlier POC.
Evidence: web/mobile `index.html`, `app.js`, `app.css`, and `package.json` present; `files_count=13`.

5. Preview endpoints work.
Evidence: web preview `200`, mobile preview `200`, runtime web js `200`.

6. Critical RFC surfaces are not implemented yet.
Evidence:
- `/v1/studio/jobs/{id}/artifacts` -> `404`
- `/v1/studio/jobs/{id}/run` -> `405`
- `/v1/studio/jobs/{id}/verification` -> `404`
- `/v1/studio/jobs/{id}/jtbd` -> `404`

7. Studio jobs are not durable across API restart.
Evidence: after `docker compose restart api`, previously created job -> `404`.

8. Runtime environment cannot execute generated Node workflows.
Evidence:
- `exec node --version` -> `node: not found`
- `exec npm --version` -> `npm: not found`
- `exec python3 --version` -> `python3: not found`

9. Terminal `exec` is currently over-permissive.
Evidence: `exec cat /etc/passwd | head -n 1` succeeded (`root:x:0:0:...`).

10. LLM response text quality is inconsistent for direct reuse.
Evidence: `result.text` length `0` while latency was `12850ms` and output resided in raw/thinking channel.

11. Tenant isolation on Studio job lookup is correct.
Evidence: `ops-token` reading `dev-token` job -> `404`.

## How Far We Can Take It Today

1. Good for: prompt-to-structured-generation demo, inspectable artifacts, live web/mobile preview, and deterministic API shell validation.
2. Partial for: fullstack app depth (backend behavior, tests, deploy artifacts are incomplete).
3. Not ready for: production-candidate output gate promised in RFC-0001.

## Footguns (Priority-Ordered)

## P0

1. **Non-durable Studio state**
Impact: generated app state disappears on restart, breaking trust and handoff.
Observed: `GET /v1/studio/jobs/{id}` -> `404` after API restart.
Required fix: persist Studio jobs/artifacts in durable storage (DB + object/filesystem volume).

2. **Missing RFC gate endpoints**
Impact: cannot prove completion of artifact manifest/run/verification/JTBD.
Observed: `404/405` on required RFC endpoints.
Required fix: implement endpoint suite and OpenAPI contracts.

3. **Unsafe terminal execution**
Impact: container internals exposed; escalation risk.
Observed: `/etc/passwd` readable via `exec`.
Required fix: command allowlist + jailed runner + path sandbox + redaction.

## P1

1. **Generated app not actually runnable as advertised by scripts**
Impact: `package.json` scripts suggest runnable flow, but runtime lacks `node/npm/python3`.
Observed: toolchain probes return not found.
Required fix: install runtime toolchain or change scripts to supported binaries; validate with run targets.

2. **No deploy/test artifacts in generated output**
Impact: cannot satisfy "deployable fullstack app" JTBD.
Observed: no deploy manifests, no test/spec artifacts.
Required fix: generate baseline deploy/test packs and enforce via manifest checks.

3. **Non-reproducible job identity across equivalent prompts**
Impact: hard to compare outputs and cache reliably outside idempotency-key replay.
Observed: different job IDs for equivalent inputs.
Required fix: add reproducible build signature and optional deterministic build mode.

## P2

1. **Model output channel mismatch**
Impact: codegen quality/parsing can fail when `text` is empty and content is buried in provider raw format.
Observed: `text_len=0`.
Required fix: robust content extraction normalization per provider.

2. **Workspace path visibility confusion**
Impact: path shown in UI is inside container, not host path.
Observed: `workspace_visible_on_host=false`.
Required fix: explicit label in UI + optional host bind mount mapping in compose.

## Field Notes: What Worked Well

1. Infer hook pattern is productive and user-visible.
2. Preview integration is fast and gives immediate confidence loops.
3. Prompt-driven domain signal extraction (CRM/RBAC/mobile/agent hints) increases relevance over plain scaffolds.
4. Tenant isolation behavior stayed correct during tests.

## Field Notes: Where "Almost Works" Trap Happened

1. Build/generation looked successful, but runtime lacked required toolchain.
2. Demo appeared persistent during a session, but restart invalidated all Studio job state.
3. Output looked fullstack-ish in UI, but verification and JTBD gates were absent.

## Recommended Next Implementation Sequence (RFC-0001)

1. Add durable Studio storage model (jobs + files + manifests).
2. Implement `/artifacts`, `/run`, `/verification`, `/jtbd` endpoints.
3. Add artifact manifest schema and quality gate.
4. Introduce restricted command runner for terminal.
5. Add runnable target executor with explicit runtime dependencies.
6. Add deploy/test artifact generation pass.
7. Add CI check to fail generation when verification/JTBD gates are missing.

## Raw Evidence Snapshot

1. Infer status: `200`, replay match: `true`, latency: `12850ms`.
2. Generated job example: `job_3d6b0c33aa3cc699`, files: `13`, workload: `13`.
3. Restart durability: job lookup `404` after restart.
4. Cross-tenant Studio access: `404`.
5. Missing RFC endpoints: `404/405`.

---

**Verdict:** POC is real and useful for interactive shaping, but still below RFC-0001 release bar. Biggest blockers are durability, missing verification/JTBD surfaces, and runtime safety/executability.

## Addendum (Cycle 2, 2026-02-19)

Follow-up implementation closed the major blockers listed above:

1. Studio jobs are persisted in Postgres (`studio_jobs`) and reloaded after API restart.
2. RFC endpoints were implemented:
   - `GET /v1/studio/jobs/{id}/artifacts`
   - `POST /v1/studio/jobs/{id}/run`
   - `GET /v1/studio/jobs/{id}/verification`
   - `GET /v1/studio/jobs/{id}/jtbd`
3. Terminal `exec` is now allowlisted and blocks shell metacharacters, absolute paths, and parent traversal.
4. Generated output now includes deploy + test artifacts (`deploy/*`, `tests/smoke.yaml`).
5. UI now includes a release gate panel to run targets and inspect artifact/verification/JTBD evidence.
6. Direct Studio generation now enforces deterministic baseline constraints (`all_mutations_idempotent`, `no_runtime_eval`) so `Run All` and verification do not fail due empty constraint input.

Validation evidence snapshot (cycle 2):

1. `HEALTH=ok`
2. `VERDICT=pass`
3. `RUN_ALL_STATUS=pass`
4. `CONSTRAINTS=["all_mutations_idempotent","no_runtime_eval"]`

## Addendum (Cycle 3, 2026-02-19)

Depth and portability upgrades were added to generated app output:

1. Generated workspace now includes backend runtime scaffold under `services/api`:
   - `go.mod`, `cmd/server/main.go`, runtime handlers, policy rules, tools catalog, integration adapters, smoke script.
2. Verification and `run target=api` now require backend runtime/test artifacts, not only OpenAPI placeholders.
3. Studio can now export the generated app as a tarball:
   - `GET /v1/studio/jobs/{id}/bundle`
   - Includes generated files + `studio_artifact_manifest.json`.
4. UI now exposes a bundle download link for each generated job.

Validation evidence snapshot (cycle 3):

1. `FILES_TOTAL=28`
2. `BACKEND_FILES=12`
3. `API_RUN_STATUS=pass`
4. `VERIFY_BACKEND=pass`
5. `BUNDLE_STATUS=200`
6. `BUNDLE_MAGIC=1f8b` (gzip)

## Addendum (Cycle 4, 2026-02-19)

Execution depth for the API run target now performs real runtime smoke checks:

1. `run target=api` executes inside the studio environment:
   - `go test ./...` in generated `services/api`
   - boot generated server with ephemeral port
   - probe `/health`
   - probe `/v1/tools`
   - execute `/v1/workflows/execute`
2. `run target=all` inherits the same runtime checks.
3. API runtime image now includes Go toolchain to support generated service execution.

Validation evidence snapshot (cycle 4):

1. `RUN_API_STATUS=pass`
2. `RUN_API_CHECK_api_runtime_go_test=pass`
3. `RUN_API_CHECK_api_runtime_health=pass`
4. `RUN_API_CHECK_api_runtime_tools_catalog=pass`
5. `RUN_API_CHECK_api_runtime_workflow_execute=pass`

## Addendum (Cycle 5, 2026-02-19)

One-command local launch flow now exists for generated jobs:

1. New CLI command:
   - `vda studio launch --job-id <job_id>`
2. Workflow executed by command:
   - downloads Studio bundle
   - extracts workspace under `output/launch/...`
   - starts generated backend service (`go` if available, otherwise docker fallback)
   - starts web and mobile static preview servers
   - prints live URLs and keeps process alive until Ctrl+C
3. Tool catalog advertises `studio.launch` for discoverability.

Validation evidence snapshot (cycle 5):

1. `API health -> ok` on launched generated service
2. `API tools count -> 4`
3. `WEB=200`
4. `MOBILE=200`
5. Ctrl+C cleanly stops API, web, and mobile launch processes
