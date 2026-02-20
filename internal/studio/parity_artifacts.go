package studio

import (
	"fmt"
	"strings"
)

func paritySupportArtifacts(slug string, conf Confirmation) []FileArtifact {
	base := fmt.Sprintf("apps/%s", slug)
	depth := normalizeDepthLabel(conf.GenerationDepth)
	entities := withDefault(conf.DataEntities, "account")
	workflows := withDefault(conf.CoreWorkflows, "approve_request")
	users := withDefault(conf.PrimaryUsers, "admin")

	behaviorScenarios := fmt.Sprintf(`depth_label: %s
categories:
  unit:
    - validate generated runtime handlers compile and route correctly
  integration:
    - execute /v1/entities/{entity}/records create/list flow
    - execute /v1/actions/execute for declared workflows
  e2e:
    - validate cms/blog/forum/email primitive endpoints return seeded data
    - validate identity register/login/invite/subdomain stubs
  accessibility:
    - verify generated web modules expose semantic labels in view models
declared_entities:
%s
declared_workflows:
%s
`, depth, toYAMLList(entities), toYAMLList(workflows))

	parityMatrix := fmt.Sprintf(`# Generated API Endpoint Matrix

This file inventories what the generated app runtime provides today versus the Violet Rails baseline.

| Surface | Generated VDA Runtime | Violet Rails baseline | Current status |
|---|---|---|---|
| Entity CRUD | /v1/entities/{entity}/records (GET/POST) | Dynamic namespace/resource CRUD | Partial parity: generated runtime provides app-local CRUD lane |
| Actions | /v1/actions/execute (POST) | api/:version/:namespace/:resource/:id/:action | Partial parity: deterministic action execution in app runtime |
| Product primitives | /v1/primitives/cms/pages, /blog/posts, /forum/threads, /email/messages | Built-in CMS/blog/forum/mailbox | Partial parity: seeded primitives generated as runtime modules |
| User lifecycle | /v1/identity/register, /login, /invitations, /roles, /subdomains/claim | Devise + invites + OTP + admin/subdomain governance | Partial parity: generated identity module + provider stubs |
| Control plane mutations | POST /v1/apps/{id}/mutations (4 classes) | Dynamic control-plane resources/forms/actions | Intentionally constrained: rich operations live in generated runtime |

## Declared Context

- Domain: %s
- Depth label: %s
- Primary users: %s
- Data entities: %s
- Workflows: %s
`, fallback(conf.Domain, "saas"), depth, strings.Join(users, ", "), strings.Join(entities, ", "), strings.Join(workflows, ", "))

	boundaryNotes := fmt.Sprintf(`# Control Plane vs Generated Runtime Boundary

The deterministic control plane remains intentionally constrained for replay-safe operations.

## Control plane responsibilities

1. Tenant-scoped auth, idempotency, replay safety.
2. Blueprint lifecycle and constrained mutation classes.
3. Studio generation orchestration.

## Generated runtime responsibilities

1. Entity CRUD for declared data entities (%s).
2. Action execution for declared workflows (%s).
3. Product primitives (CMS/blog/forum/email) as generated modules.
4. End-user identity flows (register/login/invite/roles/subdomain claim).

## Why this split exists

Violet Rails exposed a highly dynamic runtime inside one monolith. VDA intentionally keeps the control plane deterministic and relocates product-specific behavior into generated app runtime artifacts.
`, strings.Join(entities, ", "), strings.Join(workflows, ", "))

	migrationGuide := `# Migration Guide: Content, Community, and Email

This generated app includes primitive starter surfaces for CMS, blog, forum, and email.

## What is generated now

1. CMS pages endpoint with seeded records.
2. Blog posts endpoint with seeded records.
3. Forum threads endpoint with seeded records.
4. Email messages endpoint with seeded records.

## Workaround path for production migration

1. Keep generated endpoints as deterministic seams.
2. Connect production-grade providers behind integration adapters.
3. Preserve route contracts while replacing seeded handlers with domain logic.
`

	webCMSModule := `export type CMSPage = {
  slug: string;
  title: string;
  body: string;
};

export const CMS_PAGES: CMSPage[] = [
  { slug: "home", title: "Home", body: "Generated CMS placeholder content." },
];
`

	webBlogModule := `export type BlogPost = {
  slug: string;
  title: string;
  excerpt: string;
};

export const BLOG_POSTS: BlogPost[] = [
  { slug: "hello-world", title: "Hello World", excerpt: "Generated blog starter post." },
];
`

	webForumModule := `export type ForumThread = {
  id: string;
  title: string;
  author: string;
};

export const FORUM_THREADS: ForumThread[] = [
  { id: "thread-1", title: "Welcome", author: "system" },
];
`

	webEmailModule := `export type EmailMessage = {
  id: string;
  subject: string;
  status: "queued" | "sent";
};

export const EMAIL_MESSAGES: EmailMessage[] = [
  { id: "email-1", subject: "Welcome to Violet", status: "queued" },
];
`

	webAuthModule := fmt.Sprintf(`export type GeneratedRole = string;

export const GENERATED_ROLES: GeneratedRole[] = %s;
export const AUTH_PROVIDER_STUBS = ["auth0", "clerk", "supabase"] as const;
export const AUTH_BOUNDARY = "control_plane_tokens_are_separate_from_generated_app_sessions";
`, mustJSONString(users))

	rbacModel := fmt.Sprintf(`{
  "roles": %s,
  "constraints": %s,
  "auth_providers": ["auth0", "clerk", "supabase"],
  "separation": "control_plane_vs_generated_runtime"
}
`, mustJSONString(users), mustJSONString(withDefault(conf.Constraints, "all_mutations_idempotent")))

	return []FileArtifact{
		{Path: base + "/tests/behavior/scenarios.yaml", Language: "yaml", Content: behaviorScenarios},
		{Path: base + "/docs/parity/api-endpoint-matrix.md", Language: "markdown", Content: parityMatrix},
		{Path: base + "/docs/parity/control-plane-vs-runtime.md", Language: "markdown", Content: boundaryNotes},
		{Path: base + "/docs/parity/migration-guide-content-community-email.md", Language: "markdown", Content: migrationGuide},
		{Path: base + "/clients/web/modules/cms.ts", Language: "typescript", Content: webCMSModule},
		{Path: base + "/clients/web/modules/blog.ts", Language: "typescript", Content: webBlogModule},
		{Path: base + "/clients/web/modules/forum.ts", Language: "typescript", Content: webForumModule},
		{Path: base + "/clients/web/modules/email.ts", Language: "typescript", Content: webEmailModule},
		{Path: base + "/clients/web/modules/auth.ts", Language: "typescript", Content: webAuthModule},
		{Path: base + "/config/rbac.generated.json", Language: "json", Content: rbacModel},
	}
}
