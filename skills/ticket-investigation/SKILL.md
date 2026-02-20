# Skill: Ticket Investigation

Run structured ticket analysis before and during implementation.

## Workflow
1. Write JTBD in one sentence.
2. Build acceptance criteria matrix with testable behaviors.
3. Trace expected code path and likely failure points.
4. List verification constraints for current machine/environment.
5. Record findings and recommended action.

## Required ticket notes
For each ticket in `planning/release-r1/tickets.json`, capture:
1. Unknowns found.
2. Cut candidates (if appetite risk appears).
3. Evidence links (tests, logs, endpoint responses).

## Output
1. Update ticket status (`uphill` -> `downhill` -> `done` or `cut`).
2. Record unresolved risks in `planning/release-r1/risk-register.json`.
