# ADR-0002: Multi-Model Agent Runtime (Local-First Inverted Dependency)

**Date:** 2026-02-19
**Status:** Proposed

## Context
The release target requires deeper app generation, local testing loops, and model portability across local and frontier providers.

We need:
1. Coding-agent execution that can run locally with Ollama for fast iteration.
2. An architecture that can switch providers without rewriting agent orchestration.
3. A tool interface consumable by API clients, LangGraph orchestrators, and CLI operators.

## Decision
Adopt an inverted-dependency model runtime where API handlers depend on a provider-agnostic LLM service (`internal/llm`) instead of provider-specific clients.

Implemented provider classes in this slice:
1. `ollama` (local, default)
2. `frontier` (OpenAI-compatible remote endpoint)

New control interfaces:
1. `GET /v1/llm/providers` for inventory/health/model availability.
2. `POST /v1/llm/infer` for one idempotent model call.
3. `GET /v1/tools` for API-as-tools discovery and CLI mappings.
4. `cmd/vda` CLI for `llm providers`, `llm infer`, and `tools list`.

## Options considered
1. Hard-code a single provider (Ollama only).
- Rejected: blocks frontier evaluation and migration portability.

2. Build agent orchestration directly inside LangGraph before API contracts.
- Rejected: weakens deterministic governance and API-first guarantees.

3. Invert dependency with provider adapter boundary (selected).
- Accepted: keeps orchestration stable while adding/removing model backends.

## Consequences
1. Positive: local-first iteration with consistent API contracts.
2. Positive: multi-model from day one without endpoint churn.
3. Positive: clean path for LangGraph/OpenClaw style tool execution.
4. Negative: provider heterogeneity requires strict normalization and error mapping.
5. Negative: runtime generation quality still depends on model capability and prompt discipline.

## Evidence
1. Code: `internal/llm/service.go`, `internal/http/llm_handlers.go`, `internal/http/tools_handlers.go`, `cmd/vda/main.go`.
2. Contracts: `api/openapi.yaml` includes `/v1/llm/providers`, `/v1/llm/infer`, `/v1/tools`.
3. Tests: `internal/llm/service_test.go` and existing suite pass.
4. Live checks: local runbook updated with provider and runtime preview verification.
