package http

import (
	"encoding/json"
	"fmt"
	httpstd "net/http"
	"strings"
	"time"

	"github.com/restarone/violet-deterministic-api/internal/auth"
	"github.com/restarone/violet-deterministic-api/internal/studio"
)

type studioCreateJobRequest struct {
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

func (s *Server) handleStudioCreateJob(w httpstd.ResponseWriter, r *httpstd.Request) {
	claims, ok := s.authClaims(w, r)
	if !ok {
		return
	}
	idemKey, ok := s.idempotencyKey(w, r)
	if !ok {
		return
	}

	var req studioCreateJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, httpstd.StatusBadRequest, "invalid_json", nil)
		return
	}
	if strings.TrimSpace(req.Prompt) == "" {
		writeError(w, httpstd.StatusBadRequest, "prompt_required", nil)
		return
	}
	if strings.TrimSpace(req.AppName) == "" {
		req.AppName = "Generated App"
	}

	s.withIdempotency(r.Context(), w, claims.TenantID, r.URL.Path, idemKey, func() (int, []byte, error) {
		job := s.studio.CreateJob(claims.TenantID, studio.Confirmation{
			Prompt:           req.Prompt,
			AppName:          req.AppName,
			Domain:           req.Domain,
			Template:         req.Template,
			SourceSystem:     req.SourceSystem,
			PrimaryUsers:     normalizeList(req.PrimaryUsers),
			CoreWorkflows:    normalizeList(req.CoreWorkflows),
			DataEntities:     normalizeList(req.DataEntities),
			DeploymentTarget: req.DeploymentTarget,
			Region:           req.Region,
			Plan:             req.Plan,
			GenerationDepth:  req.GenerationDepth,
			Integrations:     normalizeList(req.Integrations),
			Constraints:      normalizeList(req.Constraints),
		})
		payload, err := json.Marshal(job)
		if err != nil {
			return 0, nil, err
		}
		return httpstd.StatusCreated, payload, nil
	})
}

func (s *Server) handleStudioGetJob(w httpstd.ResponseWriter, r *httpstd.Request) {
	claims, ok := s.authClaims(w, r)
	if !ok {
		return
	}
	jobID := strings.TrimSpace(r.PathValue("id"))
	if jobID == "" {
		writeError(w, httpstd.StatusBadRequest, "job_id_required", nil)
		return
	}
	job, found := s.studio.GetJob(claims.TenantID, jobID)
	if !found {
		writeError(w, httpstd.StatusNotFound, "job_not_found", nil)
		return
	}
	writeJSONValue(w, httpstd.StatusOK, job)
}

func (s *Server) handleStudioArtifacts(w httpstd.ResponseWriter, r *httpstd.Request) {
	claims, ok := s.authClaims(w, r)
	if !ok {
		return
	}
	jobID := strings.TrimSpace(r.PathValue("id"))
	if jobID == "" {
		writeError(w, httpstd.StatusBadRequest, "job_id_required", nil)
		return
	}
	manifest, found := s.studio.GetArtifacts(claims.TenantID, jobID)
	if !found {
		writeError(w, httpstd.StatusNotFound, "job_not_found", nil)
		return
	}
	writeJSONValue(w, httpstd.StatusOK, manifest)
}

type studioRunRequest struct {
	Target string `json:"target"`
}

func (s *Server) handleStudioRun(w httpstd.ResponseWriter, r *httpstd.Request) {
	claims, ok := s.authClaims(w, r)
	if !ok {
		return
	}
	idemKey, ok := s.idempotencyKey(w, r)
	if !ok {
		return
	}
	jobID := strings.TrimSpace(r.PathValue("id"))
	if jobID == "" {
		writeError(w, httpstd.StatusBadRequest, "job_id_required", nil)
		return
	}

	var req studioRunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, httpstd.StatusBadRequest, "invalid_json", nil)
		return
	}
	if strings.TrimSpace(req.Target) == "" {
		req.Target = "all"
	}

	s.withIdempotency(r.Context(), w, claims.TenantID, r.URL.Path, idemKey, func() (int, []byte, error) {
		result, found := s.studio.RunTarget(claims.TenantID, jobID, req.Target)
		if !found {
			body, _ := json.Marshal(map[string]any{"error": "job_not_found"})
			return httpstd.StatusNotFound, body, nil
		}
		payload, err := json.Marshal(result)
		if err != nil {
			return 0, nil, err
		}
		return httpstd.StatusOK, payload, nil
	})
}

func (s *Server) handleStudioVerification(w httpstd.ResponseWriter, r *httpstd.Request) {
	claims, ok := s.authClaims(w, r)
	if !ok {
		return
	}
	jobID := strings.TrimSpace(r.PathValue("id"))
	if jobID == "" {
		writeError(w, httpstd.StatusBadRequest, "job_id_required", nil)
		return
	}
	report, found := s.studio.GetVerification(claims.TenantID, jobID)
	if !found {
		writeError(w, httpstd.StatusNotFound, "job_not_found", nil)
		return
	}
	writeJSONValue(w, httpstd.StatusOK, report)
}

func (s *Server) handleStudioJTBD(w httpstd.ResponseWriter, r *httpstd.Request) {
	claims, ok := s.authClaims(w, r)
	if !ok {
		return
	}
	jobID := strings.TrimSpace(r.PathValue("id"))
	if jobID == "" {
		writeError(w, httpstd.StatusBadRequest, "job_id_required", nil)
		return
	}
	coverage, found := s.studio.GetJTBDCoverage(claims.TenantID, jobID)
	if !found {
		writeError(w, httpstd.StatusNotFound, "job_not_found", nil)
		return
	}
	writeJSONValue(w, httpstd.StatusOK, map[string]any{"jtbd_coverage": coverage})
}

func (s *Server) handleStudioBundle(w httpstd.ResponseWriter, r *httpstd.Request) {
	claims, ok := s.authClaimsForStreamOrPreview(w, r)
	if !ok {
		return
	}
	jobID := strings.TrimSpace(r.PathValue("id"))
	if jobID == "" {
		writeError(w, httpstd.StatusBadRequest, "job_id_required", nil)
		return
	}
	filename, payload, found, err := s.studio.BuildBundle(claims.TenantID, jobID)
	if err != nil {
		writeError(w, httpstd.StatusInternalServerError, "bundle_build_failed", map[string]any{"details": err.Error()})
		return
	}
	if !found {
		writeError(w, httpstd.StatusNotFound, "job_not_found", nil)
		return
	}
	w.Header().Set("Content-Type", "application/gzip")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(httpstd.StatusOK)
	_, _ = w.Write(payload)
}

type studioTerminalRequest struct {
	Command string `json:"command"`
}

func (s *Server) handleStudioTerminal(w httpstd.ResponseWriter, r *httpstd.Request) {
	claims, ok := s.authClaims(w, r)
	if !ok {
		return
	}
	idemKey, ok := s.idempotencyKey(w, r)
	if !ok {
		return
	}
	jobID := strings.TrimSpace(r.PathValue("id"))
	if jobID == "" {
		writeError(w, httpstd.StatusBadRequest, "job_id_required", nil)
		return
	}
	var req studioTerminalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, httpstd.StatusBadRequest, "invalid_json", nil)
		return
	}
	if strings.TrimSpace(req.Command) == "" {
		writeError(w, httpstd.StatusBadRequest, "command_required", nil)
		return
	}

	s.withIdempotency(r.Context(), w, claims.TenantID, r.URL.Path, idemKey, func() (int, []byte, error) {
		result, found := s.studio.RunTerminal(claims.TenantID, jobID, req.Command)
		if !found {
			body, _ := json.Marshal(map[string]any{"error": "job_not_found"})
			return httpstd.StatusNotFound, body, nil
		}
		payload, err := json.Marshal(result)
		if err != nil {
			return 0, nil, err
		}
		return httpstd.StatusOK, payload, nil
	})
}

func (s *Server) handleStudioConsole(w httpstd.ResponseWriter, r *httpstd.Request) {
	claims, ok := s.authClaims(w, r)
	if !ok {
		return
	}
	jobID := strings.TrimSpace(r.PathValue("id"))
	if jobID == "" {
		writeError(w, httpstd.StatusBadRequest, "job_id_required", nil)
		return
	}
	logs, found := s.studio.GetConsole(claims.TenantID, jobID)
	if !found {
		writeError(w, httpstd.StatusNotFound, "job_not_found", nil)
		return
	}
	writeJSONValue(w, httpstd.StatusOK, map[string]any{"logs": logs})
}

func (s *Server) handleStudioPreview(w httpstd.ResponseWriter, r *httpstd.Request) {
	claims, ok := s.authClaimsForStreamOrPreview(w, r)
	if !ok {
		return
	}
	jobID := strings.TrimSpace(r.PathValue("id"))
	if jobID == "" {
		writeError(w, httpstd.StatusBadRequest, "job_id_required", nil)
		return
	}
	client := strings.TrimSpace(r.URL.Query().Get("client"))
	token := strings.TrimSpace(r.URL.Query().Get("token"))
	html, found := s.studio.RenderPreview(claims.TenantID, jobID, client, token)
	if !found {
		writeError(w, httpstd.StatusNotFound, "job_not_found", nil)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(httpstd.StatusOK)
	_, _ = w.Write([]byte(html))
}

func (s *Server) handleStudioRuntimeAsset(w httpstd.ResponseWriter, r *httpstd.Request) {
	claims, ok := s.authClaimsForStreamOrPreview(w, r)
	if !ok {
		return
	}
	jobID := strings.TrimSpace(r.PathValue("id"))
	if jobID == "" {
		writeError(w, httpstd.StatusBadRequest, "job_id_required", nil)
		return
	}
	client := strings.TrimSpace(r.PathValue("client"))
	asset := strings.TrimSpace(r.PathValue("asset"))
	contentType, body, found := s.studio.RenderRuntimeAsset(claims.TenantID, jobID, client, asset)
	if !found {
		writeError(w, httpstd.StatusNotFound, "runtime_asset_not_found", nil)
		return
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(httpstd.StatusOK)
	_, _ = w.Write(body)
}

func (s *Server) handleStudioEvents(w httpstd.ResponseWriter, r *httpstd.Request) {
	claims, ok := s.authClaimsForStreamOrPreview(w, r)
	if !ok {
		return
	}
	jobID := strings.TrimSpace(r.PathValue("id"))
	if jobID == "" {
		writeError(w, httpstd.StatusBadRequest, "job_id_required", nil)
		return
	}

	flusher, ok := w.(httpstd.Flusher)
	if !ok {
		writeError(w, httpstd.StatusInternalServerError, "stream_unsupported", nil)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(httpstd.StatusOK)
	_, _ = w.Write([]byte("retry: 1000\n\n"))
	flusher.Flush()

	job, found := s.studio.GetJob(claims.TenantID, jobID)
	if !found {
		writeSSEEvent(w, flusher, "error", map[string]any{"error": "job_not_found"})
		return
	}
	lastRevision := jobRevision(job)
	if !writeSSEEvent(w, flusher, "job", job) {
		return
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			current, ok := s.studio.GetJob(claims.TenantID, jobID)
			if !ok {
				writeSSEEvent(w, flusher, "error", map[string]any{"error": "job_not_found"})
				return
			}
			rev := jobRevision(current)
			if rev != lastRevision {
				if !writeSSEEvent(w, flusher, "job", current) {
					return
				}
				lastRevision = rev
				continue
			}
			if _, err := w.Write([]byte(": keepalive\n\n")); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

func (s *Server) authClaimsForStreamOrPreview(w httpstd.ResponseWriter, r *httpstd.Request) (auth.Claims, bool) {
	authorization := strings.TrimSpace(r.Header.Get("Authorization"))
	if authorization != "" {
		return s.authClaims(w, r)
	}
	token := strings.TrimSpace(r.URL.Query().Get("token"))
	if token == "" {
		writeError(w, httpstd.StatusUnauthorized, auth.ErrMissingAuthHeader.Error(), nil)
		return auth.Claims{}, false
	}
	claims, err := s.auth.Authenticate("Bearer " + token)
	if err == nil {
		return claims, true
	}
	status := httpstd.StatusUnauthorized
	if err == auth.ErrInvalidAuthScheme {
		status = httpstd.StatusBadRequest
	}
	writeError(w, status, err.Error(), nil)
	return auth.Claims{}, false
}

func writeSSEEvent(w httpstd.ResponseWriter, flusher httpstd.Flusher, name string, payload any) bool {
	data, err := json.Marshal(payload)
	if err != nil {
		return false
	}
	if _, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", name, data); err != nil {
		return false
	}
	flusher.Flush()
	return true
}

func jobRevision(job studio.Job) string {
	return fmt.Sprintf(
		"%s|%s|%d|%d",
		job.Status,
		job.UpdatedAt.Format(time.RFC3339Nano),
		len(job.TerminalLogs),
		len(job.ConsoleLogs),
	)
}

func normalizeList(in []string) []string {
	out := make([]string, 0, len(in))
	seen := map[string]struct{}{}
	for _, item := range in {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	return out
}
