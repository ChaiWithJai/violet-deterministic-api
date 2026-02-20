package http

import (
	"encoding/json"
	"errors"
	"fmt"
	httpstd "net/http"
	"strings"
	"time"

	"github.com/restarone/violet-deterministic-api/internal/llm"
	"github.com/restarone/violet-deterministic-api/internal/studio"
)

type llmInferRequest struct {
	Provider         string                  `json:"provider"`
	Model            string                  `json:"model"`
	Prompt           string                  `json:"prompt"`
	System           string                  `json:"system,omitempty"`
	Temperature      float64                 `json:"temperature,omitempty"`
	MaxTokens        int                     `json:"max_tokens,omitempty"`
	PostHooks        []string                `json:"post_hooks,omitempty"`
	HookConfirmation *studioCreateJobRequest `json:"hook_confirmation,omitempty"`
}

func (s *Server) handleLLMProviders(w httpstd.ResponseWriter, r *httpstd.Request) {
	if _, ok := s.authClaims(w, r); !ok {
		return
	}
	ctx, cancel := withTimeout(r.Context(), 5*time.Second)
	defer cancel()

	providers := s.llm.ListProviders(ctx)
	writeJSONValue(w, httpstd.StatusOK, map[string]any{
		"default_provider": s.cfg.LLMDefaultProvider,
		"default_model":    s.cfg.LLMDefaultModel,
		"providers":        providers,
	})
}

func (s *Server) handleLLMInfer(w httpstd.ResponseWriter, r *httpstd.Request) {
	claims, ok := s.authClaims(w, r)
	if !ok {
		return
	}
	idemKey, ok := s.idempotencyKey(w, r)
	if !ok {
		return
	}

	var req llmInferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, httpstd.StatusBadRequest, "invalid_json", nil)
		return
	}
	if strings.TrimSpace(req.Prompt) == "" {
		writeError(w, httpstd.StatusBadRequest, "prompt_required", nil)
		return
	}

	s.withIdempotency(r.Context(), w, claims.TenantID, r.URL.Path, idemKey, func() (int, []byte, error) {
		ctx, cancel := withTimeout(r.Context(), time.Duration(s.cfg.LLMRequestTimeoutSecond)*time.Second)
		defer cancel()

		resp, err := s.llm.Infer(ctx, llm.InferRequest{
			Provider:    req.Provider,
			Model:       req.Model,
			Prompt:      req.Prompt,
			System:      req.System,
			Temperature: req.Temperature,
			MaxTokens:   req.MaxTokens,
		})
		if err != nil {
			var llmErr *llm.Error
			if errors.As(err, &llmErr) {
				status := httpstd.StatusBadGateway
				switch llmErr.Code {
				case "prompt_required", "model_required", "unknown_provider":
					status = httpstd.StatusBadRequest
				case "frontier_api_key_missing", "frontier_auth_required":
					status = httpstd.StatusServiceUnavailable
				}
				body, _ := json.Marshal(map[string]any{"error": llmErr.Code, "details": llmErr.Message})
				return status, body, nil
			}
			return 0, nil, err
		}

		hooks := make([]map[string]any, 0, len(req.PostHooks))
		if hasPostHook(req.PostHooks, "studio_generate") {
			conf := buildHookConfirmation(req, resp)
			job := s.studio.CreateJob(claims.TenantID, conf)
			hooks = append(hooks, map[string]any{
				"name":   "studio_generate",
				"status": "ok",
				"job_id": job.JobID,
				"summary": map[string]any{
					"workload_items": len(job.Workload),
					"files":          len(job.Files),
					"template":       conf.Template,
					"source_system":  conf.SourceSystem,
					"verification":   job.Verification.Verdict,
				},
				"paths": map[string]any{
					"job":            fmt.Sprintf("/v1/studio/jobs/%s", job.JobID),
					"artifacts":      fmt.Sprintf("/v1/studio/jobs/%s/artifacts", job.JobID),
					"verification":   fmt.Sprintf("/v1/studio/jobs/%s/verification", job.JobID),
					"jtbd":           fmt.Sprintf("/v1/studio/jobs/%s/jtbd", job.JobID),
					"bundle":         fmt.Sprintf("/v1/studio/jobs/%s/bundle", job.JobID),
					"preview_web":    fmt.Sprintf("/v1/studio/jobs/%s/preview?client=web", job.JobID),
					"preview_mobile": fmt.Sprintf("/v1/studio/jobs/%s/preview?client=mobile", job.JobID),
				},
			})
		}

		payload := map[string]any{
			"tenant_id": claims.TenantID,
			"result":    resp,
		}
		if len(hooks) > 0 {
			payload["hooks"] = hooks
		}
		body, err := json.Marshal(payload)
		if err != nil {
			return 0, nil, err
		}
		return httpstd.StatusOK, body, nil
	})
}

func hasPostHook(hooks []string, name string) bool {
	name = strings.TrimSpace(strings.ToLower(name))
	for _, hook := range hooks {
		if strings.TrimSpace(strings.ToLower(hook)) == name {
			return true
		}
	}
	return false
}

func buildHookConfirmation(req llmInferRequest, resp llm.InferResponse) studio.Confirmation {
	conf := studio.Confirmation{
		Prompt:           req.Prompt,
		AppName:          suggestedAppName(req.Prompt),
		Domain:           "saas",
		Template:         "violet-rails-extension",
		SourceSystem:     "violet-rails",
		PrimaryUsers:     []string{"admin", "operator"},
		CoreWorkflows:    []string{"design_app_contract", "generate_boilerplate", "run_verify_checks"},
		DataEntities:     []string{"tenant", "workspace", "subscription"},
		DeploymentTarget: "managed",
		Region:           "us-east-1",
		Plan:             "starter",
		Integrations:     []string{"stripe", "slack"},
		Constraints:      []string{"all_mutations_idempotent", "no_runtime_eval"},
	}

	mergedText := strings.ToLower(req.Prompt + "\n" + resp.Text)
	applyPromptSignals(&conf, mergedText)
	if strings.Contains(mergedText, "enterprise") {
		conf.Plan = "enterprise"
	}
	if strings.Contains(mergedText, "self-host") || strings.Contains(mergedText, "self host") {
		conf.DeploymentTarget = "self-host"
	}
	if strings.Contains(mergedText, "crm") {
		conf.Domain = "crm"
	}

	if req.HookConfirmation == nil {
		return conf
	}
	hc := req.HookConfirmation
	if v := strings.TrimSpace(hc.AppName); v != "" {
		conf.AppName = v
	}
	if v := strings.TrimSpace(hc.Domain); v != "" {
		conf.Domain = v
	}
	if v := strings.TrimSpace(hc.Template); v != "" {
		conf.Template = v
	}
	if v := strings.TrimSpace(hc.SourceSystem); v != "" {
		conf.SourceSystem = v
	}
	if v := strings.TrimSpace(hc.Plan); v != "" {
		conf.Plan = v
	}
	if v := strings.TrimSpace(hc.Region); v != "" {
		conf.Region = v
	}
	if v := strings.TrimSpace(hc.DeploymentTarget); v != "" {
		conf.DeploymentTarget = v
	}
	if v := normalizeList(hc.PrimaryUsers); len(v) > 0 {
		conf.PrimaryUsers = v
	}
	if v := normalizeList(hc.CoreWorkflows); len(v) > 0 {
		conf.CoreWorkflows = v
	}
	if v := normalizeList(hc.DataEntities); len(v) > 0 {
		conf.DataEntities = v
	}
	if v := normalizeList(hc.Integrations); len(v) > 0 {
		conf.Integrations = v
	}
	if v := normalizeList(hc.Constraints); len(v) > 0 {
		conf.Constraints = v
	}
	return conf
}

func applyPromptSignals(conf *studio.Confirmation, text string) {
	if conf == nil {
		return
	}
	switch {
	case strings.Contains(text, "crm"), strings.Contains(text, "sales"):
		conf.Domain = "crm"
		conf.PrimaryUsers = []string{"sales_manager", "account_executive", "operator"}
		conf.CoreWorkflows = []string{"capture_lead", "qualify_opportunity", "approve_quote", "issue_invoice"}
		conf.DataEntities = []string{"account", "contact", "opportunity", "invoice"}
		conf.Integrations = dedupeList(append(conf.Integrations, "salesforce", "hubspot"))
	case strings.Contains(text, "support"), strings.Contains(text, "helpdesk"):
		conf.Domain = "support"
		conf.PrimaryUsers = []string{"support_manager", "agent", "operator"}
		conf.CoreWorkflows = []string{"open_ticket", "triage_ticket", "approve_refund", "close_ticket"}
		conf.DataEntities = []string{"customer", "ticket", "conversation", "refund"}
		conf.Integrations = dedupeList(append(conf.Integrations, "zendesk", "intercom"))
	case strings.Contains(text, "marketplace"), strings.Contains(text, "commerce"), strings.Contains(text, "ecommerce"):
		conf.Domain = "commerce"
		conf.PrimaryUsers = []string{"merchant_admin", "operations", "finance"}
		conf.CoreWorkflows = []string{"publish_catalog", "approve_order", "capture_payment", "fulfill_order"}
		conf.DataEntities = []string{"merchant", "product", "order", "payment"}
		conf.Integrations = dedupeList(append(conf.Integrations, "shopify", "stripe"))
	}

	if strings.Contains(text, "mobile") {
		conf.Constraints = dedupeList(append(conf.Constraints, "ship_web_and_mobile_clients"))
	}
	if strings.Contains(text, "agent") || strings.Contains(text, "langgraph") {
		conf.CoreWorkflows = dedupeList(append(conf.CoreWorkflows, "agent_plan_act_verify_deploy"))
		conf.Constraints = dedupeList(append(conf.Constraints, "expose_api_as_tools"))
	}
	if strings.Contains(text, "rbac") || strings.Contains(text, "role") {
		conf.DataEntities = dedupeList(append(conf.DataEntities, "role", "permission"))
		conf.CoreWorkflows = dedupeList(append(conf.CoreWorkflows, "manage_roles", "grant_permissions"))
	}
}

func dedupeList(items []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		key := strings.ToLower(item)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, item)
	}
	return out
}

func suggestedAppName(prompt string) string {
	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		return "Generated App"
	}
	words := strings.Fields(prompt)
	if len(words) == 0 {
		return "Generated App"
	}
	if len(words) > 4 {
		words = words[:4]
	}
	for i := range words {
		words[i] = strings.Trim(words[i], " ,.!?;:\"'()[]{}")
		if words[i] == "" {
			continue
		}
		words[i] = strings.ToUpper(words[i][:1]) + strings.ToLower(words[i][1:])
	}
	name := strings.Join(words, " ")
	name = strings.TrimSpace(name)
	if name == "" {
		return "Generated App"
	}
	if len(name) > 48 {
		name = name[:48]
	}
	return name
}
