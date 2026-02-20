# DDIA Audit Snapshot (New Service)

## Intent
Build a deterministic API-first replacement for Violet Rails runtime concerns.

## Reliability
1. Idempotency key contract in request path.
2. Replay store for deterministic response verification.
3. Stable sorting and policy/data versioning.

## Scalability
1. Service-level composition with Gorse + Redis + Postgres.
2. Stateless API tier and horizontally scalable runtime.
3. Future separation of hot decision path vs cold sync/materialization workers.

## Maintainability
1. Explicit adapters for Gorse/GoRules/pipeline.
2. No `eval` execution model.
3. Bounded API surface and versioned contracts.
