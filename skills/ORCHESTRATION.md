# Skill Orchestration (Canonical)

Use this order to execute release work with minimum ambiguity.

## Phase 0: Understand the system
1. Run `codebase-archaeology`.
2. Produce or refresh archaeology report and subsystem risk map.

## Phase 1: Audit architecture with DDIA lens
1. Run `ddia-api-first-audit`.
2. Score reliability, scalability, and maintainability by subsystem.
3. Capture gaps and mitigations in `docs/DDIA_AUDIT.md`.

## Phase 2: Shape the release bet
1. Run `shapeup-release-r1`.
2. Keep six-week appetite fixed; cut scope when pressure appears.
3. Update `planning/release-r1/board.json` and `planning/release-r1/tickets.json`.

## Phase 3: Build deterministic core
1. Run `deterministic-go-api-implementation`.
2. Implement deterministic contracts first: idempotency, replay, hash stability, tenant boundaries.

## Phase 4: Deliver and verify tickets
1. Run `ticket-investigation` per ticket.
2. Run `platform-verification` for local truth statements.
3. Run `debug-tracing` when expected and actual diverge.

## Phase 5: Document decisions and contracts
1. Run `adr-creation` for material decisions and findings.
2. Keep ADR index current in `docs/adr/`.

## Phase 6: Agent and migration readiness
1. Run `agent-orchestration` to validate plan/act/verify/deploy APIs.
2. Run `violet-migration-parity` to validate export/import parity and cutover gates.

## Exit criteria
1. Board has only `done` or `cut` tickets.
2. Determinism mismatch and SLO gates meet thresholds.
3. Handoff docs are complete for next execution agent.
