package http

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	httpstd "net/http"
	"strings"
	"time"

	"github.com/restarone/violet-deterministic-api/internal/adapters/gorse"
	"github.com/restarone/violet-deterministic-api/internal/adapters/gorules"
	"github.com/restarone/violet-deterministic-api/internal/auth"
	"github.com/restarone/violet-deterministic-api/internal/config"
	"github.com/restarone/violet-deterministic-api/internal/decision"
	"github.com/restarone/violet-deterministic-api/internal/llm"
	"github.com/restarone/violet-deterministic-api/internal/storage"
	"github.com/restarone/violet-deterministic-api/internal/studio"
)

type Server struct {
	cfg    config.Config
	http   *httpstd.Server
	engine *decision.Engine
	store  *storage.Store
	auth   *auth.Authenticator
	policy gorules.Client
	studio *studio.Service
	llm    *llm.Service

	cleanupCtx    context.Context
	cleanupCancel context.CancelFunc
}

func NewServer(cfg config.Config) (*Server, error) {
	ctx, cancel := context.WithCancel(context.Background())
	store, err := storage.New(ctx, cfg.DatabaseURL, cfg.IdempotencyTTLSeconds, cfg.IdempotencyCleanupSeconds)
	if err != nil {
		cancel()
		return nil, err
	}
	store.StartIdempotencyCleanup(ctx)

	gorseClient := gorse.NewHTTPClient(cfg.GorseBaseURL, cfg.GorseAPIKey)
	policyClient := gorules.NewLocalClient(cfg.PolicyVersion)

	s := &Server{
		cfg:    cfg,
		engine: decision.NewEngine(cfg.PolicyVersion, cfg.DataVersion, gorseClient, policyClient),
		store:  store,
		auth:   auth.New(cfg.AuthTokens),
		policy: policyClient,
		studio: studio.NewService(studio.WithPersistence(store)),
		llm: llm.NewService(llm.Config{
			DefaultProvider:      cfg.LLMDefaultProvider,
			DefaultModel:         cfg.LLMDefaultModel,
			Timeout:              time.Duration(cfg.LLMRequestTimeoutSecond) * time.Second,
			OllamaBaseURL:        cfg.OllamaBaseURL,
			OllamaDefaultModel:   cfg.OllamaDefaultModel,
			FrontierBaseURL:      cfg.FrontierBaseURL,
			FrontierAPIKey:       cfg.FrontierAPIKey,
			FrontierDefaultModel: cfg.FrontierDefaultModel,
		}),
		cleanupCtx:    ctx,
		cleanupCancel: cancel,
	}

	mux := httpstd.NewServeMux()
	mux.Handle("GET /ui/", s.uiHandler())
	mux.HandleFunc("GET /", s.handleUIRoot)

	mux.HandleFunc("GET /v1/health", s.handleHealth)
	mux.HandleFunc("POST /v1/decisions", s.handleDecisions)
	mux.HandleFunc("POST /v1/replay", s.handleReplay)
	mux.HandleFunc("POST /v1/feedback", s.handleFeedback)

	mux.HandleFunc("POST /v1/apps", s.handleCreateApp)
	mux.HandleFunc("GET /v1/apps/{id}", s.handleGetApp)
	mux.HandleFunc("PATCH /v1/apps/{id}", s.handlePatchApp)
	mux.HandleFunc("POST /v1/apps/{id}/mutations", s.handleAppMutation)
	mux.HandleFunc("POST /v1/apps/{id}/verify", s.handleVerifyApp)
	mux.HandleFunc("POST /v1/apps/{id}/deploy-intents/self-host", s.handleDeploySelfHost)
	mux.HandleFunc("POST /v1/apps/{id}/deploy-intents/managed", s.handleDeployManaged)

	mux.HandleFunc("POST /v1/agents/plan", s.handleAgentPlan)
	mux.HandleFunc("POST /v1/agents/clarify", s.handleAgentClarify)
	mux.HandleFunc("POST /v1/agents/act", s.handleAgentAct)
	mux.HandleFunc("POST /v1/agents/verify", s.handleAgentVerify)
	mux.HandleFunc("POST /v1/agents/deploy", s.handleAgentDeploy)
	mux.HandleFunc("GET /v1/llm/providers", s.handleLLMProviders)
	mux.HandleFunc("POST /v1/llm/infer", s.handleLLMInfer)
	mux.HandleFunc("GET /v1/tools", s.handleToolsCatalog)

	mux.HandleFunc("POST /v1/migration/violet/export", s.handleMigrationExport)
	mux.HandleFunc("POST /v1/migration/violet/import", s.handleMigrationImport)

	mux.HandleFunc("POST /v1/studio/jobs", s.handleStudioCreateJob)
	mux.HandleFunc("GET /v1/studio/jobs/{id}", s.handleStudioGetJob)
	mux.HandleFunc("GET /v1/studio/jobs/{id}/artifacts", s.handleStudioArtifacts)
	mux.HandleFunc("POST /v1/studio/jobs/{id}/run", s.handleStudioRun)
	mux.HandleFunc("GET /v1/studio/jobs/{id}/verification", s.handleStudioVerification)
	mux.HandleFunc("GET /v1/studio/jobs/{id}/jtbd", s.handleStudioJTBD)
	mux.HandleFunc("GET /v1/studio/jobs/{id}/bundle", s.handleStudioBundle)
	mux.HandleFunc("GET /v1/studio/jobs/{id}/preview", s.handleStudioPreview)
	mux.HandleFunc("GET /v1/studio/jobs/{id}/runtime/{client}/{asset...}", s.handleStudioRuntimeAsset)
	mux.HandleFunc("GET /v1/studio/jobs/{id}/events", s.handleStudioEvents)
	mux.HandleFunc("POST /v1/studio/jobs/{id}/terminal", s.handleStudioTerminal)
	mux.HandleFunc("GET /v1/studio/jobs/{id}/console", s.handleStudioConsole)

	s.http = &httpstd.Server{
		Addr:              ":" + cfg.Port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	return s, nil
}

func (s *Server) Start() error {
	return s.http.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.cleanupCancel()
	closeErr := s.store.Close()
	shutdownErr := s.http.Shutdown(ctx)
	if shutdownErr != nil {
		return shutdownErr
	}
	return closeErr
}

func writeJSON(w httpstd.ResponseWriter, status int, body []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}

func writeJSONValue(w httpstd.ResponseWriter, status int, v any) {
	payload, err := json.Marshal(v)
	if err != nil {
		writeJSON(w, httpstd.StatusInternalServerError, []byte(`{"error":"marshal_failed"}`))
		return
	}
	writeJSON(w, status, payload)
}

func writeError(w httpstd.ResponseWriter, status int, code string, details map[string]any) {
	payload := map[string]any{"error": code}
	for k, v := range details {
		payload[k] = v
	}
	writeJSONValue(w, status, payload)
}

func (s *Server) authClaims(w httpstd.ResponseWriter, r *httpstd.Request) (auth.Claims, bool) {
	claims, err := s.auth.Authenticate(r.Header.Get("Authorization"))
	if err == nil {
		return claims, true
	}
	status := httpstd.StatusUnauthorized
	if errors.Is(err, auth.ErrInvalidAuthScheme) {
		status = httpstd.StatusBadRequest
	}
	writeError(w, status, err.Error(), nil)
	return auth.Claims{}, false
}

func (s *Server) idempotencyKey(w httpstd.ResponseWriter, r *httpstd.Request) (string, bool) {
	key := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if key == "" {
		writeError(w, httpstd.StatusBadRequest, "missing_idempotency_key", nil)
		return "", false
	}
	return key, true
}

func (s *Server) withIdempotency(
	ctx context.Context,
	w httpstd.ResponseWriter,
	tenantID, endpoint, key string,
	exec func() (int, []byte, error),
) {
	rec, ok, err := s.store.GetIdempotency(ctx, tenantID, endpoint, key)
	if err != nil {
		writeError(w, httpstd.StatusInternalServerError, "idempotency_read_failed", map[string]any{"details": err.Error()})
		return
	}
	if ok {
		writeJSON(w, rec.StatusCode, rec.Body)
		return
	}

	status, body, err := exec()
	if err != nil {
		writeError(w, httpstd.StatusInternalServerError, "request_failed", map[string]any{"details": err.Error()})
		return
	}
	if err := s.store.PutIdempotency(ctx, tenantID, endpoint, key, status, body); err != nil {
		writeError(w, httpstd.StatusInternalServerError, "idempotency_write_failed", map[string]any{"details": err.Error()})
		return
	}
	writeJSON(w, status, body)
}

func stableID(prefix string, parts ...string) string {
	b := []byte(strings.Join(parts, "|"))
	h := sha256.Sum256(b)
	return fmt.Sprintf("%s_%s", prefix, hex.EncodeToString(h[:])[:16])
}
