# No-Gos and Risks (R1)

## No-gos
1. No full feature parity with Violet in this cycle.
2. No dynamic user code execution path.
3. No unscoped backlog growth mid-cycle.
4. No production release without replay determinism evidence.

## Top risks
1. Determinism drift under load and asynchronous updates.
2. Policy and recommendation conflicts that produce unstable outputs.
3. Migration edge cases from legacy schema semantics.
4. Security gaps in tenant isolation for agent-operated flows.

## Risk handling
1. Track explicit owners in `planning/release-r1/risk-register.json`.
2. Keep risks visible on board until mitigation validation is complete.
