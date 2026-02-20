# API Endpoint Matrix: VDA vs Violet Rails

Updated: 2026-02-20

| Surface | Violet Deterministic API (current) | Violet Rails baseline | Gap status |
|---|---|---|---|
| Decisions + replay | `POST /v1/decisions`, `POST /v1/replay`, `POST /v1/feedback` | Dynamic app/runtime logic embedded in monolith | Different paradigm (deterministic control plane) |
| App model control plane | `POST/GET/PATCH /v1/apps`, `POST /v1/apps/{id}/mutations`, `POST /v1/apps/{id}/verify` | Dynamic namespace/resource/form/action routing | Control-plane remains constrained by design |
| Generated runtime entity CRUD | Generated app: `GET/POST /v1/entities/{entity}/records` | Dynamic resources/actions on main runtime | Added in generated runtime (partial parity) |
| Generated runtime action execution | Generated app: `POST /v1/actions/execute` | `.../:api_action` route model | Added in generated runtime (partial parity) |
| Product primitives | Generated app: `/v1/primitives/cms/pages`, `/blog/posts`, `/forum/threads`, `/email/messages` | Built-in CMS/blog/forum/mailbox | Added as generated primitives (partial parity) |
| User lifecycle + governance | Generated app: `/v1/identity/register`, `/login`, `/invitations`, `/roles`, `/subdomains/claim`, `/providers` | Devise, invites, OTP, admin, subdomain governance | Added as generated identity module + provider stubs (partial parity) |
| Studio generation | `POST /v1/studio/jobs`, artifacts/run/verification/jtbd/preview/terminal/bundle | No equivalent | VDA advantage |
| LLM orchestration | `/v1/llm/providers`, `/v1/llm/infer`, agent plan/clarify/act/verify/deploy | No equivalent | VDA advantage |
| Migration parity | `POST /v1/migration/violet/export`, `POST /v1/migration/violet/import` | Legacy runtime migration scripts | Implemented in control plane |

## Inventory: what VDA produces now

1. Deterministic control-plane APIs for app lifecycle, replay, policy, and agent orchestration.
2. Generated fullstack artifacts with runtime web/mobile clients and generated backend service.
3. Generated runtime seams for dynamic entity CRUD/actions.
4. Generated runtime primitives for CMS/blog/forum/email.
5. Generated runtime identity module + Auth0/Clerk/Supabase adapter stubs.
6. Behavioral fixtures and run-target verification with depth labeling.

## Inventory: what VDA still does not produce natively in control plane

1. Full dynamic namespace/resource runtime (Violet-style arbitrary control-plane CRUD).
2. Production-grade built-in CMS/forum/email internals with persistence/search/moderation.
3. Full human-facing auth UX stack (MFA flows, password reset UX, admin console UI).

## Decision tree

1. Keep deterministic control plane constrained.
2. Push rich product behavior into generated runtime artifacts.
3. Escalate only repeated runtime gaps back into control-plane APIs when determinism can be preserved.
