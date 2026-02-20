# Gas-Town Issue Execution Plan (Manual Fallback)

Generated: 2026-02-20T03:16:30Z

## Context
- `scripts/collect_open_issues.sh` and `scripts/build_issue_execution_plan.py` are not present in this repo.
- Plan generated manually from live GitHub issues + `planning/release-r1/tickets.json` dependencies.

## Sequence
1. **#6 / R1-024 (P0, uphill)**
   - Blocker: generated-app depth remains scaffold-level.
   - Why first: directly impacts product credibility and release gate quality.
2. **#8 / R1-026 (P0, uphill)**
   - Blocker: migration export/import endpoints missing.
   - Why second: critical deprecation and cutover gate.
3. **#4 / R1-022 (P1, bet_accepted)**
   - Product primitives strategy (generate vs native vs defer).
4. **#5 / R1-023 (P1, bet_accepted)**
   - Mutation paradigm boundary and schema-driven expansion feasibility.
5. **#7 / R1-025 (P1, bet_accepted)**
   - Generated app auth/governance depth.

## Cut Policy
- Keep P0 items mandatory for release progression.
- P1 items remain bet_accepted with explicit cut option if P0 unknowns exceed appetite.

## PR Swarm Gate Snapshot
- Open PR: `#3` (`fix/1-close-template-parity-contract-drift`)
- Comments/reviews: none unresolved.
- Checks: none reported.
- Merge gate result: **not ready** until checks/evidence are attached.
