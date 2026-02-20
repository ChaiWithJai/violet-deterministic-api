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

	healthOK, toolsOK, workflowOK, runOutput, runErr := runGeneratedServerChecks(ctx, serviceDir)
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
				Evidence: trimEvidence("go run ./cmd/server failed: "+runErr.Error()+" | "+runOutput, 240),
			},
		}
	}

	healthCheck := VerificationCheck{
		ID:       "api_runtime_health",
		Status:   passFailStatus(healthOK),
		Evidence: "GET /health responded with status=ok",
	}
	toolsCheck := VerificationCheck{
		ID:       "api_runtime_tools_catalog",
		Status:   passFailStatus(toolsOK),
		Evidence: "GET /v1/tools returned tool list",
	}
	workflowCheck := VerificationCheck{
		ID:       "api_runtime_workflow_execute",
		Status:   passFailStatus(workflowOK),
		Evidence: "POST /v1/workflows/execute returned accepted",
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
		healthCheck,
		toolsCheck,
		workflowCheck,
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

func runGeneratedServerChecks(parent context.Context, dir string) (bool, bool, bool, string, error) {
	port, err := reserveFreePort()
	if err != nil {
		return false, false, false, "", err
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
		return false, false, false, out.String(), err
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
				return false, false, false, out.String(), err
			}
			return false, false, false, out.String(), fmt.Errorf("server exited before smoke checks")
		default:
		}
		ok := checkHealth(baseURL)
		if ok {
			toolsOK := checkTools(baseURL)
			workflowOK := checkWorkflowExecute(baseURL)
			return true, toolsOK, workflowOK, out.String(), nil
		}
		time.Sleep(150 * time.Millisecond)
	}
	return false, false, false, out.String(), fmt.Errorf("health check timeout")
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

func checkHealth(baseURL string) bool {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(baseURL + "/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false
	}
	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return false
	}
	status, _ := payload["status"].(string)
	return status == "ok"
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
