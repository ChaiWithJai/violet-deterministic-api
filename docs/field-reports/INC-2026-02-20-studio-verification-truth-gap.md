# Post-Mortem v2: Studio Verification Truth Gap

## Incident Metadata
- Incident ID: `INC-2026-02-20-studio-verification-truth-gap`
- Date: 2026-02-20
- Severity: High
- Class: Governance and quality-gate failure
- Status: Open
- Customer impact: `indirect`

## Executive Summary
- What failed: A Studio verification report (`studio_vrf_1f1411cbc64b491e`) reported `verdict=pass`, but runtime `run target=all` failed for the same job revision.
- Why it mattered: The product promise is "prompt to working app"; false-positive verification undermines trust and release gating.
- What changed: Incident issue #13 was filed, and R1 planning is being updated to make runtime evidence the authoritative gate.

## Customer / Business Impact
- Customer-facing effects:
  - Operators can see a "pass" verification state while actual runtime behavior fails.
  - This appears as "vaporware vs bug" from the operator perspective.
- Business/system risk:
  - Erodes confidence in parity claims versus Violet Rails-level output.
  - Risks shipping scaffold-depth output under a behavioral-ready label.

## Detection and Timeline
- Detection path:
  - Operator question on report `studio_vrf_1f1411cbc64b491e` triggered direct runtime validation.

| Timestamp (UTC) | Event |
|---|---|
| 2026-02-20T13:23:00Z | Job persisted with verification report `studio_vrf_1f1411cbc64b491e` and `verdict=pass` |
| 2026-02-20T13:26:35Z | `POST /v1/studio/jobs/job_a58ddf535251964d/run` (`target=all`) returned `status=fail` |
| 2026-02-20 | Incident filed as GitHub issue #13 |

## Thread of Execution
- Entry point: `POST /v1/studio/jobs/{id}/run` -> `internal/http/studio_handlers.go`
- Service orchestration: `RunTarget` in `internal/studio/service.go:296`
- Runtime checks: `runGeneratedAPIRuntimeChecks` + `checkEntityRecords` in `internal/studio/runtime_exec.go:30` and `internal/studio/runtime_exec.go:284`
- Verification generation: `buildVerificationReport` in `internal/studio/rfc_contracts.go:122`
- Generated runtime behavior: entity route validation in `internal/studio/backend_artifacts.go:298`

## Root Cause Analysis
### Occurrence A
- Problem surface:
  - Verification pass can be derived from artifact presence without enforcing runtime execution pass.
- Evidence anchors:
  - `internal/studio/rfc_contracts.go:122`
  - `internal/studio/service.go:311`
  - `internal/studio/service.go:317`
- Code-smell label:
  - `truth-gap-contract`
- Code Complete challenge class:
  - `integration-contract-mismatch`
- Counter-signal:
  - Runtime checks already exist and can catch real behavioral failures.

### Occurrence B
- Problem surface:
  - Runtime smoke check hardcodes `account` entity path even though generated entities are prompt-dependent.
- Evidence anchors:
  - `internal/studio/runtime_exec.go:286`
  - `internal/studio/backend_artifacts.go:16`
  - `internal/studio/backend_artifacts.go:304`
- Code-smell label:
  - `hardcoded-domain-assumption`
- Code Complete challenge class:
  - `generalization-break`
- Counter-signal:
  - Generated runtime already declares entity list and can be queried dynamically.

### Occurrence C
- Problem surface:
  - OpenAPI generation under-documents runtime routes, enabling spec/runtime drift.
- Evidence anchors:
  - `internal/studio/service.go:384`
  - `internal/studio/backend_artifacts.go:142`
- Code-smell label:
  - `spec-runtime-drift`
- Code Complete challenge class:
  - `documentation-contract-drift`
- Counter-signal:
  - Runtime route registration is centralized and can be used to drive spec generation.

## Stochastic Lens
- The failure is data-dependent: if generated entities include `account`, smoke checks pass; otherwise they fail.
- This creates nondeterministic operator trust outcomes for semantically similar jobs.

## Inversion Lens
- How to guarantee this failure happens again:
  - Keep verification and runtime checks decoupled.
  - Keep hardcoded probe entities in runtime smoke checks.
  - Keep OpenAPI generation independent from runtime route registration.
- Controls that prevent this path:
  - Single gate where verification verdict depends on latest runtime `all` result.
  - Probe target derivation from generated entities.
  - Route/spec parity test in CI.

## Mitigations + Corrective Actions
### Patch Table
| Area | Planned change | Why |
|---|---|---|
| Runtime smoke checks | Remove hardcoded entity probe | Prevent false fail for non-`account` apps |
| Verification verdict | Gate on runtime `target=all` result | Make "pass" trustworthy |
| Generated OpenAPI | Route-complete generation | Remove spec/runtime drift |

### Action Table
| Action | Owner | Status | Due |
|---|---|---|---|
| File incident issue and evidence | Platform | Done | 2026-02-20 |
| Add R1 tickets for truth gate and parity closure | Platform | In progress | 2026-02-20 |
| Implement entity-aware runtime probe | Studio runtime | Planned | 2026-02-21 |
| Add verification/run alignment tests | QA platform | Planned | 2026-02-21 |

## Evidence Provenance
- GitHub issue: [#13](https://github.com/ChaiWithJai/violet-deterministic-api/issues/13)
- Report ID: `studio_vrf_1f1411cbc64b491e`
- Job ID: `job_a58ddf535251964d`
- Runtime call evidence:
  - `POST /v1/studio/jobs/job_a58ddf535251964d/run` with `target=all` returned `status=fail` and failed `api_runtime_entity_records`.
