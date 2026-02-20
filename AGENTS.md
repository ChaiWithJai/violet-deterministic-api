## Skills
A skill is a set of local instructions stored in a `SKILL.md` file. Use these skills for repeatable execution in this repository.

### Available skills
- codebase-archaeology: Analyze git history, hotspots, ownership, and architectural evolution before making deep changes. (file: /Users/jaibhagat/code/violet-deterministic-api/skills/codebase-archaeology/SKILL.md)
- adr-creation: Create and update ADRs for decisions, trade-offs, and discoveries. (file: /Users/jaibhagat/code/violet-deterministic-api/skills/adr-creation/SKILL.md)
- ticket-investigation: Run structured issue investigation with JTBD, acceptance criteria, and evidence. (file: /Users/jaibhagat/code/violet-deterministic-api/skills/ticket-investigation/SKILL.md)
- platform-verification: Verify behavior on the current machine and record platform constraints honestly. (file: /Users/jaibhagat/code/violet-deterministic-api/skills/platform-verification/SKILL.md)
- debug-tracing: Add and remove targeted tracing to localize behavioral drift quickly. (file: /Users/jaibhagat/code/violet-deterministic-api/skills/debug-tracing/SKILL.md)
- ddia-api-first-audit: Perform DDIA reliability/scalability/maintainability audits for this deterministic API architecture. (file: /Users/jaibhagat/code/violet-deterministic-api/skills/ddia-api-first-audit/SKILL.md)
- shapeup-release-r1: Operate the R1 Shape Up artifacts, board, and cuts using fixed-time variable-scope discipline. (file: /Users/jaibhagat/code/violet-deterministic-api/skills/shapeup-release-r1/SKILL.md)
- deterministic-go-api-implementation: Implement deterministic APIs in Go with gorse/gorules/go-pipeline seams and replay-safe contracts. (file: /Users/jaibhagat/code/violet-deterministic-api/skills/deterministic-go-api-implementation/SKILL.md)
- agent-orchestration: Define human+AI orchestration contracts (plan/act/verify/deploy) and guardrails for deterministic automation. (file: /Users/jaibhagat/code/violet-deterministic-api/skills/agent-orchestration/SKILL.md)
- violet-migration-parity: Plan and execute migration parity checks from Violet Rails to this service. (file: /Users/jaibhagat/code/violet-deterministic-api/skills/violet-migration-parity/SKILL.md)

### How to use skills
- Trigger rules: Use a skill when the user names it or when the task clearly matches its description.
- Progressive disclosure: Open only the `SKILL.md` needed for the current task and load referenced files only when necessary.
- Coordination: If multiple skills apply, use the minimal set and state the execution order.
- Fallback: If a skill cannot be applied cleanly, state why and proceed with best-effort execution.

### Orchestration
See `/Users/jaibhagat/code/violet-deterministic-api/skills/ORCHESTRATION.md` for the canonical sequence used for R1 delivery.
