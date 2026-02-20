# ADR-0001: Deterministic API-First Core

**Date:** 2026-02-18
**Status:** Proposed

## Context
Violet Rails runtime combines CMS/admin/app concerns with dynamic execution paths, reducing determinism and increasing migration risk.

## Decision
Create a standalone Go service as the deterministic API core with explicit rule/ranking adapters and replay/idempotency guarantees.

## Consequences
1. Positive: clear service boundaries, safer agent automation, independent scaling.
2. Negative: dual-run migration complexity until full cutover.
