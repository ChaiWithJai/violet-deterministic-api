package studio

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type runtimeSmokeResult struct {
	HealthOK         bool
	DepthLabelOK     bool
	ToolsOK          bool
	WorkflowOK       bool
	EntityRecordsOK  bool
	ActionExecuteOK  bool
	PrimitivesCMSOK  bool
	IdentityRoutesOK bool
	Output           string
}

func (s *Service) runGeneratedAPIRuntimeChecks(job Job) []VerificationCheck {
	serviceDir, ok := serviceWorkspacePath(job)
	if !ok {
		return []VerificationCheck{{
			ID:       "api_runtime_workspace",
			Status:   "fail",
			Evidence: "generated services/api workspace missing",
		}}
	}
	if _, err := exec.LookPath("go"); err != nil {
		return []VerificationCheck{{
			ID:       "api_runtime_toolchain",
			Status:   "fail",
			Evidence: "go toolchain unavailable in runner",
		}}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	testOutput, testErr := runGoCommand(ctx, serviceDir, "test", "./...")
	if testErr != nil {
		return []VerificationCheck{{
			ID:       "api_runtime_go_test",
			Status:   "fail",
			Evidence: trimEvidence("go test ./... failed: "+testErr.Error()+" | "+testOutput, 200),
		}}
	}

	smoke, runErr := runGeneratedServerChecks(ctx, serviceDir)
	if runErr != nil {
		return []VerificationCheck{
			{
				ID:       "api_runtime_go_test",
				Status:   "pass",
				Evidence: "go test ./... passed",
			},
			{
				ID:       "api_runtime_server_boot",
				Status:   "fail",
				Evidence: trimEvidence("go run ./cmd/server failed: "+runErr.Error()+" | "+smoke.Output, 240),
			},
		}
	}

	return []VerificationCheck{
		{
			ID:       "api_runtime_go_test",
			Status:   "pass",
			Evidence: "go test ./... passed",
		},
		{
			ID:       "api_runtime_server_boot",
			Status:   "pass",
			Evidence: "go run ./cmd/server booted successfully",
		},
		{
			ID:       "api_runtime_health",
			Status:   passFailStatus(smoke.HealthOK),
			Evidence: "GET /health responded with status=ok",
		},
		{
			ID:       "api_runtime_depth_label",
			Status:   passFailStatus(smoke.DepthLabelOK),
			Evidence: "GET /health included valid depth_label",
		},
		{
			ID:       "api_runtime_tools_catalog",
			Status:   passFailStatus(smoke.ToolsOK),
			Evidence: "GET /v1/tools returned tool list",
		},
		{
			ID:       "api_runtime_workflow_execute",
			Status:   passFailStatus(smoke.WorkflowOK),
			Evidence: "POST /v1/workflows/execute returned accepted",
		},
		{
			ID:       "api_runtime_entity_records",
			Status:   passFailStatus(smoke.EntityRecordsOK),
			Evidence: "GET /v1/entities/{entity}/records responded",
		},
		{
			ID:       "api_runtime_action_execute",
			Status:   passFailStatus(smoke.ActionExecuteOK),
			Evidence: "POST /v1/actions/execute returned accepted",
		},
		{
			ID:       "api_runtime_primitives",
			Status:   passFailStatus(smoke.PrimitivesCMSOK),
			Evidence: "GET /v1/primitives/cms/pages returned seeded payload",
		},
		{
			ID:       "api_runtime_identity",
			Status:   passFailStatus(smoke.IdentityRoutesOK),
			Evidence: "GET /v1/identity/providers returned provider stubs",
		},
	}
}

func runGoCommand(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = dir
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	return out.String(), err
}

func runGeneratedServerChecks(parent context.Context, dir string) (runtimeSmokeResult, error) {
	result := runtimeSmokeResult{}
	port, err := reserveFreePort()
	if err != nil {
		return result, err
	}
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	baseURL := "http://" + addr

	ctx, cancel := context.WithTimeout(parent, 25*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "run", "./cmd/server")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "PORT="+fmt.Sprintf("%d", port))
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Start(); err != nil {
		result.Output = out.String()
		return result, err
	}

	waitDone := make(chan error, 1)
	go func() {
		waitDone <- cmd.Wait()
	}()
	defer func() {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		select {
		case <-waitDone:
		case <-time.After(2 * time.Second):
		}
	}()

	start := time.Now()
	for time.Since(start) < 10*time.Second {
		select {
		case err := <-waitDone:
			if err != nil {
				result.Output = out.String()
				return result, err
			}
			result.Output = out.String()
			return result, fmt.Errorf("server exited before smoke checks")
		default:
		}
		healthOK, depthLabelOK := checkHealth(baseURL)
		if healthOK {
			result.HealthOK = true
			result.DepthLabelOK = depthLabelOK
			result.ToolsOK = checkTools(baseURL)
			result.WorkflowOK = checkWorkflowExecute(baseURL)
			result.EntityRecordsOK = checkEntityRecords(baseURL)
			result.ActionExecuteOK = checkExecuteAction(baseURL)
			result.PrimitivesCMSOK = checkPrimitivesCMS(baseURL)
			result.IdentityRoutesOK = checkIdentityProviders(baseURL)
			result.Output = out.String()
			return result, nil
		}
		time.Sleep(150 * time.Millisecond)
	}
	result.Output = out.String()
	return result, fmt.Errorf("health check timeout")
}

func reserveFreePort() (int, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer l.Close()
	addr := l.Addr().String()
	_, portText, found := strings.Cut(addr, ":")
	if !found {
		return 0, fmt.Errorf("port parse failed for addr %s", addr)
	}
	var port int
	_, err = fmt.Sscanf(portText, "%d", &port)
	if err != nil {
		return 0, err
	}
	return port, nil
}

func checkHealth(baseURL string) (bool, bool) {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(baseURL + "/health")
	if err != nil {
		return false, false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false, false
	}
	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return false, false
	}
	status, _ := payload["status"].(string)
	depthLabel, _ := payload["depth_label"].(string)
	return status == "ok", isDepthLabel(depthLabel)
}

func checkTools(baseURL string) bool {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(baseURL + "/v1/tools")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false
	}
	return bytes.Contains(body, []byte(`"tools"`))
}

func checkWorkflowExecute(baseURL string) bool {
	client := &http.Client{Timeout: 2 * time.Second}
	req, err := http.NewRequest(http.MethodPost, baseURL+"/v1/workflows/execute", strings.NewReader(`{"workflow":"smoke","input":{"mode":"test"}}`))
	if err != nil {
		return false
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false
	}
	return bytes.Contains(body, []byte(`"accepted"`))
}

func checkEntityRecords(baseURL string) bool {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(baseURL + "/v1/entities/account/records")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false
	}
	return bytes.Contains(body, []byte(`"records"`))
}

func checkExecuteAction(baseURL string) bool {
	client := &http.Client{Timeout: 2 * time.Second}
	req, err := http.NewRequest(http.MethodPost, baseURL+"/v1/actions/execute", strings.NewReader(`{"action":"approve_request","entity":"account","payload":{"mode":"runtime_smoke"}}`))
	if err != nil {
		return false
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false
	}
	return bytes.Contains(body, []byte(`"accepted"`))
}

func checkPrimitivesCMS(baseURL string) bool {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(baseURL + "/v1/primitives/cms/pages")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false
	}
	return bytes.Contains(body, []byte(`"pages"`))
}

func checkIdentityProviders(baseURL string) bool {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(baseURL + "/v1/identity/providers")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false
	}
	return bytes.Contains(body, []byte(`"providers"`))
}

func serviceWorkspacePath(job Job) (string, bool) {
	workspace := strings.TrimSpace(job.WorkspacePath)
	if workspace == "" {
		return "", false
	}
	for _, file := range job.Files {
		path := filepath.ToSlash(strings.TrimSpace(file.Path))
		if strings.HasSuffix(path, "/services/api/go.mod") {
			relDir := filepath.Dir(filepath.Clean(strings.TrimPrefix(path, "/")))
			absDir := filepath.Join(workspace, relDir)
			if st, err := os.Stat(absDir); err == nil && st.IsDir() {
				return absDir, true
			}
		}
	}
	return "", false
}

func trimEvidence(value string, max int) string {
	value = strings.TrimSpace(value)
	if len(value) <= max {
		return value
	}
	if max < 4 {
		return value[:max]
	}
	return value[:max-3] + "..."
}
