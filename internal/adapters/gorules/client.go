package gorules

import "context"

// Client defines an adapter seam for GoRules policy evaluation.
// Response should include policy_version and a deterministic decision trace id.
type Client interface {
	Evaluate(ctx context.Context, tenantID string, input map[string]any) (map[string]any, error)
}
