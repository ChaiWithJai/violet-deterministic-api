# Skill: ADR Creation

Create architectural decision records that preserve why, not just what.

## When to use
1. New architectural pattern is introduced.
2. A trade-off impacts reliability/scalability/maintainability.
3. A migration boundary or deprecation gate is defined.

## ADR format
Use `docs/adr/ADR-XXXX-title.md`.

Minimum sections:
1. Context
2. Decision
3. Options considered
4. Consequences (positive/negative)
5. Evidence (code/tests/metrics)

## Rules
1. Every P0 ticket that changes architecture should produce or update an ADR.
2. Link ADR IDs in ticket JSON `done_definition` notes.
3. If a decision reverses, mark old ADR as superseded instead of deleting.
