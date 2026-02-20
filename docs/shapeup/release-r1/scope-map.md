# Scope Map (R1)

## Scope S1: Deterministic Core API
- Decision, replay, and idempotency persistence.
- Canonical hashing + snapshot metadata.

## Scope S2: Policy + Ranking Runtime
- Gorse retrieval/ranking integration seam.
- GoRules policy evaluation seam.
- Deterministic pipeline ordering.

## Scope S3: App Builder Control APIs
- Create/update app blueprint entities.
- Customize workflow constraints safely.

## Scope S4: Verify APIs and Test Harness
- Act/verify contracts for human+agent use.
- Replay, fixture, and compatibility tests.

## Scope S5: Deployment Punchout
- Self-host deployment intent.
- Managed-service request/approval intent.

## Scope S6: Migration and Deprecation Readiness
- Violet export/import translator.
- Parity checks and cutover readiness criteria.
