package studio

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"strings"
	"testing"
)

func TestRenderPreviewAndRuntimeAssets(t *testing.T) {
	svc := NewService()
	job := svc.CreateJob("t_acme", Confirmation{
		Prompt:           "build runtime",
		AppName:          "Runtime Test",
		Domain:           "crm",
		Plan:             "enterprise",
		Region:           "us-east-1",
		DeploymentTarget: "managed",
		PrimaryUsers:     []string{"admin"},
		CoreWorkflows:    []string{"create_customer"},
		DataEntities:     []string{"account"},
		Integrations:     []string{"stripe"},
		Constraints:      []string{"all_mutations_idempotent"},
	})

	page, ok := svc.RenderPreview("t_acme", job.JobID, "web", "dev-token")
	if !ok {
		t.Fatalf("RenderPreview returned not found")
	}
	if !strings.Contains(page, "/runtime/web/app.css") {
		t.Fatalf("expected web css runtime asset path in preview page")
	}
	if !strings.Contains(page, "token=dev-token") {
		t.Fatalf("expected token query in preview page runtime assets")
	}

	contentType, body, ok := svc.RenderRuntimeAsset("t_acme", job.JobID, "web", "app.js")
	if !ok {
		t.Fatalf("RenderRuntimeAsset(web app.js) returned not found")
	}
	if !strings.Contains(contentType, "application/javascript") {
		t.Fatalf("expected javascript content-type, got %q", contentType)
	}
	if !strings.Contains(string(body), "Runtime Overview") {
		t.Fatalf("expected web runtime javascript content")
	}

	contentType, body, ok = svc.RenderRuntimeAsset("t_acme", job.JobID, "mobile", "app.js")
	if !ok {
		t.Fatalf("RenderRuntimeAsset(mobile app.js) returned not found")
	}
	if !strings.Contains(contentType, "application/javascript") {
		t.Fatalf("expected javascript content-type for mobile, got %q", contentType)
	}
	if !strings.Contains(string(body), "const state = { view: \"home\" }") {
		t.Fatalf("expected mobile runtime javascript content")
	}
}

func TestCreateJobIncludesRuntimeSourceArtifacts(t *testing.T) {
	svc := NewService()
	job := svc.CreateJob("t_acme", Confirmation{Prompt: "seed runtime source", AppName: "Artifacts"})

	hasWebJS := false
	hasMobileJS := false
	for _, file := range job.Files {
		if strings.HasSuffix(file.Path, "/clients/web/app.js") {
			hasWebJS = true
		}
		if strings.HasSuffix(file.Path, "/clients/mobile/app.js") {
			hasMobileJS = true
		}
	}
	if !hasWebJS {
		t.Fatalf("expected generated runtime source file for web client")
	}
	if !hasMobileJS {
		t.Fatalf("expected generated runtime source file for mobile client")
	}
}

func TestCreateJobEnforcesDeterministicPolicyConstraints(t *testing.T) {
	svc := NewService()
	job := svc.CreateJob("t_acme", Confirmation{
		Prompt:       "create app",
		AppName:      "Policy Defaults",
		Constraints:  []string{"ship_web_and_mobile_clients"},
		Integrations: []string{"stripe"},
	})

	if !hasConstraint(job.Confirmation.Constraints, "all_mutations_idempotent") {
		t.Fatalf("expected all_mutations_idempotent constraint to be enforced")
	}
	if !hasConstraint(job.Confirmation.Constraints, "no_runtime_eval") {
		t.Fatalf("expected no_runtime_eval constraint to be enforced")
	}
	if job.Verification.Verdict != "pass" {
		t.Fatalf("expected verification verdict pass, got %q", job.Verification.Verdict)
	}
}

func TestCreateJobIncludesBackendRuntimeArtifacts(t *testing.T) {
	svc := NewService()
	job := svc.CreateJob("t_acme", Confirmation{
		Prompt:       "build backend runtime",
		AppName:      "Backend Runtime",
		Integrations: []string{"stripe", "slack"},
	})

	required := []string{
		"/services/api/go.mod",
		"/services/api/cmd/server/main.go",
		"/services/api/internal/runtime/server.go",
		"/services/api/internal/runtime/server_test.go",
		"/services/api/internal/tools/catalog.go",
		"/services/api/tests/smoke.sh",
		"/services/api/internal/integrations/stripe_adapter.go",
		"/services/api/internal/integrations/slack_adapter.go",
	}
	for _, suffix := range required {
		if !hasPaths(job.Files, suffix) {
			t.Fatalf("expected generated backend artifact %s", suffix)
		}
	}

	result, found := svc.RunTarget("t_acme", job.JobID, "api")
	if !found {
		t.Fatalf("expected run target api to find job")
	}
	if result.Status != "pass" {
		t.Fatalf("expected api run target pass, got %q", result.Status)
	}
	ids := map[string]string{}
	for _, check := range result.Checks {
		ids[check.ID] = check.Status
	}
	if ids["api_runtime_go_test"] != "pass" {
		t.Fatalf("expected api_runtime_go_test pass, got %q", ids["api_runtime_go_test"])
	}
	if ids["api_runtime_health"] != "pass" {
		t.Fatalf("expected api_runtime_health pass, got %q", ids["api_runtime_health"])
	}
}

func TestBuildBundleContainsGeneratedWorkspace(t *testing.T) {
	svc := NewService()
	job := svc.CreateJob("t_acme", Confirmation{Prompt: "bundle workspace", AppName: "Bundle App"})

	filename, payload, found, err := svc.BuildBundle("t_acme", job.JobID)
	if err != nil {
		t.Fatalf("BuildBundle returned error: %v", err)
	}
	if !found {
		t.Fatalf("expected BuildBundle to find job")
	}
	if !strings.HasSuffix(filename, ".tar.gz") {
		t.Fatalf("expected bundle filename to end with .tar.gz, got %q", filename)
	}
	if len(payload) == 0 {
		t.Fatalf("expected non-empty bundle payload")
	}

	gzr, err := gzip.NewReader(bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("gzip reader: %v", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	foundManifest := false
	foundBlueprint := false
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("tar read: %v", err)
		}
		if strings.HasSuffix(hdr.Name, "/studio_artifact_manifest.json") {
			foundManifest = true
		}
		if strings.HasSuffix(hdr.Name, "/blueprint.yaml") {
			foundBlueprint = true
		}
	}
	if !foundManifest {
		t.Fatalf("expected bundle to include studio artifact manifest")
	}
	if !foundBlueprint {
		t.Fatalf("expected bundle to include blueprint.yaml")
	}
}
