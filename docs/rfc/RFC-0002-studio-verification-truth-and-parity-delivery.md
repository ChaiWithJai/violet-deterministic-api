# RFC-0002: Studio Verification Truth and Parity Delivery Gate

- Status: Proposed
- Date: 2026-02-20
- Owner: Platform + QA Platform
- Related incidents:
  - [INC-2026-02-20-studio-verification-truth-gap](../field-reports/INC-2026-02-20-studio-verification-truth-gap.md)
  - [GitHub issue #13](https://github.com/ChaiWithJai/violet-deterministic-api/issues/13)

## Problem Statement
Studio currently allows a verification report to present `verdict=pass` based on structural artifact checks while runtime `run target=all` can still fail for the same job revision. This breaks the product expectation of "working fullstack app at Violet Rails-level depth."

[PM COMMENT] `pass` must mean "safe to proceed" for operators. Any ambiguity here is a product trust break, not a cosmetic bug.
[PRE-MORTEM] If we keep structural pass and behavioral pass as separate truth systems, release pressure will optimize for the easier pass and regress trust again.

## Goals / Non-Goals
Goals:
1. Make verification truthfulness deterministic and runtime-backed.
2. Remove domain-hardcoded runtime probes from Studio runtime checks.
3. Ensure generated OpenAPI is route-complete relative to generated runtime.
4. Define a release-plan path from scaffold-level output to parity-credible output.

Non-Goals:
1. Full feature parity with every Violet Rails subsystem in one patch.
2. Replacing deterministic control-plane constraints with dynamic runtime evaluation.
3. Introducing opaque quality signals that cannot be replayed.

[PM COMMENT] Scope is fixed-time, variable-scope. Truth-gate work is not cuttable for R1.

## Proposed Decision
- Decision 1: Verification verdict is authoritative only when tied to latest `run target=all` result for the same job revision.
- Decision 2: Runtime smoke checks must derive probe entities from generated confirmation/runtime spec, not hardcoded defaults.
- Decision 3: Generated OpenAPI must include all generated runtime routes (entities, actions, primitives, identity, tools, health, workflows).
- Decision 4: R1 board gets a dedicated parity closure lane with explicit P0/P1 tickets and risk linkage.

[PM COMMENT] This decision preserves deterministic architecture while raising the quality bar from "files exist" to "behavior passes."
[PRE-MORTEM] If route/spec parity is not CI-enforced, specs will drift and operators will again distrust output claims.

## SDLC Guardrails
- CI checks:
  - Generate fixture job with non-`account` entity set and run `target=all`; fail if entity runtime check assumes `account`.
  - Fail if `verification_report.verdict=pass` while latest `run target=all` for same revision is fail/missing.
  - Fail if generated runtime routes are absent from generated OpenAPI parity assertions.
- Release checks:
  - At least one generated job at `pilot` depth with `target=all` pass and no pending runtime checks.
  - Incident closure checklist from #13 complete before R1 release-candidate gate.
- Ownership:
  - Studio runtime + QA platform joint ownership for gate logic.
- Evidence requirements:
  - Persisted verification/run evidence linked by job revision and generated timestamp.

[PRE-MORTEM] If controls are advisory only, they will fail under delivery pressure.

## Incident Mapping
- #13 / Occurrence A (verification truth gap) -> Control: runtime-backed verification verdict gate.
- #13 / Occurrence B (hardcoded entity probe) -> Control: entity-aware smoke probe derivation.
- #13 / Occurrence C (spec/runtime drift) -> Control: route-complete OpenAPI + parity assertions.

[PM COMMENT] Closure requires code, tests, and board movement; docs-only closure is invalid.

## Rollout Plan
- Phase 1 (Containment, P0):
  - Fix entity-aware runtime probe.
  - Add regression tests for non-`account` entity generation.
- Phase 2 (Truth Gate, P0):
  - Bind verification verdict to latest run-all evidence.
  - Expose runtime-evidence status in verification payload.
- Phase 3 (Parity Contract, P0/P1):
  - Route-complete OpenAPI generation.
  - Lift primitives/identity from seeded stubs toward stateful behavioral modules.
  - Add benchmark/parity scorecard against target expectation matrix.

## Success Metrics
- Compliance:
  - 100% of `verdict=pass` reports have same-revision `run target=all` pass evidence.
- Repeat incident rate:
  - Zero recurrences of "verification pass / run-all fail" across release fixtures.
- Escape rate:
  - Zero runtime probe failures caused by hardcoded entities in CI suite.

## Risks / Tradeoffs
- Risk: Runtime-backed verification increases compute time.
  - Mitigation: Cache per-job revision run-all results; invalidate on job mutation.
- Risk: Route-complete spec generation adds implementation complexity.
  - Mitigation: Derive specs from route registry templates rather than parallel manual definitions.
- Risk: P1 parity breadth crowds out deterministic core hardening.
  - Mitigation: Keep P0 truth gates mandatory; cut P1 breadth before schedule extension.

## Decision Requested
- Approve adoption of runtime-backed verification truth gate for R1.
- Approve new R1 tickets R1-027 through R1-032 and associated risk register entries.
- Approve issue #13 as release-blocking until P0 acceptance criteria are met.
