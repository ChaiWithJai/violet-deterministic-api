package studio

import (
	"fmt"
	"math"
	"path/filepath"
	"strings"
	"time"
)

type ArtifactManifest struct {
	GeneratedAt   time.Time      `json:"generated_at"`
	WorkspacePath string         `json:"workspace_path"`
	Files         []ArtifactFile `json:"files"`
	RunTargets    []RunTarget    `json:"run_targets"`
}

type ArtifactFile struct {
	Path      string `json:"path"`
	Language  string `json:"language"`
	Category  string `json:"category"`
	SizeBytes int    `json:"size_bytes"`
}

type RunTarget struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Command     string `json:"command"`
}

type VerificationReport struct {
	ReportID           string              `json:"report_id"`
	Verdict            string              `json:"verdict"`
	DepthLabel         string              `json:"depth_label"`
	BehavioralPassRate float64             `json:"behavioral_pass_rate"`
	Checks             []VerificationCheck `json:"checks"`
	GeneratedAt        time.Time           `json:"generated_at"`
}

type VerificationCheck struct {
	ID       string `json:"id"`
	Status   string `json:"status"`
	Evidence string `json:"evidence"`
}

type JTBDCoverage struct {
	ID       string `json:"id"`
	Task     string `json:"task"`
	Status   string `json:"status"`
	Evidence string `json:"evidence"`
}

type RunResult struct {
	Target      string              `json:"target"`
	Status      string              `json:"status"`
	Checks      []VerificationCheck `json:"checks"`
	GeneratedAt time.Time           `json:"generated_at"`
}

func buildArtifactManifest(job Job) ArtifactManifest {
	files := make([]ArtifactFile, 0, len(job.Files))
	for _, f := range job.Files {
		files = append(files, ArtifactFile{
			Path:      f.Path,
			Language:  f.Language,
			Category:  categorizeArtifact(f.Path),
			SizeBytes: len(f.Content),
		})
	}
	return ArtifactManifest{
		GeneratedAt:   time.Now().UTC(),
		WorkspacePath: job.WorkspacePath,
		Files:         files,
		RunTargets: []RunTarget{
			{Name: "web", Description: "Validate web runtime artifacts", Command: "studio run web"},
			{Name: "mobile", Description: "Validate mobile runtime artifacts", Command: "studio run mobile"},
			{Name: "api", Description: "Validate backend contract artifacts", Command: "studio run api"},
			{Name: "verify", Description: "Run deterministic verification checks", Command: "studio run verify"},
			{Name: "all", Description: "Run all checks", Command: "studio run all"},
		},
	}
}

func categorizeArtifact(path string) string {
	normalized := filepath.ToSlash(strings.ToLower(strings.TrimSpace(path)))
	switch {
	case strings.Contains(normalized, "/clients/web/modules/"):
		return "web_module"
	case strings.Contains(normalized, "/clients/web/"):
		return "web_client"
	case strings.Contains(normalized, "/clients/mobile/"):
		return "mobile_client"
	case strings.Contains(normalized, "/internal/identity/"):
		return "identity_module"
	case strings.Contains(normalized, "/internal/primitives/"):
		return "product_primitive"
	case strings.Contains(normalized, "/config/rbac"):
		return "auth_model"
	case strings.Contains(normalized, "/docs/parity/"):
		return "parity_doc"
	case strings.Contains(normalized, "/internal/integrations/"):
		return "integration_adapter"
	case strings.Contains(normalized, "/services/api/"):
		return "backend_service"
	case strings.Contains(normalized, "/api/"):
		return "api_contract"
	case strings.Contains(normalized, "/workflows/"):
		return "workflow"
	case strings.Contains(normalized, "/tests/"):
		return "test"
	case strings.Contains(normalized, "/deploy/"):
		return "deploy"
	case strings.Contains(normalized, "/boilerplate/"):
		return "migration_note"
	case strings.HasSuffix(normalized, "/readme.md"):
		return "docs"
	default:
		return "misc"
	}
}

func buildVerificationReport(job Job) VerificationReport {
	checks := []VerificationCheck{
		{
			ID:       "artifacts_required_present",
			Status:   passFailStatus(hasPaths(job.Files, "/clients/web/index.html", "/clients/web/app.js", "/clients/mobile/index.html", "/clients/mobile/app.js", "/api/openapi.yaml")),
			Evidence: "required runtime and api contract files",
		},
		{
			ID:       "tests_present",
			Status:   passFailStatus(hasCategory(job.ArtifactManifest.Files, "test")),
			Evidence: "at least one test artifact generated",
		},
		{
			ID:       "deploy_artifacts_present",
			Status:   passFailStatus(hasCategory(job.ArtifactManifest.Files, "deploy")),
			Evidence: "self-host and managed deploy artifacts present",
		},
		{
			ID:       "backend_runtime_present",
			Status:   passFailStatus(hasPaths(job.Files, "/services/api/go.mod", "/services/api/cmd/server/main.go", "/services/api/internal/runtime/server.go", "/services/api/Dockerfile")),
			Evidence: "generated backend runtime scaffold present",
		},
		{
			ID:       "agent_tools_contract_present",
			Status:   passFailStatus(hasPaths(job.Files, "/services/api/internal/tools/catalog.go", "/services/api/internal/tools/contracts.ts")),
			Evidence: "generated backend tool contracts present",
		},
		{
			ID:       "policy_constraints_present",
			Status:   passFailStatus(hasConstraint(job.Confirmation.Constraints, "all_mutations_idempotent")),
			Evidence: "idempotency constraint captured",
		},
		{
			ID:       "depth_label_declared",
			Status:   passFailStatus(isDepthLabel(normalizeDepthLabel(job.DepthLabel))),
			Evidence: "studio job depth label is one of prototype/pilot/production-candidate",
		},
		{
			ID:       "behavioral_fixtures_present",
			Status:   passFailStatus(hasPaths(job.Files, "/tests/behavior/scenarios.yaml", "/services/api/tests/behavior.sh")),
			Evidence: "behavioral fixture definitions generated for app + api runtime",
		},
		{
			ID:       "behavioral_runtime_modules_present",
			Status:   passFailStatus(hasPaths(job.Files, "/services/api/internal/runtime/entity_actions.go", "/services/api/internal/runtime/behavior_test.go")),
			Evidence: "generated api runtime includes entity/action handlers and behavioral tests",
		},
		{
			ID:       "behavioral_primitives_modules_present",
			Status:   passFailStatus(hasPaths(job.Files, "/services/api/internal/primitives/module.go", "/clients/web/modules/cms.ts", "/clients/web/modules/blog.ts", "/clients/web/modules/forum.ts", "/clients/web/modules/email.ts")),
			Evidence: "generated CMS/blog/forum/email modules present in runtime and web artifacts",
		},
		{
			ID:       "behavioral_identity_modules_present",
			Status:   passFailStatus(hasPaths(job.Files, "/services/api/internal/identity/module.go", "/services/api/internal/identity/providers/auth0_adapter.go", "/services/api/internal/identity/providers/clerk_adapter.go", "/services/api/internal/identity/providers/supabase_adapter.go", "/clients/web/modules/auth.ts", "/config/rbac.generated.json")),
			Evidence: "generated identity lifecycle module, provider stubs, and RBAC model present",
		},
		{
			ID:       "boundary_docs_present",
			Status:   passFailStatus(hasPaths(job.Files, "/docs/parity/api-endpoint-matrix.md", "/docs/parity/control-plane-vs-runtime.md", "/docs/parity/migration-guide-content-community-email.md")),
			Evidence: "generated docs inventory runtime parity and migration boundaries",
		},
	}

	verdict := "pass"
	for _, check := range checks {
		if check.Status == "fail" {
			verdict = "fail"
			break
		}
	}
	return VerificationReport{
		ReportID:           makeID("studio_vrf", job.TenantID, job.JobID, fmt.Sprintf("%d", job.UpdatedAt.UnixNano())),
		Verdict:            verdict,
		DepthLabel:         normalizeDepthLabel(job.DepthLabel),
		BehavioralPassRate: behavioralPassRate(checks),
		Checks:             checks,
		GeneratedAt:        time.Now().UTC(),
	}
}

func buildJTBDCoverage(job Job) []JTBDCoverage {
	out := []JTBDCoverage{
		{
			ID:       "jtbd_create_app",
			Task:     "Create app from prompt",
			Status:   passFailStatus(len(job.Files) > 0 && len(job.Workload) > 0),
			Evidence: fmt.Sprintf("files=%d workload=%d", len(job.Files), len(job.Workload)),
		},
		{
			ID:       "jtbd_customize_safely",
			Task:     "Customize safely",
			Status:   passFailStatus(hasConstraint(job.Confirmation.Constraints, "all_mutations_idempotent")),
			Evidence: "constraint all_mutations_idempotent present",
		},
		{
			ID:       "jtbd_validate_behavior",
			Task:     "Validate behavior before deploy",
			Status:   passFailStatus(job.Verification.Verdict == "pass" && job.Verification.BehavioralPassRate >= 1),
			Evidence: fmt.Sprintf("verification verdict=%s behavioral_pass_rate=%.2f", job.Verification.Verdict, job.Verification.BehavioralPassRate),
		},
		{
			ID:       "jtbd_operate_human_ai",
			Task:     "Operate with human and AI agents",
			Status:   passFailStatus(hasPaths(job.Files, "/src/agent_contract.ts", "/services/api/internal/tools/catalog.go")),
			Evidence: "agent contract and backend tools catalog generated",
		},
		{
			ID:       "jtbd_backend_runtime",
			Task:     "Run generated backend service",
			Status:   passFailStatus(hasPaths(job.Files, "/services/api/go.mod", "/services/api/cmd/server/main.go", "/services/api/internal/runtime/server.go", "/services/api/internal/runtime/entity_actions.go")),
			Evidence: "backend runtime scaffold + behavioral entity/action handlers generated",
		},
		{
			ID:       "jtbd_product_primitives",
			Task:     "Deliver product primitives in generated runtime",
			Status:   passFailStatus(hasPaths(job.Files, "/services/api/internal/primitives/module.go", "/clients/web/modules/cms.ts", "/clients/web/modules/blog.ts", "/clients/web/modules/forum.ts", "/clients/web/modules/email.ts")),
			Evidence: "generated primitives modules for cms/blog/forum/email",
		},
		{
			ID:       "jtbd_user_lifecycle_governance",
			Task:     "Deliver generated user lifecycle and governance seams",
			Status:   passFailStatus(hasPaths(job.Files, "/services/api/internal/identity/module.go", "/config/rbac.generated.json", "/clients/web/modules/auth.ts")),
			Evidence: "generated identity routes, RBAC model, and web auth module",
		},
		{
			ID:       "jtbd_ship",
			Task:     "Ship self-host or managed",
			Status:   passFailStatus(hasCategory(job.ArtifactManifest.Files, "deploy")),
			Evidence: "deploy artifacts generated",
		},
	}
	return out
}

func runTargetChecks(job Job, target string) []VerificationCheck {
	switch strings.ToLower(strings.TrimSpace(target)) {
	case "web":
		return []VerificationCheck{{
			ID:       "web_runtime",
			Status:   passFailStatus(hasPaths(job.Files, "/clients/web/index.html", "/clients/web/app.js", "/clients/web/app.css")),
			Evidence: "web runtime assets present",
		}}
	case "mobile":
		return []VerificationCheck{{
			ID:       "mobile_runtime",
			Status:   passFailStatus(hasPaths(job.Files, "/clients/mobile/index.html", "/clients/mobile/app.js", "/clients/mobile/app.css")),
			Evidence: "mobile runtime assets present",
		}}
	case "api":
		return []VerificationCheck{
			{ID: "api_openapi", Status: passFailStatus(hasPaths(job.Files, "/api/openapi.yaml")), Evidence: "openapi generated"},
			{ID: "api_agent_contract", Status: passFailStatus(hasPaths(job.Files, "/src/agent_contract.ts")), Evidence: "agent contract generated"},
			{ID: "api_service_runtime", Status: passFailStatus(hasPaths(job.Files, "/services/api/go.mod", "/services/api/cmd/server/main.go", "/services/api/internal/runtime/server.go")), Evidence: "backend runtime scaffold generated"},
			{ID: "api_service_tests", Status: passFailStatus(hasPaths(job.Files, "/services/api/internal/runtime/server_test.go", "/services/api/tests/smoke.sh")), Evidence: "backend runtime tests generated"},
			{ID: "api_dynamic_entity_runtime", Status: passFailStatus(hasPaths(job.Files, "/services/api/internal/runtime/entity_actions.go")), Evidence: "generated runtime entity CRUD and action handlers present"},
			{ID: "api_primitives_modules", Status: passFailStatus(hasPaths(job.Files, "/services/api/internal/primitives/module.go")), Evidence: "generated primitives runtime module present"},
			{ID: "api_identity_modules", Status: passFailStatus(hasPaths(job.Files, "/services/api/internal/identity/module.go", "/services/api/internal/identity/providers/auth0_adapter.go", "/services/api/internal/identity/providers/clerk_adapter.go", "/services/api/internal/identity/providers/supabase_adapter.go")), Evidence: "generated identity lifecycle runtime module and provider stubs present"},
			{ID: "api_behavioral_fixtures", Status: passFailStatus(hasPaths(job.Files, "/tests/behavior/scenarios.yaml", "/services/api/tests/behavior.sh", "/services/api/internal/runtime/behavior_test.go")), Evidence: "generated behavioral scenarios and executable fixtures present"},
		}
	case "verify":
		return job.Verification.Checks
	default:
		all := []VerificationCheck{}
		all = append(all, runTargetChecks(job, "web")...)
		all = append(all, runTargetChecks(job, "mobile")...)
		all = append(all, runTargetChecks(job, "api")...)
		all = append(all, job.Verification.Checks...)
		return all
	}
}

func hasPaths(files []FileArtifact, suffixes ...string) bool {
	for _, suffix := range suffixes {
		found := false
		for _, file := range files {
			if strings.HasSuffix(filepath.ToSlash(file.Path), suffix) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func hasCategory(files []ArtifactFile, category string) bool {
	for _, file := range files {
		if file.Category == category {
			return true
		}
	}
	return false
}

func hasConstraint(constraints []string, needle string) bool {
	needle = strings.ToLower(strings.TrimSpace(needle))
	for _, constraint := range constraints {
		if strings.ToLower(strings.TrimSpace(constraint)) == needle {
			return true
		}
	}
	return false
}

func behavioralPassRate(checks []VerificationCheck) float64 {
	behavioral := 0
	passed := 0
	for _, check := range checks {
		if strings.HasPrefix(check.ID, "behavioral_") {
			behavioral++
			if check.Status == "pass" {
				passed++
			}
		}
	}
	if behavioral == 0 {
		return 0
	}
	ratio := float64(passed) / float64(behavioral)
	return math.Round(ratio*100) / 100
}

func isDepthLabel(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "prototype", "pilot", "production-candidate":
		return true
	default:
		return false
	}
}

func passFailStatus(ok bool) string {
	if ok {
		return "pass"
	}
	return "fail"
}
