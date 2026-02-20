# Skill: Deterministic Go API Implementation

Implement deterministic API behaviors with explicit contracts and adapter seams.

## Architectural intent
1. Go control/data plane with deterministic request processing.
2. Adapter seams for `gorse`, `gorules`, and `go-pipeline`.
3. Machine-readable API contracts for human and AI operators.

## Non-negotiables
1. Stable hashing from canonical request normalization.
2. Idempotency on all mutating endpoints.
3. Replay fidelity for stored decision snapshots.
4. Explicit `policy_version`, `data_version`, `decision_id`, and timestamp in responses.

## Implementation order
1. Request normalization and hashing utilities.
2. Durable idempotency + replay storage.
3. Auth and tenant scoping.
4. Gorse adapter integration.
5. GoRules adapter integration.
6. Pipeline stage tracing and observability.

## Output
1. Endpoint contracts updated in `api/openapi.yaml`.
2. Tests for determinism and replay behavior.
3. Runbook updates with smoke-test evidence.
