# Skill: DDIA API-First Audit

Audit architecture and implementation against DDIA principles for reliability, scalability, and maintainability.

## Inputs
1. `docs/DDIA_AUDIT.md`
2. `docs/EXTRACTION_ROADMAP.md`
3. `docs/DEPRECATION_PLAN.md`
4. `planning/release-r1/*.json`

## Reliability checks
1. Idempotency semantics and key scoping.
2. Deterministic replay fidelity and snapshot metadata.
3. Failure handling for dependency outages (gorse/gorules/storage).
4. Tenant isolation and auditability.

## Scalability checks
1. Stateless API tier boundaries.
2. Hot-path vs cold-path separation.
3. Backpressure, timeout, and queueing strategy.
4. Data model partitioning and index strategy.

## Maintainability checks
1. Clear adapter boundaries for gorse/gorules/go-pipeline.
2. Versioned contracts and deprecation flow.
3. Testability and fixture-based replay harness.
4. Operational runbooks and observability coverage.

## Output
Update `docs/DDIA_AUDIT.md` with:
1. red/yellow/green score per axis,
2. concrete gap list,
3. bounded remediation steps mapped to ticket IDs.
