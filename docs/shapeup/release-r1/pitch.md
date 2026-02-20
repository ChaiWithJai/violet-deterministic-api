# Pitch: R1 Prompt-to-Production Foundation

## Problem
People want to prompt an assistant to build a personalized SaaS app, customize it, verify it works, and deploy either self-hosted or managed. Current Violet-era architecture can do many things but is not deterministic enough for safe human+agent operations at scale.

## Appetite
**6 weeks** (2026-02-23 to 2026-04-03), fixed time and variable scope.

## Solution (shaped)
Deliver a deterministic platform slice with these vertical outcomes:
1. Deterministic decision + replay + idempotency API path.
2. Policy/ranking execution seam wired for Gorse + GoRules.
3. App-building control API for minimal viable SaaS blueprint lifecycle.
4. Verification API that returns machine-readable pass/fail evidence.
5. Deployment intent API for self-host vs managed punchout.
6. Agent-compatible orchestration contract (DeepAgents and OpenClaw-like friendly).

## Boundaries
1. One release bet only, no parallel “big rewrite” branches.
2. One canonical API contract version for R1.
3. One deterministic ranking strategy and one policy versioning strategy.

## Rabbit holes anticipated
1. Determinism drift from asynchronous state updates.
2. Gorse score non-determinism from cache/model changes across runs.
3. Rule versioning and rollback strategy across environments.
4. Migration compatibility assumptions from Violet JSONB semantics.
5. Self-host and managed deployment security divergence.

## Rabbit-hole mitigations
1. Snapshot metadata on every response.
2. Replay fixture harness with golden files.
3. Policy artifact registry and version pinning.
4. Import adapter with strict schema translation and rejects.
5. Deployment profiles with explicit security baseline checks.

## No-gos (out of bounds)
1. Full Violet parity in one cycle.
2. Runtime dynamic code execution for custom user logic.
3. Unbounded plugin marketplace in R1.
4. Multi-region production rollout in this bet.

## Expected outcome
By cycle end, teams can run an end-to-end deterministic flow and show credible path to comprehensive SaaS-building capabilities without architectural regressions.
