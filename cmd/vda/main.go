package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	switch os.Args[1] {
	case "llm":
		handleLLM(os.Args[2:])
	case "tools":
		handleTools(os.Args[2:])
	case "studio":
		handleStudio(os.Args[2:])
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "vda CLI")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  vda llm providers [--base-url URL] [--token TOKEN]")
	fmt.Fprintln(os.Stderr, "  vda llm infer --prompt TEXT [--provider ollama|frontier] [--model MODEL] [--system TEXT] [--temperature N] [--max-tokens N] [--base-url URL] [--token TOKEN]")
	fmt.Fprintln(os.Stderr, "  vda tools list [--base-url URL] [--token TOKEN]")
	fmt.Fprintln(os.Stderr, "  vda studio launch --job-id JOB_ID [--base-url URL] [--token TOKEN] [--out-dir DIR] [--api-port N] [--web-port N] [--mobile-port N]")
}

func handleLLM(args []string) {
	if len(args) < 1 {
		usage()
		os.Exit(2)
	}
	cmd := args[0]
	switch cmd {
	case "providers":
		fs := flag.NewFlagSet("providers", flag.ExitOnError)
		baseURL := fs.String("base-url", getenv("VDA_BASE_URL", "http://localhost:4020"), "API base URL")
		token := fs.String("token", getenv("VDA_TOKEN", "dev-token"), "bearer token")
		_ = fs.Parse(args[1:])
		resp, err := doRequest(http.MethodGet, strings.TrimRight(*baseURL, "/")+"/v1/llm/providers", *token, "", nil)
		must(err)
		printJSON(resp)
	case "infer":
		fs := flag.NewFlagSet("infer", flag.ExitOnError)
		baseURL := fs.String("base-url", getenv("VDA_BASE_URL", "http://localhost:4020"), "API base URL")
		token := fs.String("token", getenv("VDA_TOKEN", "dev-token"), "bearer token")
		provider := fs.String("provider", "", "provider name (ollama|frontier)")
		model := fs.String("model", "", "model id")
		prompt := fs.String("prompt", "", "prompt text")
		system := fs.String("system", "", "system prompt")
		temperature := fs.Float64("temperature", 0, "temperature")
		maxTokens := fs.Int("max-tokens", 0, "max tokens")
		_ = fs.Parse(args[1:])
		if strings.TrimSpace(*prompt) == "" {
			must(fmt.Errorf("--prompt is required"))
		}
		body := map[string]any{
			"provider":    *provider,
			"model":       *model,
			"prompt":      *prompt,
			"system":      *system,
			"temperature": *temperature,
			"max_tokens":  *maxTokens,
		}
		idem := fmt.Sprintf("cli-llm-infer-%d", time.Now().UTC().UnixNano())
		resp, err := doRequest(http.MethodPost, strings.TrimRight(*baseURL, "/")+"/v1/llm/infer", *token, idem, body)
		must(err)
		printJSON(resp)
	default:
		usage()
		os.Exit(2)
	}
}

func handleTools(args []string) {
	if len(args) < 1 || args[0] != "list" {
		usage()
		os.Exit(2)
	}
	fs := flag.NewFlagSet("tools list", flag.ExitOnError)
	baseURL := fs.String("base-url", getenv("VDA_BASE_URL", "http://localhost:4020"), "API base URL")
	token := fs.String("token", getenv("VDA_TOKEN", "dev-token"), "bearer token")
	_ = fs.Parse(args[1:])
	resp, err := doRequest(http.MethodGet, strings.TrimRight(*baseURL, "/")+"/v1/tools", *token, "", nil)
	must(err)
	printJSON(resp)
}

func doRequest(method, endpoint, token, idem string, body any) (map[string]any, error) {
	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(payload)
	}
	req, err := http.NewRequest(method, endpoint, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if strings.TrimSpace(token) != "" {
		req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(token))
	}
	if strings.TrimSpace(idem) != "" {
		req.Header.Set("Idempotency-Key", idem)
	}
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	out := map[string]any{"status": resp.StatusCode}
	if len(respBytes) == 0 {
		out["body"] = map[string]any{}
		return out, nil
	}
	var payload any
	if err := json.Unmarshal(respBytes, &payload); err != nil {
		out["body"] = string(respBytes)
		return out, nil
	}
	out["body"] = payload
	return out, nil
}

func printJSON(v any) {
	b, err := json.MarshalIndent(v, "", "  ")
	must(err)
	fmt.Println(string(b))
}

func must(err error) {
	if err == nil {
		return
	}
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
