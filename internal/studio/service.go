package studio

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type Confirmation struct {
	Prompt           string   `json:"prompt"`
	AppName          string   `json:"app_name"`
	Domain           string   `json:"domain"`
	Template         string   `json:"template"`
	SourceSystem     string   `json:"source_system"`
	PrimaryUsers     []string `json:"primary_users"`
	CoreWorkflows    []string `json:"core_workflows"`
	DataEntities     []string `json:"data_entities"`
	DeploymentTarget string   `json:"deployment_target"`
	Region           string   `json:"region"`
	Plan             string   `json:"plan"`
	GenerationDepth  string   `json:"generation_depth"`
	Integrations     []string `json:"integrations"`
	Constraints      []string `json:"constraints"`
}

type WorkloadItem struct {
	Phase         string `json:"phase"`
	Task          string `json:"task"`
	Owner         string `json:"owner"`
	EstimateHours int    `json:"estimate_hours"`
	Status        string `json:"status"`
}

type FileArtifact struct {
	Path     string `json:"path"`
	Language string `json:"language"`
	Content  string `json:"content"`
}

type Job struct {
	JobID            string             `json:"job_id"`
	TenantID         string             `json:"tenant_id"`
	Status           string             `json:"status"`
	DepthLabel       string             `json:"depth_label"`
	CreatedAt        time.Time          `json:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at"`
	WorkspacePath    string             `json:"workspace_path"`
	Confirmation     Confirmation       `json:"confirmation"`
	Workload         []WorkloadItem     `json:"workload"`
	Files            []FileArtifact     `json:"files"`
	ArtifactManifest ArtifactManifest   `json:"artifact_manifest"`
	Verification     VerificationReport `json:"verification_report"`
	JTBDCoverage     []JTBDCoverage     `json:"jtbd_coverage"`
	TerminalLogs     []string           `json:"terminal_logs"`
	ConsoleLogs      []string           `json:"console_logs"`
	PreviewWorkload  string             `json:"preview_workload"`
	PreviewCodePath  string             `json:"preview_code_path"`
	PreviewTerminal  string             `json:"preview_terminal"`
	PreviewConsole   string             `json:"preview_console"`
}

type TerminalResult struct {
	Command string   `json:"command"`
	Output  []string `json:"output"`
	Cwd     string   `json:"cwd"`
}

type StudioPersistence interface {
	SaveStudioJob(ctx context.Context, tenantID, jobID string, payload []byte) error
	GetStudioJob(ctx context.Context, tenantID, jobID string) ([]byte, bool, error)
}

type Option func(*Service)

func WithPersistence(p StudioPersistence) Option {
	return func(s *Service) {
		s.persistence = p
	}
}

type Service struct {
	mu            sync.RWMutex
	jobs          map[string]Job
	workspaceRoot string
	persistence   StudioPersistence
}

func NewService(opts ...Option) *Service {
	root := strings.TrimSpace(os.Getenv("VDA_STUDIO_ROOT"))
	if root == "" {
		root = filepath.Join(".", "output", "studio")
	}
	_ = os.MkdirAll(root, 0o755)
	svc := &Service{
		jobs:          map[string]Job{},
		workspaceRoot: root,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(svc)
		}
	}
	return svc
}

func (s *Service) CreateJob(tenantID string, conf Confirmation) Job {
	now := time.Now().UTC()
	slug := slugify(conf.AppName)
	if slug == "" {
		slug = "generated-app"
	}
	jobID := makeID("job", tenantID, conf.Prompt, conf.AppName, now.Format(time.RFC3339Nano))

	if conf.DeploymentTarget == "" {
		conf.DeploymentTarget = "managed"
	}
	if conf.Region == "" {
		conf.Region = "us-east-1"
	}
	if conf.Plan == "" {
		conf.Plan = "starter"
	}
	conf.GenerationDepth = normalizeDepthLabel(conf.GenerationDepth)
	if conf.Template == "" {
		conf.Template = "violet-rails-extension"
	}
	if conf.SourceSystem == "" {
		conf.SourceSystem = "violet-rails"
	}
	if len(conf.PrimaryUsers) == 0 {
		conf.PrimaryUsers = []string{"admin", "operator"}
	}
	if len(conf.CoreWorkflows) == 0 {
		conf.CoreWorkflows = []string{"create_record", "approve_record", "notify_user"}
	}
	if len(conf.DataEntities) == 0 {
		conf.DataEntities = []string{"account", "workspace", "activity"}
	}
	conf.Constraints = mergeConstraints(conf.Constraints, []string{"all_mutations_idempotent", "no_runtime_eval"})

	workload := buildWorkload(conf)
	files := buildFiles(slug, conf)
	workspacePath, materializeErr := s.materializeWorkspace(jobID, files)
	terminal := []string{
		"$ scaffold init --template deterministic-saas",
		fmt.Sprintf("$ scaffold app --name \"%s\" --region %s --plan %s", conf.AppName, conf.Region, conf.Plan),
		"$ scaffold contracts --human-api --agent-api",
		"$ scaffold verify --checks schema,policy,deploy_preflight",
		fmt.Sprintf("$ scaffold deploy --target %s", conf.DeploymentTarget),
		"$ build complete",
	}
	console := []string{
		"[planner] prompt parsed and normalized",
		"[designer] structured confirmation converted to blueprint",
		"[builder] workload graph created",
		"[builder] code artifacts generated",
		"[runner] verify/deploy hooks prepared",
	}
	if materializeErr == nil {
		terminal = append(terminal, "$ cd "+workspacePath)
		console = append(console, "[builder] workspace materialized: "+workspacePath)
	} else {
		console = append(console, "[builder] workspace materialization_failed: "+materializeErr.Error())
	}

	job := Job{
		JobID:           jobID,
		TenantID:        tenantID,
		Status:          "generated",
		DepthLabel:      conf.GenerationDepth,
		CreatedAt:       now,
		UpdatedAt:       now,
		WorkspacePath:   workspacePath,
		Confirmation:    conf,
		Workload:        workload,
		Files:           files,
		TerminalLogs:    terminal,
		ConsoleLogs:     console,
		PreviewWorkload: "workload",
		PreviewCodePath: files[0].Path,
		PreviewTerminal: terminal[len(terminal)-1],
		PreviewConsole:  console[len(console)-1],
	}
	s.enrichJob(&job)

	s.mu.Lock()
	s.jobs[jobID] = job
	s.persistJobLocked(job)
	s.mu.Unlock()
	return job
}

func (s *Service) GetJob(tenantID, jobID string) (Job, bool) {
	s.mu.RLock()
	job, ok := s.jobs[jobID]
	s.mu.RUnlock()
	if ok {
		if job.TenantID != tenantID {
			return Job{}, false
		}
		if changed := s.ensureWorkspace(&job); changed {
			s.mu.Lock()
			s.jobs[jobID] = job
			s.persistJobLocked(job)
			s.mu.Unlock()
		}
		return job, true
	}

	loaded, found := s.loadPersistedJob(tenantID, jobID)
	if !found {
		return Job{}, false
	}
	s.mu.Lock()
	s.jobs[jobID] = loaded
	s.mu.Unlock()
	return loaded, true
}

func (s *Service) GetConsole(tenantID, jobID string) ([]string, bool) {
	job, ok := s.GetJob(tenantID, jobID)
	if !ok {
		return nil, false
	}
	return append([]string(nil), job.ConsoleLogs...), true
}

func (s *Service) RunTerminal(tenantID, jobID, command string) (TerminalResult, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	job, ok := s.jobs[jobID]
	if !ok || job.TenantID != tenantID {
		return TerminalResult{}, false
	}

	cwd := "/workspace"
	if strings.TrimSpace(job.WorkspacePath) != "" {
		cwd = job.WorkspacePath
	}
	result := TerminalResult{Command: command, Cwd: cwd}
	execPrefix := "exec "
	if strings.HasPrefix(strings.TrimSpace(command), execPrefix) && strings.TrimSpace(job.WorkspacePath) != "" {
		out, err := runShellCommand(job.WorkspacePath, strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(command), execPrefix)))
		result.Output = out
		if err != nil {
			result.Output = append(result.Output, "command_error: "+err.Error())
		}
	} else {
		result.Output = runPseudoCommand(command, job.Files)
	}

	ts := time.Now().UTC().Format(time.RFC3339)
	job.TerminalLogs = append(job.TerminalLogs, "$ "+command)
	job.TerminalLogs = append(job.TerminalLogs, result.Output...)
	job.ConsoleLogs = append(job.ConsoleLogs, fmt.Sprintf("[terminal %s] command executed: %s", ts, command))
	job.UpdatedAt = time.Now().UTC()
	s.enrichJob(&job)
	s.jobs[jobID] = job
	s.persistJobLocked(job)
	return result, true
}

func (s *Service) GetArtifacts(tenantID, jobID string) (ArtifactManifest, bool) {
	job, ok := s.GetJob(tenantID, jobID)
	if !ok {
		return ArtifactManifest{}, false
	}
	return job.ArtifactManifest, true
}

func (s *Service) GetVerification(tenantID, jobID string) (VerificationReport, bool) {
	job, ok := s.GetJob(tenantID, jobID)
	if !ok {
		return VerificationReport{}, false
	}
	return job.Verification, true
}

func (s *Service) GetJTBDCoverage(tenantID, jobID string) ([]JTBDCoverage, bool) {
	job, ok := s.GetJob(tenantID, jobID)
	if !ok {
		return nil, false
	}
	return append([]JTBDCoverage(nil), job.JTBDCoverage...), true
}

func (s *Service) RunTarget(tenantID, jobID, target string) (RunResult, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	job, ok := s.jobs[jobID]
	if !ok || job.TenantID != tenantID {
		loaded, found := s.loadPersistedJob(tenantID, jobID)
		if !found {
			return RunResult{}, false
		}
		job = loaded
	}

	if changed := s.ensureWorkspace(&job); changed {
		job.UpdatedAt = time.Now().UTC()
	}
	checks := runTargetChecks(job, target)
	targetName := strings.TrimSpace(strings.ToLower(target))
	if targetName == "" {
		targetName = "all"
	}
	if targetName == "api" || targetName == "all" {
		checks = append(checks, s.runGeneratedAPIRuntimeChecks(job)...)
	}
	status := "pass"
	for _, check := range checks {
		if check.Status == "fail" {
			status = "fail"
			break
		}
	}

	result := RunResult{
		Target:      targetName,
		Status:      status,
		Checks:      checks,
		GeneratedAt: time.Now().UTC(),
	}
	job.ConsoleLogs = append(job.ConsoleLogs, fmt.Sprintf("[runner %s] run target=%s status=%s", time.Now().UTC().Format(time.RFC3339), result.Target, result.Status))
	job.Status = "generated"
	job.UpdatedAt = time.Now().UTC()
	s.enrichJob(&job)
	s.jobs[jobID] = job
	s.persistJobLocked(job)
	return result, true
}

func buildWorkload(conf Confirmation) []WorkloadItem {
	items := []WorkloadItem{
		{Phase: "Shape", Task: "Finalize app contract and data entities", Owner: "product", EstimateHours: 4, Status: "ready"},
		{Phase: "Shape", Task: fmt.Sprintf("Align template: %s from %s", conf.Template, conf.SourceSystem), Owner: "platform", EstimateHours: 3, Status: "ready"},
		{Phase: "Build", Task: "Scaffold app blueprint and APIs", Owner: "platform", EstimateHours: 6, Status: "ready"},
		{Phase: "Build", Task: "Wire customization rules and mutation policies", Owner: "policy", EstimateHours: 5, Status: "ready"},
		{Phase: "Verify", Task: "Run machine-readable verification checks", Owner: "qa", EstimateHours: 3, Status: "ready"},
		{Phase: "Deploy", Task: fmt.Sprintf("Prepare %s deployment intent", conf.DeploymentTarget), Owner: "release", EstimateHours: 2, Status: "ready"},
	}
	for _, wf := range conf.CoreWorkflows {
		items = append(items, WorkloadItem{Phase: "Build", Task: "Implement workflow: " + wf, Owner: "platform", EstimateHours: 2, Status: "ready"})
	}
	return items
}

func buildFiles(slug string, conf Confirmation) []FileArtifact {
	entities := toYAMLList(conf.DataEntities)
	users := toYAMLList(conf.PrimaryUsers)
	workflows := toYAMLList(conf.CoreWorkflows)
	integrations := toYAMLList(conf.Integrations)
	constraints := toYAMLList(conf.Constraints)

	blueprint := fmt.Sprintf(`app:
  name: %s
  domain: %s
  template: %s
  source_system: %s
  plan: %s
  region: %s
  deployment_target: %s
primary_users:
%s
entities:
%s
workflows:
%s
integrations:
%s
constraints:
%s
`, conf.AppName, conf.Domain, conf.Template, conf.SourceSystem, conf.Plan, conf.Region, conf.DeploymentTarget, users, entities, workflows, integrations, constraints)

	openapi := fmt.Sprintf(`openapi: 3.1.0
info:
  title: %s API
  version: 0.1.0
paths:
  /v1/%s/health:
    get:
      responses:
        '200':
          description: OK
  /v1/%s/workflows/execute:
    post:
      responses:
        '200':
          description: Workflow execution result
`, conf.AppName, slug, slug)

	agentContract := fmt.Sprintf(`export interface AgentPlanRequest {
  prompt: string;
  target: "%s";
}

export interface AgentActRequest {
  mutationClass: string;
  payload: Record<string, unknown>;
}

export interface AgentVerifyResponse {
  verdict: "pass" | "fail";
  checks: Array<{ id: string; status: "pass" | "fail"; evidence: string }>;
}
`, conf.DeploymentTarget)

	workflowsJSON := fmt.Sprintf(`{
  "workflows": %s
}
`, mustJSONString(conf.CoreWorkflows))

	selfHostDeploy := fmt.Sprintf(`version: "3.9"
services:
  web:
    image: ghcr.io/violet/%s-web:latest
    ports:
      - "8080:8080"
    environment:
      - APP_NAME=%s
  api:
    image: ghcr.io/violet/%s-api:latest
    ports:
      - "8090:8090"
    environment:
      - POLICY_VERSION=v1
`, slug, conf.AppName, slug)

	managedDeployIntent := fmt.Sprintf(`{
  "target": "managed",
  "app_name": %q,
  "region": %q,
  "plan": %q,
  "requires_approval": true
}
`, conf.AppName, conf.Region, conf.Plan)

	smokeTest := fmt.Sprintf(`name: generated-smoke
description: Validate generated runtime and API contract for %s
checks:
  - id: web_runtime
    assert: clients/web/index.html exists
  - id: mobile_runtime
    assert: clients/mobile/index.html exists
  - id: api_contract
    assert: api/openapi.yaml exists
`, conf.AppName)

	packageJSON := fmt.Sprintf(`{
  "name": "%s",
  "private": true,
  "version": "0.1.0",
  "description": "Generated Violet Rails extension scaffold",
  "scripts": {
    "serve:web": "python3 -m http.server 4173 -d ./clients/web",
    "serve:mobile": "python3 -m http.server 4174 -d ./clients/mobile"
  }
}
`, slug)

	readme := fmt.Sprintf("# %s\n\nGenerated from prompt-driven confirmation.\n\n## Template\n\n- `%s` (source: `%s`)\n\n## Run\n\n- Validate blueprint\n- Execute verify checks\n- Create deploy intent (%s)\n",
		conf.AppName,
		conf.Template,
		conf.SourceSystem,
		conf.DeploymentTarget,
	)

	files := []FileArtifact{
		{Path: fmt.Sprintf("apps/%s/README.md", slug), Language: "markdown", Content: readme},
		{Path: fmt.Sprintf("apps/%s/package.json", slug), Language: "json", Content: packageJSON},
		{Path: fmt.Sprintf("apps/%s/blueprint.yaml", slug), Language: "yaml", Content: blueprint},
		{Path: fmt.Sprintf("apps/%s/api/openapi.yaml", slug), Language: "yaml", Content: openapi},
		{Path: fmt.Sprintf("apps/%s/src/agent_contract.ts", slug), Language: "typescript", Content: agentContract},
		{Path: fmt.Sprintf("apps/%s/workflows/definitions.json", slug), Language: "json", Content: workflowsJSON},
		{Path: fmt.Sprintf("apps/%s/tests/smoke.yaml", slug), Language: "yaml", Content: smokeTest},
		{Path: fmt.Sprintf("apps/%s/deploy/self-host/docker-compose.yaml", slug), Language: "yaml", Content: selfHostDeploy},
		{Path: fmt.Sprintf("apps/%s/deploy/managed/deploy-intent.json", slug), Language: "json", Content: managedDeployIntent},
		{Path: fmt.Sprintf("apps/%s/boilerplate/violet_rails_extension.md", slug), Language: "markdown", Content: violetRailsExtensionNotes(conf)},
	}
	files = append(files, runtimeSourceArtifacts(slug, conf)...)
	files = append(files, paritySupportArtifacts(slug, conf)...)
	files = append(files, backendRuntimeArtifacts(slug, conf)...)
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return files
}

func runPseudoCommand(command string, files []FileArtifact) []string {
	cmd := strings.TrimSpace(command)
	if cmd == "" {
		return []string{"no command provided"}
	}
	parts := strings.Fields(cmd)
	lookup := map[string]FileArtifact{}
	for _, f := range files {
		lookup[f.Path] = f
	}

	switch parts[0] {
	case "help":
		return []string{"supported: help, pwd, ls [prefix], tree, cat <path>, grep <term>, exec <shell-command>"}
	case "pwd":
		return []string{"/workspace"}
	case "ls":
		prefix := ""
		if len(parts) > 1 {
			prefix = strings.Trim(parts[1], "/")
		}
		return listPaths(files, prefix)
	case "tree":
		return listPaths(files, "")
	case "cat":
		if len(parts) < 2 {
			return []string{"usage: cat <path>"}
		}
		path := strings.Trim(parts[1], "/")
		f, ok := lookup[path]
		if !ok {
			return []string{"file not found: " + path}
		}
		return strings.Split(f.Content, "\n")
	case "grep":
		if len(parts) < 2 {
			return []string{"usage: grep <term>"}
		}
		term := strings.ToLower(strings.Join(parts[1:], " "))
		matches := []string{}
		for _, f := range files {
			for idx, line := range strings.Split(f.Content, "\n") {
				if strings.Contains(strings.ToLower(line), term) {
					matches = append(matches, fmt.Sprintf("%s:%d:%s", f.Path, idx+1, line))
				}
			}
		}
		if len(matches) == 0 {
			return []string{"no matches"}
		}
		return matches
	default:
		return []string{"unsupported command: " + parts[0], "try: help"}
	}
}

func listPaths(files []FileArtifact, prefix string) []string {
	out := []string{}
	for _, f := range files {
		if prefix == "" || strings.HasPrefix(f.Path, prefix) {
			out = append(out, f.Path)
		}
	}
	if len(out) == 0 {
		return []string{"(empty)"}
	}
	sort.Strings(out)
	return out
}

func (s *Service) enrichJob(job *Job) {
	if job == nil {
		return
	}
	if strings.TrimSpace(job.DepthLabel) == "" {
		job.DepthLabel = normalizeDepthLabel(job.Confirmation.GenerationDepth)
	}
	s.ensureWorkspace(job)
	job.ArtifactManifest = buildArtifactManifest(*job)
	job.Verification = buildVerificationReport(*job)
	job.JTBDCoverage = buildJTBDCoverage(*job)
}

func (s *Service) ensureWorkspace(job *Job) bool {
	if job == nil {
		return false
	}
	workspace := strings.TrimSpace(job.WorkspacePath)
	if workspace != "" {
		if st, err := os.Stat(workspace); err == nil && st.IsDir() {
			return false
		}
	}
	path, err := s.materializeWorkspace(job.JobID, job.Files)
	if err != nil {
		job.ConsoleLogs = append(job.ConsoleLogs, "[builder] workspace_rematerialization_failed: "+err.Error())
		return false
	}
	job.WorkspacePath = path
	job.ConsoleLogs = append(job.ConsoleLogs, "[builder] workspace_rematerialized: "+path)
	return true
}

func (s *Service) persistJobLocked(job Job) {
	if s.persistence == nil {
		return
	}
	payload, err := json.Marshal(job)
	if err != nil {
		return
	}
	_ = s.persistence.SaveStudioJob(context.Background(), job.TenantID, job.JobID, payload)
}

func (s *Service) loadPersistedJob(tenantID, jobID string) (Job, bool) {
	if s.persistence == nil {
		return Job{}, false
	}
	payload, found, err := s.persistence.GetStudioJob(context.Background(), tenantID, jobID)
	if err != nil || !found {
		return Job{}, false
	}
	var job Job
	if err := json.Unmarshal(payload, &job); err != nil {
		return Job{}, false
	}
	if job.TenantID != tenantID {
		return Job{}, false
	}
	if changed := s.ensureWorkspace(&job); changed {
		job.UpdatedAt = time.Now().UTC()
	}
	s.enrichJob(&job)
	return job, true
}

func (s *Service) materializeWorkspace(jobID string, files []FileArtifact) (string, error) {
	root := strings.TrimSpace(s.workspaceRoot)
	if root == "" {
		return "", nil
	}
	workspacePath := filepath.Join(root, jobID)
	if err := os.MkdirAll(workspacePath, 0o755); err != nil {
		return "", err
	}

	for _, f := range files {
		rel := filepath.Clean(strings.TrimPrefix(strings.TrimSpace(f.Path), "/"))
		if rel == "." || rel == "" {
			continue
		}
		abs := filepath.Join(workspacePath, rel)
		absClean := filepath.Clean(abs)
		workspaceClean := filepath.Clean(workspacePath) + string(os.PathSeparator)
		if !strings.HasPrefix(absClean+string(os.PathSeparator), workspaceClean) {
			return "", fmt.Errorf("invalid artifact path: %s", f.Path)
		}
		if err := os.MkdirAll(filepath.Dir(absClean), 0o755); err != nil {
			return "", err
		}
		if err := os.WriteFile(absClean, []byte(f.Content), 0o644); err != nil {
			return "", err
		}
	}
	return workspacePath, nil
}

func runShellCommand(cwd, command string) ([]string, error) {
	command = strings.TrimSpace(command)
	if command == "" {
		return []string{"usage: exec <shell-command>"}, nil
	}
	name, args, err := parseExecCommand(command)
	if err != nil {
		return []string{err.Error()}, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = cwd
	out, err := cmd.CombinedOutput()
	lines := strings.Split(strings.TrimRight(string(out), "\n"), "\n")
	if len(lines) == 1 && lines[0] == "" {
		lines = []string{"(no output)"}
	}
	if ctx.Err() == context.DeadlineExceeded {
		return append(lines, "command_timeout: exceeded 20s"), ctx.Err()
	}
	return lines, err
}

func parseExecCommand(command string) (string, []string, error) {
	forbidden := []string{"|", ";", "&", ">", "<", "`", "$("}
	for _, token := range forbidden {
		if strings.Contains(command, token) {
			return "", nil, fmt.Errorf("exec_rejected: forbidden shell token %q", token)
		}
	}
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return "", nil, fmt.Errorf("exec_rejected: empty command")
	}
	name := strings.TrimSpace(parts[0])
	allowed := map[string]struct{}{
		"pwd":  {},
		"ls":   {},
		"cat":  {},
		"grep": {},
		"head": {},
		"tail": {},
		"wc":   {},
		"find": {},
		"sed":  {},
		"echo": {},
	}
	if _, ok := allowed[name]; !ok {
		return "", nil, fmt.Errorf("exec_rejected: command %q not allowlisted", name)
	}
	args := parts[1:]
	for _, arg := range args {
		if strings.HasPrefix(arg, "/") {
			return "", nil, fmt.Errorf("exec_rejected: absolute paths are blocked")
		}
		if strings.Contains(arg, "..") {
			return "", nil, fmt.Errorf("exec_rejected: parent traversal is blocked")
		}
	}
	return name, args, nil
}

func toYAMLList(items []string) string {
	if len(items) == 0 {
		return "  - none"
	}
	rows := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		rows = append(rows, "  - "+item)
	}
	if len(rows) == 0 {
		return "  - none"
	}
	return strings.Join(rows, "\n")
}

func normalizeDepthLabel(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "prototype":
		return "prototype"
	case "production-candidate":
		return "production-candidate"
	case "pilot":
		return "pilot"
	default:
		return "pilot"
	}
}

func makeID(prefix string, parts ...string) string {
	payload := strings.Join(parts, "|")
	h := sha256.Sum256([]byte(payload))
	return prefix + "_" + hex.EncodeToString(h[:])[:16]
}

func slugify(in string) string {
	in = strings.ToLower(strings.TrimSpace(in))
	if in == "" {
		return ""
	}
	var b strings.Builder
	lastDash := false
	for _, ch := range in {
		isAlnum := (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9')
		if isAlnum {
			b.WriteRune(ch)
			lastDash = false
			continue
		}
		if !lastDash {
			b.WriteRune('-')
			lastDash = true
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "app"
	}
	return out
}

func violetRailsExtensionNotes(conf Confirmation) string {
	return fmt.Sprintf(`# Violet Rails Extension Boilerplate

## Source

- template: %s
- source_system: %s

## Generated intent

- Preserve deterministic API boundaries while extending Violet Rails behavior.
- Keep mutating actions idempotent and replay-safe.
- Expose all operator surfaces as API tools for human + AI loops.

## Suggested next implementation files

1. clients/web/src/modules/app-shell.tsx
2. clients/mobile/src/screens/home.tsx
3. services/api/src/routes/tenant-tools.ts
4. services/api/src/orchestration/langgraph-hooks.ts
`, conf.Template, conf.SourceSystem)
}

func mustJSONString(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func mergeConstraints(existing, required []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(existing)+len(required))
	appendConstraint := func(value string) {
		value = strings.TrimSpace(value)
		if value == "" {
			return
		}
		key := strings.ToLower(value)
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		out = append(out, value)
	}
	for _, value := range existing {
		appendConstraint(value)
	}
	for _, value := range required {
		appendConstraint(value)
	}
	return out
}
