# Agent Integration Plan (DeepAgents / OpenClaw-style)

## Objective
Expose deterministic building blocks that AI agents can orchestrate while humans retain control.

## API capabilities for agents
1. Plan: `POST /v1/agents/plan` proposes deterministic blueprint with checks.
2. Act: `POST /v1/agents/act` applies one policy-checked mutation.
3. Verify: `POST /v1/agents/verify` returns machine-readable verdict/check evidence.
4. Deploy: `POST /v1/agents/deploy` creates self-host or managed deploy intent.
5. Provider inventory: `GET /v1/llm/providers` lists local/frontier model availability.
6. Model call: `POST /v1/llm/infer` runs one provider-agnostic inference step.
7. Tool catalog: `GET /v1/tools` exposes API endpoints as tool descriptors + CLI mappings.

## Guardrails
1. Every mutating action requires idempotency key.
2. Plan/act/verify/deploy responses include versioned metadata for deterministic replay and audit (`policy_version`, `data_version`, or immutable intent/report ids).
3. Human override endpoint for approval gates.

## Current status
1. Endpoints are implemented and documented in `api/openapi.yaml`.
2. Local smoke results are tracked in `docs/runbooks/local-smoke.md`.
3. Inverted dependency for model adapters is captured in `docs/adr/ADR-0002-multimodel-agent-runtime.md`.
