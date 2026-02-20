# Skill: Debug Tracing

Localize behavioral drift by instrumenting deterministic pipeline boundaries.

## Where to trace first
1. Request normalization and hash generation.
2. Candidate retrieval (Gorse adapter).
3. Policy evaluation (GoRules adapter).
4. Final ranking and tie-breaker stage.
5. Replay retrieval and serialization path.

## Workflow
1. Add temporary structured logs at stage boundaries.
2. Capture trace IDs and compare expected vs actual ordering.
3. Identify first divergence point.
4. Remove or downgrade temporary logs after root cause is validated.

## Output
1. Root-cause note tied to ticket ID.
2. Regression test added when bug is fixed.
