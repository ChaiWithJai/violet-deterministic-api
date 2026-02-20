# Parity Delivery Plan (Post-Incident)

Date: 2026-02-20  
Related: `RFC-0002`, incident #13, tickets `R1-027` to `R1-032`

## Current vs Required

| Surface | Current state | Required for "Violet Rails-level working app" | Ticket(s) |
|---|---|---|---|
| Verification truth | Structural checks can pass while runtime can fail | `pass` only when same-revision `run target=all` passes | `R1-027` |
| Runtime entity checks | Hardcoded `account` probe | Probe derives from generated entities | `R1-028` |
| Generated API contract | OpenAPI under-documents runtime routes | Route-complete OpenAPI parity with runtime handlers | `R1-029` |
| Product primitives depth | Seeded/demo read-only seams | Stateful write/read behavior + fixtures | `R1-030` |
| Identity depth | Stub provider/session responses | Behavioral auth/invite/role/subdomain workflows | `R1-030` |
| Artifact/operator trust | Container-local paths can look missing | Explicit retrieval/visibility contract | `R1-031` |
| Parity claim rigor | Narrative and static matrix | Quantified scorecard and benchmark gate | `R1-032` |

## Execution Tree

1. **Containment (must complete first)**
   - `R1-028`: remove hardcoded entity assumptions.
   - `R1-027`: make verification verdict runtime-backed.

2. **Contract hardening (second gate)**
   - `R1-029`: route-complete OpenAPI + parity assertions.

3. **Depth uplift (third gate, cuttable if needed)**
   - `R1-030`: stateful primitives/identity modules.
   - `R1-031`: artifact visibility and retrieval clarity.

4. **Release claim gate (final)**
   - `R1-032`: benchmark scorecard in release packet.

## Cut Policy (Fixed Time, Variable Scope)

1. `R1-027`, `R1-028`, `R1-029` are P0 and non-cuttable for R1 parity claims.
2. `R1-030`, `R1-031`, `R1-032` are P1 and can be cut only if explicitly recorded in board + RFC notes.
3. No release-candidate claim can bypass incident #13 closure criteria.
