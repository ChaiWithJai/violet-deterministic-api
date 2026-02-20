package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestListProvidersAndInferOllama(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/tags":
			_ = json.NewEncoder(w).Encode(map[string]any{"models": []map[string]any{{"name": "glm-4.7"}, {"name": "qwen2.5-coder:7b"}}})
		case "/api/generate":
			_ = json.NewEncoder(w).Encode(map[string]any{"response": "generated local response", "prompt_eval_count": 11, "eval_count": 24})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	svc := NewService(Config{
		DefaultProvider:    "ollama",
		DefaultModel:       "glm-4.7",
		Timeout:            2 * time.Second,
		OllamaBaseURL:      server.URL,
		OllamaDefaultModel: "glm-4.7",
	})

	providers := svc.ListProviders(context.Background())
	if len(providers) != 2 {
		t.Fatalf("expected 2 providers, got %d", len(providers))
	}
	if !providers[0].Reachable {
		t.Fatalf("expected ollama to be reachable, got error=%q", providers[0].Error)
	}
	if providers[0].Name != "ollama" {
		t.Fatalf("expected first provider=ollama, got %q", providers[0].Name)
	}

	resp, err := svc.Infer(context.Background(), InferRequest{Prompt: "hello", Provider: "ollama"})
	if err != nil {
		t.Fatalf("infer error: %v", err)
	}
	if resp.Text != "generated local response" {
		t.Fatalf("unexpected infer response text: %q", resp.Text)
	}
	if resp.Provider != "ollama" {
		t.Fatalf("unexpected provider: %q", resp.Provider)
	}
	if resp.Model != "glm-4.7" {
		t.Fatalf("unexpected model: %q", resp.Model)
	}
}

func TestInferFrontier(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/models":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{{"id": "gpt-4o-mini"}}})
		case "/chat/completions":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"choices": []map[string]any{{"message": map[string]any{"content": "frontier response"}}},
				"usage":   map[string]any{"prompt_tokens": 5, "completion_tokens": 7},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	svc := NewService(Config{
		DefaultProvider:      "frontier",
		DefaultModel:         "gpt-4o-mini",
		Timeout:              2 * time.Second,
		FrontierBaseURL:      server.URL,
		FrontierAPIKey:       "demo-key",
		FrontierDefaultModel: "gpt-4o-mini",
		OllamaBaseURL:        "http://127.0.0.1:1",
		OllamaDefaultModel:   "glm-4.7",
	})

	providers := svc.ListProviders(context.Background())
	if len(providers) != 2 {
		t.Fatalf("expected 2 providers, got %d", len(providers))
	}
	if !providers[1].Reachable {
		t.Fatalf("expected frontier reachable, got error=%q", providers[1].Error)
	}

	resp, err := svc.Infer(context.Background(), InferRequest{Prompt: "hello", Provider: "frontier", Model: "gpt-4o-mini"})
	if err != nil {
		t.Fatalf("infer error: %v", err)
	}
	if resp.Text != "frontier response" {
		t.Fatalf("unexpected response text: %q", resp.Text)
	}
	if resp.SourceMode != "frontier" {
		t.Fatalf("expected source mode frontier, got %q", resp.SourceMode)
	}
}
