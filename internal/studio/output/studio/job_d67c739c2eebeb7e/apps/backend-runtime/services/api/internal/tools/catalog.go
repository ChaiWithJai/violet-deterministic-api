package tools

import "generated/backend-runtime/services/api/internal/runtime"

func Catalog() []runtime.Tool {
	return []runtime.Tool{
		{
			Name:        "plan",
			Method:      "POST",
			Path:        "/v1/agents/plan",
			Description: "Plan structured work for the tenant",
		},
		{
			Name:        "act",
			Method:      "POST",
			Path:        "/v1/agents/act",
			Description: "Execute an idempotent mutation",
		},
		{
			Name:        "verify",
			Method:      "POST",
			Path:        "/v1/agents/verify",
			Description: "Verify outputs before deployment",
		},
		{
			Name:        "deploy",
			Method:      "POST",
			Path:        "/v1/agents/deploy",
			Description: "Prepare self-host or managed deploy intent",
		},
	}
}
