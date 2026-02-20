# RFC-0001: Fullstack Violet Rails Output Contract (Prompt -> Real App)

**Date:** 2026-02-19  
**Status:** Proposed  
**Owners:** Platform + Agent Runtime + DX

## 1) Problem Statement

We proved a working POC for:
1. prompt capture,
2. structured confirmation,
3. model call + hook,
4. workload/code/preview surfaces.

But the output still misses the real-world expectation: users want a **usable fullstack SaaS application** (web + mobile + backend + deployable runtime), not only scaffold-like artifacts and synthetic previews.

This RFC defines the product and engineering contract to close that gap.

## 2) Real User Expectation (JTBD)

For this platform to be credible, users must be able to complete these JTBDs from one flow:

1. **Create app from prompt**
When I describe my SaaS idea, I get a runnable app package with meaningful domain flows.

2. **Customize safely**
When I change entities, workflows, policies, roles, and integrations, the app updates without breaking deterministic guarantees.

3. **Validate behavior before deploy**
When I preview web/mobile and run checks, I can trust that what I see is what ships.

4. **Operate with human + AI agents**
When I expose tools/contracts, agents can plan/act/verify/deploy under guardrails.

5. **Ship to managed or self-host**
When I choose deployment mode, I get concrete deploy artifacts and verification gates.

## 3) Goals

1. Output a **materialized fullstack app workspace** per generation job.
2. Support runnable web and mobile clients backed by generated app logic.
3. Generate backend API/tool contracts and deterministic mutation/verify flows.
4. Produce machine-verifiable quality evidence (tests, checks, reports).
5. Keep local-first model routing (Ollama) with frontier escalation support.

## 4) Non-Goals (This RFC Scope Boundary)

1. Full parity with every historical Violet Rails capability in one cycle.
2. Autonomous production deploy without explicit approval gates.
3. Unlimited plugin marketplace or arbitrary runtime code execution.

## 5) Proposed Output Contract

A successful generation job MUST produce:

1. **Repo/workspace path** (real files on disk).
2. **Web client** (runnable).
3. **Mobile client** (runnable/emulatable).
4. **Backend service/tool surface** with deterministic mutation + verify hooks.
5. **Schema + workflow definitions**.
6. **Policy and role model defaults (RBAC)**.
7. **Integration stubs/adapters** for requested providers.
8. **Tests and verification report**.
9. **Deployment artifacts** for self-host and managed targets.
10. **Artifact manifest** with file list, run commands, and coverage mapping.

## 6) Architecture (Target)

### 6.1 Generation Pipeline

Use LangGraph orchestration with explicit nodes:
1. `plan`: interpret prompt + constraints + JTBD target set.
2. `clarify`: request/resolve structured confirmation fields.
3. `design`: produce domain model, workflows, RBAC, integration contract.
4. `scaffold`: create base fullstack workspace.
5. `code`: implement modules and API tools.
6. `verify`: run tests/checks/replay determinism gates.
7. `package`: produce deploy intents + artifact manifest.

### 6.2 Model Routing

1. Local-first: Ollama for high-iteration loops.
2. Frontier lane: OpenAI-compatible endpoints for harder synthesis.
3. Task routing policy:
- structural planning: local
- deep coding/refactor: frontier-eligible
- verify/reporting: local

### 6.3 Runtime Integration

1. Generation workspace is materialized and persisted by job ID.
2. Terminal supports safe built-ins + explicit `exec` mode.
3. Preview renders generated client runtime, not static placeholders.

## 7) API and Schema Changes

## 7.1 Extend Studio Job Model

Add required fields:
1. `artifact_manifest`
2. `run_targets` (web, mobile, api)
3. `verification_report`
4. `jtbd_coverage`
5. `deploy_artifacts`

## 7.2 New/Extended Endpoints

1. `POST /v1/studio/jobs` (accepts generation profile + depth target).
2. `GET /v1/studio/jobs/{id}/artifacts` (manifest, not just raw files).
3. `POST /v1/studio/jobs/{id}/run` (named run targets/check suites).
4. `GET /v1/studio/jobs/{id}/verification` (machine-readable report).
5. `GET /v1/studio/jobs/{id}/jtbd` (coverage and failed jobs-to-be-done).

## 8) UX Contract

The main screen must present:

1. Prompt + structured confirmation alignment.
2. Generation depth toggle (`prototype`, `pilot`, `production-candidate`).
3. Live status for each pipeline node (`plan`, `design`, `code`, `verify`, `package`).
4. Web/mobile previews from generated runtime.
5. Code explorer and real workspace terminal.
6. Verification and JTBD coverage panel.
7. Deploy readiness panel with explicit blockers.

## 9) Definition of Done (Release Gate)

A generation job is considered successful only if all pass:

1. Workspace materialized with full artifact manifest.
2. Web preview interactive and loads generated app behavior.
3. Mobile preview interactive and reflects same domain state.
4. Required test suite passes (`unit + smoke + contract`).
5. Determinism checks pass for mutating API flows.
6. At least 3 declared JTBD scenarios pass in job report.
7. Self-host deploy package generated.
8. Managed deploy intent payload generated with approval gates.

## 10) Delivery Plan (Fixed Time, Variable Scope)

### Phase A (2 weeks): Real Output Foundation
1. Artifact manifest + workspace contract.
2. Generated web/mobile runtime from files (not inline simulation).
3. Job run targets and verification schema.

### Phase B (2 weeks): Fullstack Depth
1. Backend module generation and tool surface synthesis.
2. RBAC + policy workflows from confirmation.
3. Integration adapter stubs and test fixtures.

### Phase C (2 weeks): Ship Readiness
1. JTBD suite + quality gates in pipeline.
2. Deploy artifacts + release checklist automation.
3. Cut non-critical scopes per Shape Up appetite if needed.

## 11) Risks and Cuts

1. **Risk:** model drift causes inconsistent code quality.
- **Mitigation:** deterministic templating + constrained generators + verifier gate.

2. **Risk:** preview success but runtime fails in standalone execution.
- **Mitigation:** run targets must execute from generated workspace before pass.

3. **Risk:** scope explosion to “full parity now”.
- **Mitigation:** enforce phase gates and cut list using fixed-time discipline.

## 12) Immediate Next Actions

1. Convert this RFC into board/ticket updates (epic + dependency graph).
2. Add OpenAPI changes for manifest/run/verification endpoints.
3. Implement pipeline node telemetry in Studio SSE stream.
4. Add CI gate: reject “generated” jobs without passing verification report.

## 13) Decision-Tree Execution Addendum (2026-02-20)

This RFC now links to release issue `R1-021` (`planning/release-r1/tickets.json`) and GitHub issue `#1` (`https://github.com/ChaiWithJai/violet-deterministic-api/issues/1`) to force one explicit execution branch and prevent mixed-strategy drift.

### Branch A: Deterministic Platform First

1. Lock reliability/security/performance gates first.
2. Complete migration parity/export-import path before deeper template breadth.
3. Defer template-depth expansion unless release gates are green.

### Branch B: Template-Grade Output First

1. First eliminate UI/API contract drift (plan/clarify/studio response shapes).
2. Pick one canonical template lane (no multi-template spread in same cycle).
3. Add real auth/tenant/billing/domain modules while preserving determinism gates as hard blockers.

### Required Branch Selection Rule

1. Exactly one branch is active per cycle.
2. Scope from the non-selected branch is cut or explicitly deferred.
3. Advancement requires passing evidence from `/v1/studio/jobs/{id}/run` target `all` and matching verification/JTBD outputs.

### Current Gap Inventory Summary

1. Platform determinism is materially implemented (`idempotency`, `replay`, tenant auth, studio run/verification/jtbd/bundle).
2. Generated fullstack output remains scaffold-heavy versus template-grade baselines.
3. Migration parity endpoints and production-hardening evidence are still outstanding gates.
4. End-user harness PRD target (`/ui/harness.html`) is not currently served; operator UI remains `/ui/`.

## 14) Violet Rails Parity Gap Analysis (2026-02-19)

This section formalizes the results of a systematic parity audit between VDA's implemented API surface and Violet Rails' production capabilities. The audit was conducted by comparing every VDA endpoint handler against the Violet Rails route map (`violet_routes.rb`).

### 14.1 Parity Matrix

| VDA Endpoint Family | What VDA Does Today | Violet Rails Equivalent | Parity | Shortfall |
|---|---|---|---|---|
| `/v1/decisions`, `/v1/replay`, `/v1/feedback` | Deterministic decisioning, stored replay, idempotent feedback | Dynamic API namespace/resource layer | Partial | Different paradigm: VDA does recommendation ranking, not broad data CRUD |
| `/v1/apps`, mutations, verify, deploy-intents | Blueprint CRUD, constrained mutations (4 classes), verify, deploy intent | Namespaces, resources, forms, settings/workflow | Partial | Mutation model is narrow vs Violet's dynamic resource/action system |
| `/v1/agents/{plan,clarify,act,verify,deploy}` | Full agent choreography with actor attribution | No direct equivalent | **VDA Advantage** | Net-new capability |
| `/v1/llm/providers`, `/v1/llm/infer` | Provider discovery, inference, `studio_generate` hook | No first-class equivalent | **VDA Advantage** | Net-new capability |
| `/v1/studio/jobs` + sub-endpoints | Workspace generation, artifacts, run targets, verification, JTBD, preview, bundle, terminal, console, SSE | Built-in product surfaces (CMS, blog, forum, mailbox) | Partial | Output is scaffold-heavy; verification is structural not behavioral |
| `/v1/migration/violet/{export,import}` | Planned (R1-014, R1-015) but **not implemented** | Export/import implied by deprecation plan | **Missing** | No programmatic migration path exists |
| Auth + governance | Token-based tenant auth (`token:tenant_id:subject`) | Devise, invites, OTP, sysadmin, subdomain | Partial | Strong request auth but no user lifecycle/admin/invite UX |
| Content/community/email | None | CMS, blog, forum, mailbox/email tracking | **Missing** | Major gap vs "out-of-box SaaS platform" promise |
| Analytics/GraphQL/ops | None | Ahoy analytics, Blazer, GraphQL, Sidekiq web | **Missing** | Operational tooling not present |

### 14.2 Five Shortfall Areas (tracked as R1-022 through R1-026)

**Shortfall 1: Built-in product primitives** (GitHub #4, R1-022)
- CMS, blog, forum, and email are zero in VDA's API surface.
- Recommended path: generate as Studio output modules, not native platform features.
- Architectural rationale: dynamic content rendering conflicts with replay-safe determinism.

**Shortfall 2: Dynamic API namespace/resource model** (GitHub #5, R1-023)
- VDA's 4-class mutation model (`set_name`, `set_plan`, `set_region`, `set_feature_flag`) vs Violet's arbitrary resource CRUD.
- Recommended path: accept as architectural boundary — rich data operations live in generated-app runtime, not VDA control plane. Evaluate schema-driven mutation expansion for R2.

**Shortfall 3: Generated app depth** (GitHub #6, R1-024)
- Verification is structural (schema/policy/deploy_preflight), not behavioral.
- Recommended path: generate behavioral test fixtures alongside code; evolve run targets to execute behavioral tests; add depth labels (prototype/pilot/production-candidate).

**Shortfall 4: User lifecycle governance** (GitHub #7, R1-025)
- Token-based control-plane auth is solid but no end-user auth modules (registration, login, roles, invitations).
- Recommended path: generate auth modules from confirmation metadata; document control-plane vs app-level auth separation.

**Shortfall 5: Migration parity** (GitHub #8, R1-026)
- Export/import endpoints are documented as required gates but have zero implementation.
- Recommended path: elevate to P0; implement with roundtrip fidelity tests; block deprecation timeline on passing fixtures.

### 14.3 Intentional Trade-Off

Some gaps are architectural choices, not oversights:
- **No dynamic runtime eval** preserves determinism and replay safety.
- **Constrained mutation classes** keep the control plane auditable.
- **Token-only auth** preserves replay fidelity (no session state to leak).

The correct framing: VDA is a **deterministic build/deploy engine**, not a dynamic application runtime. Where Violet Rails is a monolith that runs apps, VDA is a factory that generates, verifies, and deploys them.

### 14.4 Risk Register Additions

| Risk ID | Title | Severity | Exit Criteria |
|---|---|---|---|
| RISK-006 | Product primitive gap blocks tenant migration | High | Migration guide published for workaround paths |
| RISK-007 | Scaffold-level depth undermines credibility | High | Generated job passes behavioral JTBD scenarios |
| RISK-008 | Migration endpoint absence blocks deprecation | Critical | Roundtrip export→import→export produces identical bundles |

### 14.5 Board Impact

- **R1-024** (depth) and **R1-026** (migration) added to **Uphill** — these are P0 blockers.
- **R1-022** (primitives), **R1-023** (mutation model), **R1-025** (auth governance) added to **Bet Accepted** — P1, scoped for current cycle with cut option.
- Circuit breaker rule remains: if S1/S2 unknowns unresolved by 2026-03-13, cut S5/S6 breadth.

---

**Bottom line:** POC validated direction; this RFC changes success criteria from "generated something" to "generated a runnable, verifiable, deployable fullstack app that satisfies explicit JTBDs." The parity analysis confirms: VDA's deterministic core is materially implemented, but five shortfall areas must be addressed — either by building, generating, or explicitly deferring — before "replace Violet Rails" is credible.
