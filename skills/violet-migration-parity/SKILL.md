# Skill: Violet Migration Parity

Execute migration from Violet Rails with measurable parity and rollback-safe cutover.

## Inputs
1. Legacy data/export structures from Violet.
2. `docs/EXTRACTION_ROADMAP.md`
3. `docs/DEPRECATION_PLAN.md`

## Workflow
1. Define canonical translation rules from Violet models to new API models.
2. Build fixture corpus for representative tenant/application states.
3. Compare old vs new behavior with replay and verification endpoints.
4. Track parity mismatches and classify as acceptable/unacceptable.
5. Gate cutover on SLO + parity thresholds.

## Cutover gates
1. Replay mismatch <= target threshold.
2. Auth/tenant isolation fully enforced.
3. Rollback artifact and read-only window prepared.

## Output
1. Migration report with mismatch taxonomy.
2. Cutover recommendation and rollback plan.
