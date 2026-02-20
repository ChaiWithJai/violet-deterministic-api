package runtime

import (
	"encoding/json"
	"net/http"
	"time"
)

type Spec struct {
	AppName    string   `json:"app_name"`
	Domain     string   `json:"domain"`
	Plan       string   `json:"plan"`
	Region     string   `json:"region"`
	Users      []string `json:"users"`
	Entities   []string `json:"entities"`
	Workflows  []string `json:"workflows"`
	ToolRoutes []Tool   `json:"tool_routes"`
}

type Tool struct {
	Name        string `json:"name"`
	Method      string `json:"method"`
	Path        string `json:"path"`
	Description string `json:"description"`
}

type Server struct {
	spec Spec
	mux  *http.ServeMux
}

func NewServer(spec Spec) *Server {
	s := &Server{
		spec: spec,
		mux:  http.NewServeMux(),
	}
	s.routes()
	return s
}

func (s *Server) Handler() http.Handler {
	return s.mux
}

func (s *Server) routes() {
	s.mux.HandleFunc("/health", s.handleHealth)
	s.mux.HandleFunc("/v1/tools", s.handleTools)
	s.mux.HandleFunc("/v1/workflows/execute", s.handleExecuteWorkflow)
	s.mux.HandleFunc("/v1/entities", s.handleEntities)
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":     "ok",
		"app_name":   s.spec.AppName,
		"domain":     s.spec.Domain,
		"generated":  true,
		"checked_at": time.Now().UTC().Format(time.RFC3339),
	})
}

func (s *Server) handleTools(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"tools": s.spec.ToolRoutes,
		"count": len(s.spec.ToolRoutes),
	})
}

func (s *Server) handleExecuteWorkflow(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Workflow string         `json:"workflow"`
		Input    map[string]any `json:"input"`
	}
	var req request
	_ = json.NewDecoder(r.Body).Decode(&req)
	writeJSON(w, http.StatusOK, map[string]any{
		"workflow":  req.Workflow,
		"status":    "accepted",
		"idempotent": true,
		"input":     req.Input,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (s *Server) handleEntities(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"entities": s.spec.Entities,
		"count":    len(s.spec.Entities),
	})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	payload, _ := json.Marshal(body)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(payload)
}
