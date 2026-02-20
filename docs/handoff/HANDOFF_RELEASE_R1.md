# Handoff: Release R1 (Prompt-to-Production SaaS API)

**Repository:** `/Users/jaibhagat/code/violet-deterministic-api`  
**Prepared:** 2026-02-18  
**Handoff owner:** Platform shaping track  
**Incoming team:** API core, platform, agent orchestration, developer experience

## 1) Mission

Ship **R1** as the first externally meaningful release of the deterministic API platform:
1. prompt to build an application,
2. customize safely,
3. verify functionality,
4. deploy via self-host or managed-service punchout,
5. expose APIs suitable for both human operators and AI agents.
6. keep orchestration contracts compatible with OpenClaw-like execution loops while preserving deterministic behavior.

This release prioritizes deterministic behavior, safety boundaries, and operator trust over maximal feature breadth.

## 2) Why this exists

Violet Rails proved product demand and feature breadth, but the runtime is hard to scale safely for AI-operated workflows because of monolithic coupling and dynamic execution patterns.

R1 is the replacement runway: a deterministic, service-oriented API control/data plane that can grow to Violet-level breadth while remaining operationally auditable.

## 3) Current state snapshot

### Done now
1. Deterministic Go scaffold exists with:
- `/v1/health`
- `/v1/decisions`
- `/v1/feedback`
- `/v1/replay`
2. App lifecycle APIs are implemented:
- `/v1/apps`
- `/v1/apps/{id}`
- `/v1/apps/{id}/mutations`
- `/v1/apps/{id}/verify`
- `/v1/apps/{id}/deploy-intents/{target}`
3. Agent orchestration APIs are implemented:
- `/v1/agents/plan`
- `/v1/agents/act`
- `/v1/agents/verify`
- `/v1/agents/deploy`
4. Deterministic ranking tie-breakers are implemented (`score desc`, `item_id asc`).
5. Durable replay/idempotency storage is backed by Postgres and validated across API restart.
6. Auth and tenant claims are enforced on mutating endpoints.
7. Docker demo topology is present (`api`, `postgres`, `redis`, `gorse`).
8. Trial UI is embedded and served at `/ui/` for end-to-end operator testing.
9. Studio generation APIs are available for prompt -> confirmation -> workload/code/terminal preview (`/v1/studio/jobs` family).
10. Multi-model agent runtime APIs are available (`/v1/llm/providers`, `/v1/llm/infer`) with local-first Ollama defaults and OpenAI-compatible frontier adapter.
11. Tool catalog and CLI bridge are available (`/v1/tools`, `cmd/vda`) for API-as-tools and operator scripting.
12. DDIA extraction and deprecation baseline docs are present.

### Not done yet
1. Full Violet migration parity tooling and deprecation cutover gates.
2. Full production-hardening pack (load/SLO suite + threat model signoff).
3. Deeper externalized policy runtime parity beyond seeded local GoRules seam.

## 4) Shape Up framing used for this handoff

This handoff adopts Shape Up principles from `/Users/jaibhagat/Downloads/shape-up.pdf`:
1. **Shaped work** before assignment: solved enough + bounded enough to start.
2. **Fixed time, variable scope**: one six-week cycle, scope hammered to ship.
3. **Bet, not backlog**: this release board is a finite bet with explicit no-gos.
4. **Scope map + hill chart language**: progress measured by unknowns cleared, not task count.
5. **Cooldown planned** for hardening and shaping next bet.

## 5) Release R1 operational plan

### Cycle dates
- **Cycle start:** 2026-02-23
- **Cycle end:** 2026-04-03
- **Cooldown:** 2026-04-06 to 2026-04-17

### Release objective (narrow)
Deliver one end-to-end path: **Prompt -> Model -> Policy -> Build spec -> Verify -> Deploy intent** with deterministic APIs and traceability.

### Release objective (broad)
Establish platform conventions that can scale toward Violet-level comprehensiveness without reintroducing monolithic/runtime hazards.

## 6) Canonical planning artifacts

1. Shape Up docs index: `/Users/jaibhagat/code/violet-deterministic-api/docs/shapeup/release-r1/README.md`
2. Pitch: `/Users/jaibhagat/code/violet-deterministic-api/docs/shapeup/release-r1/pitch.md`
3. Scope map: `/Users/jaibhagat/code/violet-deterministic-api/docs/shapeup/release-r1/scope-map.md`
4. Board JSON: `/Users/jaibhagat/code/violet-deterministic-api/planning/release-r1/board.json`
5. In-depth tickets JSON: `/Users/jaibhagat/code/violet-deterministic-api/planning/release-r1/tickets.json`
6. Milestones: `/Users/jaibhagat/code/violet-deterministic-api/planning/release-r1/milestones.json`
7. Risk register: `/Users/jaibhagat/code/violet-deterministic-api/planning/release-r1/risk-register.json`
8. Fullstack output RFC: `/Users/jaibhagat/code/violet-deterministic-api/docs/rfc/RFC-0001-fullstack-violet-rails-output.md`
9. RFC field report: `/Users/jaibhagat/code/violet-deterministic-api/docs/field-reports/FR-2026-02-19-rfc-0001-ralph-junior.md`

## 7) First 72 hours checklist for the kickoff agent

1. Read pitch + scope map + board JSON in full.
2. Confirm cycle boundaries and freeze no-gos.
3. Start with uphill scopes only:
- deterministic persistence contract,
- auth/tenant boundary,
- Gorse + GoRules deterministic pipeline integration.
4. Post kickoff status using the ticket IDs in board order.
5. Keep release branch strategy simple and traceable.

## 8) Engineering guardrails

1. No `eval` or equivalent dynamic execution in request path.
2. Every mutating endpoint must support idempotency semantics.
3. Every decision response must return `decision_id`, `policy_version`, `data_version`, `generated_at`.
4. Replay must be exact for same decision snapshot.
5. Any nondeterministic dependency must be versioned or isolated behind snapshot metadata.

## 9) Release gate (must-pass)

1. p95 decision latency target met in demo/controlled load profile.
2. Replay determinism mismatch <= 0.5% on fixture suite.
3. Tenant isolation controls validated for read and write paths.
4. Self-host deployment flow documented and reproducible.
5. Managed-service punchout flow contract documented.

## 10) Escalation protocol

1. If scope threatens cycle boundary, cut scope before extending time.
2. If a ticket violates no-gos, split and re-shape it.
3. If deterministic contract is compromised, block merge.
4. If migration parity cannot be proven, hold deprecation steps.

## 11) Notes for next shaper

This release is intentionally foundation-heavy. The win condition is not “all Violet features.” The win condition is a **credible deterministic platform core** that can carry those features safely over subsequent bets.
