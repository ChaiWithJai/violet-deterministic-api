# Multi-Model Coding Agent Runtime (LangGraph + Ollama + Frontier)

## Objective
Run coding agents with local-first models while preserving deterministic API contracts and easy provider switching.

## Architecture

### 1) Inverted dependency boundary
1. API handlers call `internal/llm` service.
2. `internal/llm` selects provider adapters (`ollama`, `frontier`) at runtime.
3. Agent orchestration contracts (`/v1/agents/*`) stay stable regardless of model backend.

### 2) Provider roles
1. `ollama` for local fast feedback and offline-ish loops.
2. `frontier` for high-capability escalations via OpenAI-compatible endpoint.
3. Routing is request-level (`provider`, `model`) with defaults from env.

### 3) Tool interface
1. `/v1/tools` lists tool descriptors for agent runtimes.
2. `/v1/agents/plan`, `/v1/agents/act`, `/v1/agents/verify`, `/v1/agents/deploy` are core orchestration tools.
3. `/v1/llm/providers` and `/v1/llm/infer` expose provider inventory and one-call inference for runtime control/testing.

## API interfaces
1. `GET /v1/llm/providers`
2. `POST /v1/llm/infer`
3. `GET /v1/tools`

## CLI interfaces
1. `vda tools list`
2. `vda llm providers`
3. `vda llm infer --provider ollama --model glm-4.7-flash:latest --prompt "..."`

## LangGraph wiring pattern

Use LangGraph graph nodes with HTTP tool calls to this API:
1. `plan` node -> call `/v1/agents/plan`
2. `act` node -> call `/v1/agents/act`
3. `verify` node -> call `/v1/agents/verify`
4. `deploy` node -> call `/v1/agents/deploy`
5. optional `model_router` node -> call `/v1/llm/infer` for non-mutating synthesis

This keeps mutation and deploy paths governed by deterministic API guardrails, not raw model side effects.

## Local test setup

1. Start Ollama locally and pull target model:
- `ollama pull glm-4.7-flash`

2. Run API stack:
- `docker compose -f docker-compose.demo.yml up -d --build`

3. Verify provider health:
- `curl -s http://localhost:4020/v1/llm/providers -H "Authorization: Bearer dev-token" | jq`

4. Run one inference call:
- `curl -s -X POST http://localhost:4020/v1/llm/infer -H "Authorization: Bearer dev-token" -H "Idempotency-Key: infer-1" -H "Content-Type: application/json" -d '{"provider":"ollama","model":"glm-4.7-flash:latest","prompt":"Return only: ok"}' | jq`

5. Run OpenAI-compatible lane (frontier adapter, local by default in demo compose):
- `curl -s -X POST http://localhost:4020/v1/llm/infer -H "Authorization: Bearer dev-token" -H "Idempotency-Key: infer-2" -H "Content-Type: application/json" -d '{"provider":"frontier","model":"glm-4.7-flash:latest","prompt":"Return only: ok"}' | jq`

## Model strategy from day one

1. Local baseline lane:
- Deterministic prompts on local models (Ollama) for fast loop testing.

2. Frontier lane:
- Same API contract, provider switched to `frontier` for complex tasks.

3. Policy lane:
- Mutating operations continue through `/v1/agents/*` with policy checks and idempotency.

## Recommended rollout

1. Week 1: local-only coding loops with `/v1/llm/infer` and `/v1/tools`.
2. Week 2: LangGraph node graph against `/v1/agents/*` + `/v1/llm/*`.
3. Week 3: add provider routing policy (task class -> model tier), then enable frontier fallback.
4. Week 4+: add offline eval harness and regression suite per model/provider combination.
