package http

import (
	httpstd "net/http"
)

func (s *Server) handleToolsCatalog(w httpstd.ResponseWriter, r *httpstd.Request) {
	if _, ok := s.authClaims(w, r); !ok {
		return
	}
	writeJSONValue(w, httpstd.StatusOK, map[string]any{
		"tools": []map[string]any{
			{
				"name":        "agent.plan",
				"description": "Create deterministic app blueprint from prompt",
				"method":      "POST",
				"path":        "/v1/agents/plan",
				"cli":         "vda tools list",
			},
			{
				"name":        "agent.clarify",
				"description": "Run structured clarification loop and return targeted follow-up questions",
				"method":      "POST",
				"path":        "/v1/agents/clarify",
				"cli":         "curl -X POST /v1/agents/clarify",
			},
			{
				"name":        "agent.act",
				"description": "Apply one policy-checked mutation",
				"method":      "POST",
				"path":        "/v1/agents/act",
				"cli":         "curl -X POST /v1/agents/act",
			},
			{
				"name":        "agent.verify",
				"description": "Run machine-readable verification checks",
				"method":      "POST",
				"path":        "/v1/agents/verify",
				"cli":         "curl -X POST /v1/agents/verify",
			},
			{
				"name":        "agent.deploy",
				"description": "Request self-host or managed deploy intent",
				"method":      "POST",
				"path":        "/v1/agents/deploy",
				"cli":         "curl -X POST /v1/agents/deploy",
			},
			{
				"name":        "llm.providers",
				"description": "List configured model providers with health and models",
				"method":      "GET",
				"path":        "/v1/llm/providers",
				"cli":         "vda llm providers --token <token>",
			},
			{
				"name":        "llm.infer",
				"description": "Run one model call against local or frontier provider",
				"method":      "POST",
				"path":        "/v1/llm/infer",
				"cli":         "vda llm infer --provider ollama --model glm-4.7 --prompt '...'",
			},
			{
				"name":        "studio.launch",
				"description": "Download generated app bundle and launch api/web/mobile locally in one command",
				"method":      "GET",
				"path":        "/v1/studio/jobs/{id}/bundle",
				"cli":         "vda studio launch --job-id <job_id>",
			},
		},
	})
}
