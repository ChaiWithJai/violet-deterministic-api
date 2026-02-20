package gorules

import "context"

type LocalClient struct {
	PolicyVersion string
}

func NewLocalClient(policyVersion string) *LocalClient {
	return &LocalClient{PolicyVersion: policyVersion}
}

func (c *LocalClient) Evaluate(_ context.Context, _ string, input map[string]any) (map[string]any, error) {
	blockedTags := []string{}
	if ctxRaw, ok := input["context"]; ok {
		if m, ok := ctxRaw.(map[string]string); ok {
			if deny, ok := m["deny_tag"]; ok && deny != "" {
				blockedTags = append(blockedTags, deny)
			}
		}
	}
	allowedMutations := []string{"set_name", "set_plan", "set_region", "set_feature_flag"}
	if classRaw, ok := input["mutation_class"]; ok {
		if class, ok := classRaw.(string); ok {
			allowed := false
			for _, x := range allowedMutations {
				if class == x {
					allowed = true
					break
				}
			}
			return map[string]any{
				"allowed":           allowed,
				"policy_version":    c.PolicyVersion,
				"blocked_tags":      blockedTags,
				"allowed_mutations": allowedMutations,
			}, nil
		}
	}
	return map[string]any{
		"allowed":        true,
		"policy_version": c.PolicyVersion,
		"blocked_tags":   blockedTags,
	}, nil
}
