package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Config struct {
	DefaultProvider string
	DefaultModel    string
	Timeout         time.Duration

	OllamaBaseURL      string
	OllamaDefaultModel string

	FrontierBaseURL      string
	FrontierAPIKey       string
	FrontierDefaultModel string
}

type Service struct {
	cfg    Config
	client *http.Client
}

type ProviderInfo struct {
	Name         string   `json:"name"`
	Kind         string   `json:"kind"`
	BaseURL      string   `json:"base_url"`
	DefaultModel string   `json:"default_model"`
	Reachable    bool     `json:"reachable"`
	Models       []string `json:"models"`
	Error        string   `json:"error,omitempty"`
}

type InferRequest struct {
	Provider    string  `json:"provider"`
	Model       string  `json:"model"`
	Prompt      string  `json:"prompt"`
	System      string  `json:"system,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
	MaxTokens   int     `json:"max_tokens,omitempty"`
}

type InferResponse struct {
	Provider   string         `json:"provider"`
	Model      string         `json:"model"`
	Text       string         `json:"text"`
	LatencyMS  int64          `json:"latency_ms"`
	Usage      map[string]any `json:"usage,omitempty"`
	Raw        map[string]any `json:"raw,omitempty"`
	Generated  time.Time      `json:"generated_at"`
	SourceMode string         `json:"source_mode"`
}

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if strings.TrimSpace(e.Message) != "" {
		return e.Message
	}
	return e.Code
}

func NewService(cfg Config) *Service {
	if strings.TrimSpace(cfg.DefaultProvider) == "" {
		cfg.DefaultProvider = "ollama"
	}
	if strings.TrimSpace(cfg.DefaultModel) == "" {
		cfg.DefaultModel = "glm-4.7-flash:latest"
	}
	if strings.TrimSpace(cfg.OllamaBaseURL) == "" {
		cfg.OllamaBaseURL = "http://host.docker.internal:11434"
	}
	if strings.TrimSpace(cfg.OllamaDefaultModel) == "" {
		cfg.OllamaDefaultModel = cfg.DefaultModel
	}
	if strings.TrimSpace(cfg.FrontierBaseURL) == "" {
		cfg.FrontierBaseURL = "http://host.docker.internal:11434/v1"
	}
	if strings.TrimSpace(cfg.FrontierDefaultModel) == "" {
		cfg.FrontierDefaultModel = cfg.DefaultModel
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 45 * time.Second
	}
	return &Service{
		cfg: cfg,
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

func (s *Service) ListProviders(ctx context.Context) []ProviderInfo {
	providers := []ProviderInfo{
		{Name: "ollama", Kind: "local", BaseURL: strings.TrimSpace(s.cfg.OllamaBaseURL), DefaultModel: strings.TrimSpace(s.cfg.OllamaDefaultModel)},
		{Name: "frontier", Kind: "remote", BaseURL: strings.TrimSpace(s.cfg.FrontierBaseURL), DefaultModel: strings.TrimSpace(s.cfg.FrontierDefaultModel)},
	}
	for i := range providers {
		info := &providers[i]
		models, err := s.listModelsForProvider(ctx, info.Name)
		if err != nil {
			info.Reachable = false
			info.Error = err.Error()
			continue
		}
		info.Reachable = true
		info.Models = models
	}
	return providers
}

func (s *Service) Infer(ctx context.Context, req InferRequest) (InferResponse, error) {
	provider := strings.ToLower(strings.TrimSpace(req.Provider))
	if provider == "" {
		provider = strings.ToLower(strings.TrimSpace(s.cfg.DefaultProvider))
	}
	if provider == "" {
		provider = "ollama"
	}
	if strings.TrimSpace(req.Prompt) == "" {
		return InferResponse{}, &Error{Code: "prompt_required", Message: "prompt is required"}
	}
	model := strings.TrimSpace(req.Model)
	if model == "" {
		model = s.defaultModelForProvider(provider)
	}
	if model == "" {
		model = strings.TrimSpace(s.cfg.DefaultModel)
	}
	if model == "" {
		return InferResponse{}, &Error{Code: "model_required", Message: "model is required"}
	}

	start := time.Now().UTC()
	var (
		text       string
		usage      map[string]any
		raw        map[string]any
		sourceMode string
		err        error
	)
	if provider == "ollama" {
		text, usage, raw, err = s.inferOllama(ctx, model, req)
		sourceMode = "local"
	} else if provider == "frontier" {
		text, usage, raw, err = s.inferFrontier(ctx, model, req)
		sourceMode = "frontier"
	} else {
		return InferResponse{}, &Error{Code: "unknown_provider", Message: fmt.Sprintf("unsupported provider: %s", provider)}
	}
	if err != nil {
		var typed *Error
		if errors.As(err, &typed) {
			return InferResponse{}, typed
		}
		return InferResponse{}, &Error{Code: "provider_error", Message: err.Error()}
	}

	return InferResponse{
		Provider:   provider,
		Model:      model,
		Text:       text,
		LatencyMS:  time.Since(start).Milliseconds(),
		Usage:      usage,
		Raw:        raw,
		Generated:  time.Now().UTC(),
		SourceMode: sourceMode,
	}, nil
}

func (s *Service) defaultModelForProvider(provider string) string {
	switch provider {
	case "ollama":
		return strings.TrimSpace(s.cfg.OllamaDefaultModel)
	case "frontier":
		return strings.TrimSpace(s.cfg.FrontierDefaultModel)
	default:
		return strings.TrimSpace(s.cfg.DefaultModel)
	}
}

func (s *Service) listModelsForProvider(ctx context.Context, provider string) ([]string, error) {
	switch provider {
	case "ollama":
		return s.listOllamaModels(ctx)
	case "frontier":
		return s.listFrontierModels(ctx)
	default:
		return nil, &Error{Code: "unknown_provider", Message: fmt.Sprintf("unsupported provider: %s", provider)}
	}
}

func (s *Service) listOllamaModels(ctx context.Context) ([]string, error) {
	url := joinURL(s.cfg.OllamaBaseURL, "/api/tags")
	respBody, status, err := s.doJSON(ctx, http.MethodGet, url, nil, nil)
	if err != nil {
		return nil, &Error{Code: "ollama_unreachable", Message: err.Error()}
	}
	if status < 200 || status >= 300 {
		return nil, &Error{Code: "ollama_unreachable", Message: fmt.Sprintf("ollama returned %d", status)}
	}
	var payload struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	if err := json.Unmarshal(respBody, &payload); err != nil {
		return nil, &Error{Code: "ollama_decode_failed", Message: err.Error()}
	}
	models := make([]string, 0, len(payload.Models))
	for _, item := range payload.Models {
		name := strings.TrimSpace(item.Name)
		if name == "" {
			continue
		}
		models = append(models, name)
	}
	return models, nil
}

func (s *Service) inferOllama(ctx context.Context, model string, req InferRequest) (string, map[string]any, map[string]any, error) {
	url := joinURL(s.cfg.OllamaBaseURL, "/api/generate")
	body := map[string]any{
		"model":  model,
		"prompt": req.Prompt,
		"stream": false,
	}
	if strings.TrimSpace(req.System) != "" {
		body["system"] = strings.TrimSpace(req.System)
	}
	options := map[string]any{}
	if req.Temperature > 0 {
		options["temperature"] = req.Temperature
	}
	if req.MaxTokens > 0 {
		options["num_predict"] = req.MaxTokens
	}
	if len(options) > 0 {
		body["options"] = options
	}
	respBody, status, err := s.doJSON(ctx, http.MethodPost, url, nil, body)
	if err != nil {
		return "", nil, nil, &Error{Code: "ollama_unreachable", Message: err.Error()}
	}
	if status < 200 || status >= 300 {
		return "", nil, nil, &Error{Code: "ollama_infer_failed", Message: fmt.Sprintf("ollama returned %d", status)}
	}
	var payload map[string]any
	if err := json.Unmarshal(respBody, &payload); err != nil {
		return "", nil, nil, &Error{Code: "ollama_decode_failed", Message: err.Error()}
	}
	text, _ := payload["response"].(string)
	usage := map[string]any{}
	if v, ok := payload["prompt_eval_count"]; ok {
		usage["prompt_tokens"] = v
	}
	if v, ok := payload["eval_count"]; ok {
		usage["completion_tokens"] = v
	}
	if len(usage) == 0 {
		usage = nil
	}
	return text, usage, payload, nil
}

func (s *Service) listFrontierModels(ctx context.Context) ([]string, error) {
	url := joinURL(s.cfg.FrontierBaseURL, "/models")
	headers := s.frontierHeaders()
	respBody, status, err := s.doJSON(ctx, http.MethodGet, url, headers, nil)
	if err != nil {
		return nil, &Error{Code: "frontier_unreachable", Message: err.Error()}
	}
	if status < 200 || status >= 300 {
		if status == http.StatusUnauthorized && strings.TrimSpace(s.cfg.FrontierAPIKey) == "" {
			return nil, &Error{Code: "frontier_auth_required", Message: "frontier endpoint requires API key; set FRONTIER_API_KEY"}
		}
		return nil, &Error{Code: "frontier_unreachable", Message: fmt.Sprintf("frontier returned %d", status)}
	}
	var payload struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &payload); err != nil {
		return nil, &Error{Code: "frontier_decode_failed", Message: err.Error()}
	}
	models := make([]string, 0, len(payload.Data))
	for _, item := range payload.Data {
		id := strings.TrimSpace(item.ID)
		if id == "" {
			continue
		}
		models = append(models, id)
	}
	return models, nil
}

func (s *Service) inferFrontier(ctx context.Context, model string, req InferRequest) (string, map[string]any, map[string]any, error) {
	url := joinURL(s.cfg.FrontierBaseURL, "/chat/completions")
	headers := s.frontierHeaders()
	messages := make([]map[string]string, 0, 2)
	if strings.TrimSpace(req.System) != "" {
		messages = append(messages, map[string]string{"role": "system", "content": strings.TrimSpace(req.System)})
	}
	messages = append(messages, map[string]string{"role": "user", "content": req.Prompt})
	body := map[string]any{
		"model":    model,
		"messages": messages,
	}
	if req.Temperature > 0 {
		body["temperature"] = req.Temperature
	}
	if req.MaxTokens > 0 {
		body["max_tokens"] = req.MaxTokens
	}
	respBody, status, err := s.doJSON(ctx, http.MethodPost, url, headers, body)
	if err != nil {
		return "", nil, nil, &Error{Code: "frontier_unreachable", Message: err.Error()}
	}
	if status < 200 || status >= 300 {
		if status == http.StatusUnauthorized && strings.TrimSpace(s.cfg.FrontierAPIKey) == "" {
			return "", nil, nil, &Error{Code: "frontier_auth_required", Message: "frontier endpoint requires API key; set FRONTIER_API_KEY"}
		}
		return "", nil, nil, &Error{Code: "frontier_infer_failed", Message: fmt.Sprintf("frontier returned %d", status)}
	}
	var payload map[string]any
	if err := json.Unmarshal(respBody, &payload); err != nil {
		return "", nil, nil, &Error{Code: "frontier_decode_failed", Message: err.Error()}
	}
	text := ""
	if choices, ok := payload["choices"].([]any); ok && len(choices) > 0 {
		if first, ok := choices[0].(map[string]any); ok {
			if message, ok := first["message"].(map[string]any); ok {
				if content, ok := message["content"].(string); ok {
					text = content
				}
			}
		}
	}
	usage, _ := payload["usage"].(map[string]any)
	return text, usage, payload, nil
}

func (s *Service) doJSON(ctx context.Context, method, endpoint string, headers map[string]string, body any) ([]byte, int, error) {
	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return nil, 0, err
		}
		reader = bytes.NewReader(payload)
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, reader)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return respBody, resp.StatusCode, nil
}

func joinURL(base, path string) string {
	base = strings.TrimRight(strings.TrimSpace(base), "/")
	path = "/" + strings.TrimLeft(strings.TrimSpace(path), "/")
	return base + path
}

func (s *Service) frontierHeaders() map[string]string {
	headers := map[string]string{}
	apiKey := strings.TrimSpace(s.cfg.FrontierAPIKey)
	if apiKey != "" {
		headers["Authorization"] = "Bearer " + apiKey
	}
	return headers
}
