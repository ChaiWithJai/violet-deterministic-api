# PRD: End-User Web Client Harness (Lovable-Style Flow)

**Status:** Draft for execution  
**Owner:** Product + Platform + Studio Runtime  
**Date:** 2026-02-19

## 1. Problem

Current Studio is powerful for operators, but end users need a guided, confidence-building flow that feels like:

1. "I describe my app idea in plain language"
2. "I confirm the structure quickly"
3. "I see the app taking shape live"
4. "I trust quality before shipping"

Without this harness, users see too much implementation detail too early, and cannot tell whether outputs are production-intent or scaffold noise.

## 2. Goals

1. Provide a prompt-first web client for end users.
2. Preserve structured confirmation and deterministic constraints.
3. Show live web/mobile previews and execution evidence in one screen.
4. Keep code and console inspectable without forcing users into operator-only UX.
5. Create a direct bridge from intent -> verified output -> launch/download.

## 3. Non-Goals

1. Full autonomous production deployment in this phase.
2. Replacing the existing operator Studio console (`/ui/`) in this phase.
3. Multi-user collaboration/session sharing in this phase.

## 4. Primary User JTBD

1. When I describe my SaaS idea, I want a generated app plan and scaffold I can verify quickly.
2. When I adjust scope, I want the generated output to stay aligned to my intent.
3. When I review previews and checks, I want clear pass/fail signals before using the result.
4. When I accept output, I want a bundle or launch path immediately.

## 5. Experience Principles

1. Prompt-first: one large intent entry as primary action.
2. Progressive disclosure: show detail only when needed.
3. Trust through evidence: every generation has verification/JTBD visibility.
4. One surface, many loops: prompt, align, generate, preview, verify, download.

## 6. Functional Requirements

### 6.1 Prompt + Drafting

1. User can submit prompt and receive a draft scope (`/v1/agents/plan`).
2. UI applies suggested blueprint defaults (name/plan/region).
3. UI provides signal cards (workflows/entities/constraints) to make implicit assumptions explicit.

### 6.2 Confirmation

1. User can edit structured confirmation fields.
2. User can run a chat-style clarification loop to answer targeted follow-up questions and map answers back into the confirmation schema (`/v1/agents/clarify`).
3. User can generate directly from confirmation (`POST /v1/studio/jobs`).
4. User can generate from prompt via model hook (`POST /v1/llm/infer` with `studio_generate`).

### 6.3 Live Build Harness

1. Show job status, stream state, and workspace path.
2. Subscribe to job events (`/v1/studio/jobs/{id}/events`).
3. Render live web/mobile previews (`/preview?client=web|mobile`).
4. Render code explorer from generated file list.
5. Render console logs and workload timeline.

### 6.4 Quality and Readiness

1. Run targets from harness (`web/mobile/api/verify/all`).
2. Show verification and JTBD reports inline.
3. Expose downloadable bundle (`/v1/studio/jobs/{id}/bundle`).

### 6.5 Accessibility and Responsiveness

1. Must work on desktop and mobile viewports.
2. Focus states and semantic labels required for inputs/buttons.

## 7. API Dependencies

1. `POST /v1/agents/plan`
2. `POST /v1/agents/clarify`
3. `POST /v1/llm/infer`
4. `POST /v1/studio/jobs`
5. `GET /v1/studio/jobs/{id}`
6. `GET /v1/studio/jobs/{id}/events`
7. `GET /v1/studio/jobs/{id}/preview`
8. `GET /v1/studio/jobs/{id}/artifacts`
9. `POST /v1/studio/jobs/{id}/run`
10. `GET /v1/studio/jobs/{id}/verification`
11. `GET /v1/studio/jobs/{id}/jtbd`
12. `GET /v1/studio/jobs/{id}/bundle`

## 8. Success Metrics

1. Time from first prompt -> first clickable preview under 90 seconds (p50 local environment).
2. At least 80 percent of generated runs include a quality gate check execution (`run all`).
3. Verification pass rate visible for every generated job.
4. Bundle download action available for every generated job.

## 9. Risks and Mitigations

1. **Risk:** Model output inconsistency causes confusing draft quality.
   - **Mitigation:** Keep confirmation editable and visible before final generation.
2. **Risk:** Runtime execution depends on local toolchain/runtime availability.
   - **Mitigation:** explicit check outputs in run target evidence; fallback paths documented.
3. **Risk:** End-user harness becomes too operator-heavy.
   - **Mitigation:** progressive disclosure and clear user-facing language.

## 10. Phase Plan

### Phase A (Now)

1. Ship dedicated harness screen under `/ui/harness.html`.
2. Wire prompt -> clarify -> confirm -> generate -> preview -> verify loop.
3. Include bundle and run-target controls.

### Phase B

1. Add saved clarification sessions and compare revisions.
2. Add richer alignment diffs across generations.
3. Add artifact quality scorecards across generated app revisions.

### Phase C

1. Optional managed deploy handoff wizard.
2. Team collaboration and shareable links.

## 11. Acceptance Criteria

1. End user can complete full loop on one screen:
   - prompt
   - clarify
   - confirmation
   - generation
   - preview
   - quality gate
   - bundle access
2. Harness loads at `/ui/harness.html` with no backend route changes required.
3. Existing operator Studio (`/ui/`) remains available.
4. README links to the new PRD and harness entrypoint.
