package http

import (
	"context"
	"encoding/json"
	"fmt"
	httpstd "net/http"
	"sort"
	"strings"
	"time"

	"github.com/restarone/violet-deterministic-api/internal/decision"
	"github.com/restarone/violet-deterministic-api/internal/storage"
)

func (s *Server) handleHealth(w httpstd.ResponseWriter, _ *httpstd.Request) {
	writeJSONValue(w, httpstd.StatusOK, map[string]any{
		"status":                            "ok",
		"service":                           "violet-deterministic-api",
		"policy_version":                    s.cfg.PolicyVersion,
		"data_version":                      s.cfg.DataVersion,
		"idempotency_cleanup_deleted_total": s.store.IdempotencyCleanupDeletedTotal(),
	})
}

func (s *Server) handleDecisions(w httpstd.ResponseWriter, r *httpstd.Request) {
	claims, ok := s.authClaims(w, r)
	if !ok {
		return
	}
	idemKey, ok := s.idempotencyKey(w, r)
	if !ok {
		return
	}

	var req decision.DecisionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, httpstd.StatusBadRequest, "invalid_json", nil)
		return
	}
	if req.TenantID == "" {
		req.TenantID = claims.TenantID
	}
	if req.TenantID != claims.TenantID {
		writeError(w, httpstd.StatusForbidden, "tenant_mismatch", nil)
		return
	}

	s.withIdempotency(r.Context(), w, claims.TenantID, r.URL.Path, idemKey, func() (int, []byte, error) {
		resp := s.engine.Decide(r.Context(), req)
		payload, err := json.Marshal(resp)
		if err != nil {
			return 0, nil, err
		}
		if err := s.store.SaveDecision(
			r.Context(),
			resp.DecisionID,
			claims.TenantID,
			resp.DecisionHash,
			resp.PolicyVersion,
			resp.DataVersion,
			resp.GeneratedAt,
			payload,
		); err != nil {
			return 0, nil, err
		}
		return httpstd.StatusOK, payload, nil
	})
}

func (s *Server) handleReplay(w httpstd.ResponseWriter, r *httpstd.Request) {
	claims, ok := s.authClaims(w, r)
	if !ok {
		return
	}

	var body struct {
		DecisionID string `json:"decision_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.DecisionID == "" {
		writeError(w, httpstd.StatusBadRequest, "invalid_request", nil)
		return
	}

	payload, found, err := s.store.GetDecisionPayload(r.Context(), body.DecisionID, claims.TenantID)
	if err != nil {
		writeError(w, httpstd.StatusInternalServerError, "replay_read_failed", map[string]any{"details": err.Error()})
		return
	}
	if !found {
		writeError(w, httpstd.StatusNotFound, "decision_not_found", nil)
		return
	}
	writeJSON(w, httpstd.StatusOK, payload)
}

func (s *Server) handleFeedback(w httpstd.ResponseWriter, r *httpstd.Request) {
	claims, ok := s.authClaims(w, r)
	if !ok {
		return
	}
	idemKey, ok := s.idempotencyKey(w, r)
	if !ok {
		return
	}

	var evt decision.FeedbackEvent
	if err := json.NewDecoder(r.Body).Decode(&evt); err != nil {
		writeError(w, httpstd.StatusBadRequest, "invalid_json", nil)
		return
	}

	s.withIdempotency(r.Context(), w, claims.TenantID, r.URL.Path, idemKey, func() (int, []byte, error) {
		payload, err := json.Marshal(map[string]any{
			"status":      "accepted",
			"decision_id": evt.DecisionID,
			"event_type":  evt.EventType,
			"actor":       claims.Subject,
			"tenant_id":   claims.TenantID,
		})
		if err != nil {
			return 0, nil, err
		}
		return httpstd.StatusAccepted, payload, nil
	})
}

type createAppRequest struct {
	Name      string         `json:"name"`
	Blueprint map[string]any `json:"blueprint"`
}

func (s *Server) handleCreateApp(w httpstd.ResponseWriter, r *httpstd.Request) {
	claims, ok := s.authClaims(w, r)
	if !ok {
		return
	}
	idemKey, ok := s.idempotencyKey(w, r)
	if !ok {
		return
	}

	var req createAppRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, httpstd.StatusBadRequest, "invalid_json", nil)
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		writeError(w, httpstd.StatusBadRequest, "name_required", nil)
		return
	}
	if req.Blueprint == nil {
		req.Blueprint = map[string]any{}
	}

	s.withIdempotency(r.Context(), w, claims.TenantID, r.URL.Path, idemKey, func() (int, []byte, error) {
		now := time.Now().UTC()
		app := storage.App{
			ID:        stableID("app", claims.TenantID, idemKey, req.Name),
			TenantID:  claims.TenantID,
			Name:      req.Name,
			Blueprint: req.Blueprint,
			Version:   1,
			CreatedAt: now,
			UpdatedAt: now,
		}
		created, err := s.store.CreateApp(r.Context(), app)
		if err != nil {
			return 0, nil, err
		}
		payload, err := json.Marshal(map[string]any{
			"app":            created,
			"policy_version": s.cfg.PolicyVersion,
			"data_version":   s.cfg.DataVersion,
		})
		if err != nil {
			return 0, nil, err
		}
		return httpstd.StatusCreated, payload, nil
	})
}

func (s *Server) handleGetApp(w httpstd.ResponseWriter, r *httpstd.Request) {
	claims, ok := s.authClaims(w, r)
	if !ok {
		return
	}
	appID := r.PathValue("id")
	if appID == "" {
		writeError(w, httpstd.StatusBadRequest, "app_id_required", nil)
		return
	}
	app, found, err := s.store.GetApp(r.Context(), claims.TenantID, appID)
	if err != nil {
		writeError(w, httpstd.StatusInternalServerError, "app_read_failed", map[string]any{"details": err.Error()})
		return
	}
	if !found {
		writeError(w, httpstd.StatusNotFound, "app_not_found", nil)
		return
	}
	writeJSONValue(w, httpstd.StatusOK, map[string]any{"app": app})
}

type patchAppRequest struct {
	Name           *string        `json:"name,omitempty"`
	BlueprintPatch map[string]any `json:"blueprint_patch,omitempty"`
}

func (s *Server) handlePatchApp(w httpstd.ResponseWriter, r *httpstd.Request) {
	claims, ok := s.authClaims(w, r)
	if !ok {
		return
	}
	idemKey, ok := s.idempotencyKey(w, r)
	if !ok {
		return
	}
	appID := r.PathValue("id")

	var req patchAppRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, httpstd.StatusBadRequest, "invalid_json", nil)
		return
	}

	s.withIdempotency(r.Context(), w, claims.TenantID, r.URL.Path, idemKey, func() (int, []byte, error) {
		app, found, err := s.store.GetApp(r.Context(), claims.TenantID, appID)
		if err != nil {
			return 0, nil, err
		}
		if !found {
			return httpstd.StatusNotFound, mustJSON(map[string]any{"error": "app_not_found"}), nil
		}
		if req.Name != nil && strings.TrimSpace(*req.Name) != "" {
			app.Name = strings.TrimSpace(*req.Name)
		}
		if req.BlueprintPatch != nil {
			for k, v := range req.BlueprintPatch {
				app.Blueprint[k] = v
			}
		}
		app.Version++
		app.UpdatedAt = time.Now().UTC()
		if err := s.store.UpdateApp(r.Context(), app); err != nil {
			return 0, nil, err
		}
		payload, err := json.Marshal(map[string]any{"app": app})
		if err != nil {
			return 0, nil, err
		}
		return httpstd.StatusOK, payload, nil
	})
}

type appMutationRequest struct {
	Class string `json:"class"`
	Path  string `json:"path,omitempty"`
	Value any    `json:"value,omitempty"`
}

func (s *Server) handleAppMutation(w httpstd.ResponseWriter, r *httpstd.Request) {
	claims, ok := s.authClaims(w, r)
	if !ok {
		return
	}
	idemKey, ok := s.idempotencyKey(w, r)
	if !ok {
		return
	}
	appID := r.PathValue("id")

	var req appMutationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, httpstd.StatusBadRequest, "invalid_json", nil)
		return
	}

	s.withIdempotency(r.Context(), w, claims.TenantID, r.URL.Path, idemKey, func() (int, []byte, error) {
		resp, status, err := s.executeMutation(r.Context(), claims.TenantID, appID, idemKey, req)
		if err != nil {
			return 0, nil, err
		}
		payload, err := json.Marshal(resp)
		if err != nil {
			return 0, nil, err
		}
		return status, payload, nil
	})
}

func (s *Server) executeMutation(ctx context.Context, tenantID, appID, idemKey string, req appMutationRequest) (map[string]any, int, error) {
	app, found, err := s.store.GetApp(ctx, tenantID, appID)
	if err != nil {
		return nil, 0, err
	}
	if !found {
		return map[string]any{"error": "app_not_found"}, httpstd.StatusNotFound, nil
	}

	policyOut, err := s.policy.Evaluate(ctx, tenantID, map[string]any{"mutation_class": req.Class})
	if err != nil {
		return nil, 0, err
	}
	if allowed, ok := policyOut["allowed"].(bool); ok && !allowed {
		return map[string]any{"error": "mutation_not_allowed", "class": req.Class}, httpstd.StatusForbidden, nil
	}

	beforeRaw, _ := json.Marshal(app)
	if err := applyMutation(&app, req); err != nil {
		return map[string]any{"error": "invalid_mutation", "details": err.Error()}, httpstd.StatusBadRequest, nil
	}
	app.Version++
	app.UpdatedAt = time.Now().UTC()
	if err := s.store.UpdateApp(ctx, app); err != nil {
		return nil, 0, err
	}
	afterRaw, _ := json.Marshal(app)
	mutationPayload, _ := json.Marshal(req)

	mutationID := stableID("mut", tenantID, appID, idemKey, req.Class)
	if err := s.store.SaveMutation(ctx, mutationID, tenantID, appID, req.Class, beforeRaw, afterRaw, mutationPayload); err != nil {
		return nil, 0, err
	}

	return map[string]any{
		"mutation_id":    mutationID,
		"policy_version": s.cfg.PolicyVersion,
		"app":            app,
	}, httpstd.StatusOK, nil
}

func applyMutation(app *storage.App, req appMutationRequest) error {
	switch req.Class {
	case "set_name":
		v, ok := req.Value.(string)
		if !ok || strings.TrimSpace(v) == "" {
			return fmt.Errorf("set_name requires non-empty string value")
		}
		app.Name = strings.TrimSpace(v)
	case "set_plan":
		v, ok := req.Value.(string)
		if !ok || strings.TrimSpace(v) == "" {
			return fmt.Errorf("set_plan requires non-empty string value")
		}
		if app.Blueprint == nil {
			app.Blueprint = map[string]any{}
		}
		app.Blueprint["plan"] = strings.TrimSpace(v)
	case "set_region":
		v, ok := req.Value.(string)
		if !ok || strings.TrimSpace(v) == "" {
			return fmt.Errorf("set_region requires non-empty string value")
		}
		if app.Blueprint == nil {
			app.Blueprint = map[string]any{}
		}
		app.Blueprint["region"] = strings.TrimSpace(v)
	case "set_feature_flag":
		if strings.TrimSpace(req.Path) == "" {
			return fmt.Errorf("set_feature_flag requires path")
		}
		v, ok := req.Value.(bool)
		if !ok {
			return fmt.Errorf("set_feature_flag requires bool value")
		}
		if app.Blueprint == nil {
			app.Blueprint = map[string]any{}
		}
		features, _ := app.Blueprint["features"].(map[string]any)
		if features == nil {
			features = map[string]any{}
		}
		features[req.Path] = v
		app.Blueprint["features"] = features
	default:
		return fmt.Errorf("unsupported mutation class")
	}
	return nil
}

func (s *Server) handleVerifyApp(w httpstd.ResponseWriter, r *httpstd.Request) {
	claims, ok := s.authClaims(w, r)
	if !ok {
		return
	}
	idemKey, ok := s.idempotencyKey(w, r)
	if !ok {
		return
	}
	appID := r.PathValue("id")

	s.withIdempotency(r.Context(), w, claims.TenantID, r.URL.Path, idemKey, func() (int, []byte, error) {
		resp, status, err := s.executeVerify(r.Context(), claims.TenantID, appID, idemKey)
		if err != nil {
			return 0, nil, err
		}
		payload, err := json.Marshal(resp)
		if err != nil {
			return 0, nil, err
		}
		return status, payload, nil
	})
}

func (s *Server) executeVerify(ctx context.Context, tenantID, appID, idemKey string) (map[string]any, int, error) {
	app, found, err := s.store.GetApp(ctx, tenantID, appID)
	if err != nil {
		return nil, 0, err
	}
	if !found {
		return map[string]any{"error": "app_not_found"}, httpstd.StatusNotFound, nil
	}

	checks := []map[string]any{}
	schemaPass := strings.TrimSpace(app.Name) != ""
	if app.Blueprint == nil {
		schemaPass = false
	}
	checks = append(checks, map[string]any{
		"id":       "schema",
		"status":   passFail(schemaPass),
		"evidence": "app name and blueprint present",
	})

	policyOut, err := s.policy.Evaluate(ctx, tenantID, map[string]any{"surface": "verify"})
	if err != nil {
		return nil, 0, err
	}
	policyPass := true
	if allowed, ok := policyOut["allowed"].(bool); ok {
		policyPass = allowed
	}
	checks = append(checks, map[string]any{
		"id":       "policy",
		"status":   passFail(policyPass),
		"evidence": fmt.Sprintf("policy_version=%s", s.cfg.PolicyVersion),
	})

	_, hasPlan := app.Blueprint["plan"]
	_, hasRegion := app.Blueprint["region"]
	preflightPass := hasPlan && hasRegion
	checks = append(checks, map[string]any{
		"id":       "deploy_preflight",
		"status":   passFail(preflightPass),
		"evidence": "plan and region set",
	})

	verdict := "pass"
	for _, check := range checks {
		if check["status"] == "fail" {
			verdict = "fail"
			break
		}
	}

	reportID := stableID("vrf", tenantID, appID, idemKey, fmt.Sprintf("%d", app.Version))
	resp := map[string]any{
		"report_id":      reportID,
		"app_id":         appID,
		"tenant_id":      tenantID,
		"verdict":        verdict,
		"checks":         checks,
		"policy_version": s.cfg.PolicyVersion,
		"data_version":   s.cfg.DataVersion,
		"generated_at":   time.Now().UTC(),
	}
	payload, _ := json.Marshal(resp)
	if err := s.store.SaveVerifyReport(ctx, reportID, tenantID, appID, payload); err != nil {
		return nil, 0, err
	}
	return resp, httpstd.StatusOK, nil
}

func passFail(ok bool) string {
	if ok {
		return "pass"
	}
	return "fail"
}

func (s *Server) handleDeploySelfHost(w httpstd.ResponseWriter, r *httpstd.Request) {
	s.handleDeployIntent(w, r, "self-host")
}

func (s *Server) handleDeployManaged(w httpstd.ResponseWriter, r *httpstd.Request) {
	s.handleDeployIntent(w, r, "managed")
}

type deployIntentRequest struct {
	Profile map[string]any `json:"profile"`
}

func (s *Server) handleDeployIntent(w httpstd.ResponseWriter, r *httpstd.Request, target string) {
	claims, ok := s.authClaims(w, r)
	if !ok {
		return
	}
	idemKey, ok := s.idempotencyKey(w, r)
	if !ok {
		return
	}
	appID := r.PathValue("id")

	var req deployIntentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, httpstd.StatusBadRequest, "invalid_json", nil)
		return
	}

	s.withIdempotency(r.Context(), w, claims.TenantID, r.URL.Path, idemKey, func() (int, []byte, error) {
		resp, status, err := s.executeDeploy(r.Context(), claims.TenantID, appID, idemKey, target, req)
		if err != nil {
			return 0, nil, err
		}
		payload, err := json.Marshal(resp)
		if err != nil {
			return 0, nil, err
		}
		return status, payload, nil
	})
}

func (s *Server) executeDeploy(ctx context.Context, tenantID, appID, idemKey, target string, req deployIntentRequest) (map[string]any, int, error) {
	app, found, err := s.store.GetApp(ctx, tenantID, appID)
	if err != nil {
		return nil, 0, err
	}
	if !found {
		return map[string]any{"error": "app_not_found"}, httpstd.StatusNotFound, nil
	}

	_, hasPlan := app.Blueprint["plan"]
	_, hasRegion := app.Blueprint["region"]
	if !hasPlan || !hasRegion {
		return map[string]any{
			"error":   "preflight_failed",
			"details": "app blueprint must include plan and region before deploy intent",
		}, httpstd.StatusBadRequest, nil
	}

	intentID := stableID("dep", tenantID, appID, target, idemKey)
	resp := map[string]any{
		"intent_id":           intentID,
		"app_id":              appID,
		"tenant_id":           tenantID,
		"target":              target,
		"approval_required":   true,
		"status":              "pending_approval",
		"profile":             req.Profile,
		"policy_version":      s.cfg.PolicyVersion,
		"data_version":        s.cfg.DataVersion,
		"requested_at":        time.Now().UTC(),
		"orchestration_hints": map[string]any{"next": []string{"human_approval", "execution"}},
	}
	payload, _ := json.Marshal(resp)
	if err := s.store.SaveDeployIntent(ctx, intentID, tenantID, appID, target, payload); err != nil {
		return nil, 0, err
	}
	return resp, httpstd.StatusAccepted, nil
}

func mustJSON(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}

func withTimeout(ctx context.Context, d time.Duration) (context.Context, context.CancelFunc) {
	if d <= 0 {
		return context.WithCancel(ctx)
	}
	return context.WithTimeout(ctx, d)
}

type agentPlanRequest struct {
	Prompt string `json:"prompt"`
	Name   string `json:"name,omitempty"`
}

func (s *Server) handleAgentPlan(w httpstd.ResponseWriter, r *httpstd.Request) {
	claims, ok := s.authClaims(w, r)
	if !ok {
		return
	}
	idemKey, ok := s.idempotencyKey(w, r)
	if !ok {
		return
	}
	var req agentPlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, httpstd.StatusBadRequest, "invalid_json", nil)
		return
	}
	if strings.TrimSpace(req.Prompt) == "" {
		writeError(w, httpstd.StatusBadRequest, "prompt_required", nil)
		return
	}
	s.withIdempotency(r.Context(), w, claims.TenantID, r.URL.Path, idemKey, func() (int, []byte, error) {
		name := req.Name
		if strings.TrimSpace(name) == "" {
			name = "Generated App"
		}
		plan := "starter"
		if strings.Contains(strings.ToLower(req.Prompt), "enterprise") {
			plan = "enterprise"
		}
		resp := map[string]any{
			"plan_id":   stableID("plan", claims.TenantID, idemKey, req.Prompt),
			"tenant_id": claims.TenantID,
			"name":      name,
			"suggested_blueprint": map[string]any{
				"plan":   plan,
				"region": "us-east-1",
			},
			"checks":         []string{"schema", "policy", "deploy_preflight"},
			"policy_version": s.cfg.PolicyVersion,
			"data_version":   s.cfg.DataVersion,
		}
		payload, err := json.Marshal(resp)
		if err != nil {
			return 0, nil, err
		}
		return httpstd.StatusOK, payload, nil
	})
}

type agentClarifyRequest struct {
	Prompt       string                 `json:"prompt"`
	Confirmation studioCreateJobRequest `json:"confirmation"`
	Answers      map[string]string      `json:"answers,omitempty"`
}

type agentClarifyQuestion struct {
	ID      string   `json:"id"`
	Field   string   `json:"field"`
	Prompt  string   `json:"prompt"`
	Why     string   `json:"why"`
	Options []string `json:"options,omitempty"`
}

func (s *Server) handleAgentClarify(w httpstd.ResponseWriter, r *httpstd.Request) {
	claims, ok := s.authClaims(w, r)
	if !ok {
		return
	}
	idemKey, ok := s.idempotencyKey(w, r)
	if !ok {
		return
	}
	var req agentClarifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, httpstd.StatusBadRequest, "invalid_json", nil)
		return
	}
	prompt := strings.TrimSpace(req.Prompt)
	if prompt == "" {
		prompt = strings.TrimSpace(req.Confirmation.Prompt)
	}
	if prompt == "" {
		writeError(w, httpstd.StatusBadRequest, "prompt_required", nil)
		return
	}

	s.withIdempotency(r.Context(), w, claims.TenantID, r.URL.Path, idemKey, func() (int, []byte, error) {
		conf := req.Confirmation
		conf.Prompt = prompt
		applyPromptDefaultsForClarify(&conf, prompt)
		applyClarifyAnswers(&conf, req.Answers)
		conf.PrimaryUsers = normalizeList(conf.PrimaryUsers)
		conf.CoreWorkflows = normalizeList(conf.CoreWorkflows)
		conf.DataEntities = normalizeList(conf.DataEntities)
		conf.Integrations = normalizeList(conf.Integrations)
		conf.Constraints = dedupeList(conf.Constraints)

		questions := buildClarificationQuestions(prompt, conf, req.Answers)
		missing := clarificationMissingFields(questions)
		resp := map[string]any{
			"clarification_id":     stableID("clarify", claims.TenantID, idemKey, prompt),
			"tenant_id":            claims.TenantID,
			"answer_count":         len(req.Answers),
			"ready_to_generate":    len(questions) == 0,
			"remaining_questions":  len(questions),
			"missing_fields":       missing,
			"summary":              fmt.Sprintf("Captured %d answer(s). %d clarification question(s) remain.", len(req.Answers), len(questions)),
			"updated_confirmation": conf,
			"questions":            questions,
			"policy_version":       s.cfg.PolicyVersion,
			"data_version":         s.cfg.DataVersion,
		}
		payload, err := json.Marshal(resp)
		if err != nil {
			return 0, nil, err
		}
		return httpstd.StatusOK, payload, nil
	})
}

func applyPromptDefaultsForClarify(conf *studioCreateJobRequest, prompt string) {
	if conf == nil {
		return
	}
	text := strings.ToLower(prompt)

	if strings.TrimSpace(conf.Prompt) == "" {
		conf.Prompt = strings.TrimSpace(prompt)
	}
	if strings.TrimSpace(conf.Template) == "" {
		conf.Template = "violet-rails-extension"
	}
	if strings.TrimSpace(conf.SourceSystem) == "" {
		conf.SourceSystem = "violet-rails"
	}
	if strings.TrimSpace(conf.Plan) == "" {
		conf.Plan = "starter"
	}
	if strings.TrimSpace(conf.Region) == "" {
		conf.Region = "us-east-1"
	}
	if strings.TrimSpace(conf.DeploymentTarget) == "" {
		conf.DeploymentTarget = "managed"
	}

	if hasAnyToken(text, "enterprise", "soc2", "sso") {
		conf.Plan = "enterprise"
	}
	if hasAnyToken(text, "self-host", "self host", "on-prem", "on prem", "kubernetes") {
		conf.DeploymentTarget = "self-host"
	}

	if strings.TrimSpace(conf.Domain) == "" || strings.EqualFold(conf.Domain, "saas") {
		switch {
		case hasAnyToken(text, "crm", "sales"):
			conf.Domain = "crm"
		case hasAnyToken(text, "support", "helpdesk"):
			conf.Domain = "support"
		case hasAnyToken(text, "marketplace", "commerce", "ecommerce"):
			conf.Domain = "commerce"
		default:
			if strings.TrimSpace(conf.Domain) == "" {
				conf.Domain = "saas"
			}
		}
	}

}

func applyClarifyAnswers(conf *studioCreateJobRequest, answers map[string]string) {
	if conf == nil || len(answers) == 0 {
		return
	}
	keys := make([]string, 0, len(answers))
	for key := range answers {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		field := canonicalClarifyField(key)
		value := strings.TrimSpace(answers[key])
		if value == "" {
			continue
		}
		switch field {
		case "app_name":
			conf.AppName = value
		case "domain":
			conf.Domain = value
		case "template":
			conf.Template = value
		case "source_system":
			conf.SourceSystem = value
		case "plan":
			switch strings.ToLower(value) {
			case "starter", "enterprise":
				conf.Plan = strings.ToLower(value)
			default:
				conf.Plan = value
			}
		case "region":
			conf.Region = value
		case "deployment_target":
			switch strings.ToLower(value) {
			case "self-host", "self host":
				conf.DeploymentTarget = "self-host"
			case "managed":
				conf.DeploymentTarget = "managed"
			default:
				conf.DeploymentTarget = value
			}
		case "primary_users":
			if parsed := parseDelimitedList(value); len(parsed) > 0 {
				conf.PrimaryUsers = parsed
			}
		case "core_workflows":
			if parsed := parseDelimitedList(value); len(parsed) > 0 {
				conf.CoreWorkflows = parsed
			}
		case "data_entities":
			if parsed := parseDelimitedList(value); len(parsed) > 0 {
				conf.DataEntities = parsed
			}
		case "integrations":
			parsed := parseDelimitedList(value)
			if len(parsed) == 1 && strings.EqualFold(parsed[0], "none") {
				conf.Integrations = nil
			} else if len(parsed) > 0 {
				conf.Integrations = parsed
			}
		case "constraints":
			parsed := parseDelimitedList(value)
			filtered := make([]string, 0, len(parsed))
			for _, item := range parsed {
				if strings.HasPrefix(strings.ToLower(item), "no_extra_") {
					continue
				}
				filtered = append(filtered, item)
			}
			if len(filtered) > 0 {
				conf.Constraints = dedupeList(append(conf.Constraints, filtered...))
			}
		}
	}
}

func buildClarificationQuestions(prompt string, conf studioCreateJobRequest, answers map[string]string) []agentClarifyQuestion {
	answered := clarificationAnswerSet(answers)
	questions := make([]agentClarifyQuestion, 0, 8)
	asked := map[string]struct{}{}
	add := func(field, question, why string, options []string) {
		if _, ok := answered[field]; ok {
			return
		}
		if _, ok := asked[field]; ok {
			return
		}
		asked[field] = struct{}{}
		questions = append(questions, agentClarifyQuestion{
			ID:      field,
			Field:   field,
			Prompt:  question,
			Why:     why,
			Options: options,
		})
	}

	text := strings.ToLower(prompt)
	if isDefaultAppName(conf.AppName) {
		add("app_name", "What should we call this app?", "App name is still generic and will appear across previews, package names, and artifacts.", nil)
	}
	if strings.TrimSpace(conf.Domain) == "" || (strings.EqualFold(conf.Domain, "saas") && !hasAnyToken(text, "crm", "support", "helpdesk", "marketplace", "commerce", "ecommerce", "billing", "hr", "analytics")) {
		add("domain", "Which domain best matches your product?", "Domain sharpens default workflows, entities, and quality checks.", []string{"crm", "support", "commerce", "billing", "project-management"})
	}
	if strings.TrimSpace(conf.Plan) == "" || (hasAnyToken(text, "enterprise", "soc2", "sso") && !strings.EqualFold(conf.Plan, "enterprise")) {
		add("plan", "What release tier should we target first?", "Tier changes security and operational guardrails in generated scaffolds.", []string{"starter", "enterprise"})
	}
	if strings.TrimSpace(conf.DeploymentTarget) == "" || (hasAnyToken(text, "self-host", "self host", "on-prem", "on prem", "kubernetes") && !strings.EqualFold(conf.DeploymentTarget, "self-host")) {
		add("deployment_target", "Where should this app run first?", "Hosting target controls deploy scripts and infra assumptions.", []string{"managed", "self-host"})
	}
	if len(conf.PrimaryUsers) < 2 {
		add("primary_users", "Who are the primary users (comma or newline separated)?", "User roles drive API contracts, permissions, and UI navigation.", nil)
	}
	if len(conf.CoreWorkflows) < 3 {
		add("core_workflows", "List 3-5 must-have workflows.", "Workflow depth determines generated routes, screens, and verify targets.", nil)
	}
	if len(conf.DataEntities) < 3 {
		add("data_entities", "List key data entities (comma or newline separated).", "Entities define schema, CRUD flows, and API contract coverage.", nil)
	}
	if len(conf.Integrations) == 0 {
		add("integrations", "Any integrations needed for v1?", "Integrations affect secret wiring, jobs, and deployment requirements.", []string{"stripe", "slack", "sendgrid", "none"})
	}

	requiredConstraints := make([]string, 0, 2)
	if hasAnyToken(text, "mobile", "ios", "android") && !containsFold(conf.Constraints, "ship_web_and_mobile_clients") {
		requiredConstraints = append(requiredConstraints, "ship_web_and_mobile_clients")
	}
	if hasAnyToken(text, "agent", "ai", "langgraph", "tool", "cli") && !containsFold(conf.Constraints, "expose_api_as_tools") {
		requiredConstraints = append(requiredConstraints, "expose_api_as_tools")
	}
	if len(requiredConstraints) > 0 {
		options := append([]string{}, requiredConstraints...)
		options = append(options, "no_extra_constraints")
		add("constraints", fmt.Sprintf("Should we enforce these constraints: %s ?", strings.Join(requiredConstraints, ", ")), "Constraints guarantee generated output stays aligned with your stated operating model.", options)
	}

	if len(questions) > 3 {
		return questions[:3]
	}
	return questions
}

func clarificationAnswerSet(answers map[string]string) map[string]struct{} {
	set := make(map[string]struct{}, len(answers))
	for key, value := range answers {
		if strings.TrimSpace(value) == "" {
			continue
		}
		set[canonicalClarifyField(key)] = struct{}{}
	}
	return set
}

func clarificationMissingFields(questions []agentClarifyQuestion) []string {
	out := make([]string, 0, len(questions))
	seen := map[string]struct{}{}
	for _, question := range questions {
		field := strings.TrimSpace(question.Field)
		if field == "" {
			continue
		}
		if _, ok := seen[field]; ok {
			continue
		}
		seen[field] = struct{}{}
		out = append(out, field)
	}
	return out
}

func canonicalClarifyField(field string) string {
	switch strings.ToLower(strings.TrimSpace(field)) {
	case "name", "app", "app_name":
		return "app_name"
	case "domain":
		return "domain"
	case "template":
		return "template"
	case "source", "source_system":
		return "source_system"
	case "plan", "tier":
		return "plan"
	case "region":
		return "region"
	case "deployment", "deployment_target", "target", "hosting":
		return "deployment_target"
	case "users", "primary_users", "personas":
		return "primary_users"
	case "workflows", "core_workflows":
		return "core_workflows"
	case "entities", "data_entities", "models":
		return "data_entities"
	case "integrations":
		return "integrations"
	case "constraints":
		return "constraints"
	default:
		return strings.ToLower(strings.TrimSpace(field))
	}
}

func parseDelimitedList(raw string) []string {
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		return r == '\n' || r == ',' || r == ';'
	})
	return normalizeList(parts)
}

func containsFold(items []string, target string) bool {
	target = strings.TrimSpace(strings.ToLower(target))
	for _, item := range items {
		if strings.TrimSpace(strings.ToLower(item)) == target {
			return true
		}
	}
	return false
}

func hasAnyToken(text string, tokens ...string) bool {
	for _, token := range tokens {
		if strings.Contains(text, strings.ToLower(token)) {
			return true
		}
	}
	return false
}

func isDefaultAppName(name string) bool {
	trimmed := strings.TrimSpace(strings.ToLower(name))
	return trimmed == "" || trimmed == "generated app"
}

type agentActRequest struct {
	AppID string `json:"app_id"`
	Class string `json:"class"`
	Path  string `json:"path,omitempty"`
	Value any    `json:"value,omitempty"`
}

func (s *Server) handleAgentAct(w httpstd.ResponseWriter, r *httpstd.Request) {
	claims, ok := s.authClaims(w, r)
	if !ok {
		return
	}
	idemKey, ok := s.idempotencyKey(w, r)
	if !ok {
		return
	}
	var req agentActRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, httpstd.StatusBadRequest, "invalid_json", nil)
		return
	}
	if strings.TrimSpace(req.AppID) == "" {
		writeError(w, httpstd.StatusBadRequest, "app_id_required", nil)
		return
	}
	s.withIdempotency(r.Context(), w, claims.TenantID, r.URL.Path, idemKey, func() (int, []byte, error) {
		resp, status, err := s.executeMutation(r.Context(), claims.TenantID, req.AppID, idemKey, appMutationRequest{
			Class: req.Class,
			Path:  req.Path,
			Value: req.Value,
		})
		if err != nil {
			return 0, nil, err
		}
		resp["actor"] = "agent"
		resp["subject"] = claims.Subject
		payload, err := json.Marshal(resp)
		if err != nil {
			return 0, nil, err
		}
		return status, payload, nil
	})
}

type agentVerifyRequest struct {
	AppID string `json:"app_id"`
}

func (s *Server) handleAgentVerify(w httpstd.ResponseWriter, r *httpstd.Request) {
	claims, ok := s.authClaims(w, r)
	if !ok {
		return
	}
	idemKey, ok := s.idempotencyKey(w, r)
	if !ok {
		return
	}
	var req agentVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, httpstd.StatusBadRequest, "invalid_json", nil)
		return
	}
	if strings.TrimSpace(req.AppID) == "" {
		writeError(w, httpstd.StatusBadRequest, "app_id_required", nil)
		return
	}
	s.withIdempotency(r.Context(), w, claims.TenantID, r.URL.Path, idemKey, func() (int, []byte, error) {
		resp, status, err := s.executeVerify(r.Context(), claims.TenantID, req.AppID, idemKey)
		if err != nil {
			return 0, nil, err
		}
		resp["actor"] = "agent"
		resp["subject"] = claims.Subject
		payload, err := json.Marshal(resp)
		if err != nil {
			return 0, nil, err
		}
		return status, payload, nil
	})
}

type agentDeployRequest struct {
	AppID   string         `json:"app_id"`
	Target  string         `json:"target"`
	Profile map[string]any `json:"profile"`
}

func (s *Server) handleAgentDeploy(w httpstd.ResponseWriter, r *httpstd.Request) {
	claims, ok := s.authClaims(w, r)
	if !ok {
		return
	}
	idemKey, ok := s.idempotencyKey(w, r)
	if !ok {
		return
	}
	var req agentDeployRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, httpstd.StatusBadRequest, "invalid_json", nil)
		return
	}
	if strings.TrimSpace(req.AppID) == "" {
		writeError(w, httpstd.StatusBadRequest, "app_id_required", nil)
		return
	}
	target := strings.TrimSpace(req.Target)
	if target != "self-host" && target != "managed" {
		writeError(w, httpstd.StatusBadRequest, "invalid_target", map[string]any{"supported": []string{"self-host", "managed"}})
		return
	}
	s.withIdempotency(r.Context(), w, claims.TenantID, r.URL.Path, idemKey, func() (int, []byte, error) {
		resp, status, err := s.executeDeploy(r.Context(), claims.TenantID, req.AppID, idemKey, target, deployIntentRequest{Profile: req.Profile})
		if err != nil {
			return 0, nil, err
		}
		resp["actor"] = "agent"
		resp["subject"] = claims.Subject
		payload, err := json.Marshal(resp)
		if err != nil {
			return 0, nil, err
		}
		return status, payload, nil
	})
}
