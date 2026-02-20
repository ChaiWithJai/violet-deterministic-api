# Skill: Agent Orchestration

Define deterministic orchestration contracts for human and AI agents.

## Goal
Support prompt-to-production workflows where agents can plan, act, verify, and request deployment intent while humans retain governance controls.

## Contract model
1. Plan API: propose app blueprint and mutation plan.
2. Act API: apply bounded, policy-checked mutations.
3. Verify API: return machine-readable pass/fail evidence.
4. Deploy API: emit self-host or managed-service punchout intent.

## Guardrails
1. Human approval hooks for privileged actions.
2. Idempotency and replay on mutation actions.
3. Tenant and policy boundaries on every action.
4. Decision traceability for post-incident review.

## Integration target
This contract should be consumable by LangChain DeepAgents style orchestrators and OpenClaw-like execution loops without introducing nondeterministic side effects.

## Output
1. Contract examples in docs and OpenAPI schema.
2. Security and governance requirements mapped to release tickets.
