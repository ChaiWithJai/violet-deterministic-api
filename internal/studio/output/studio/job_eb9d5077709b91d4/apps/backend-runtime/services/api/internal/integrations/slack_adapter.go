package integrations

// SlackAdapter is a generated integration seam for deterministic workflows.
type SlackAdapter struct{}

func (a SlackAdapter) Name() string {
	return "slack"
}

func (a SlackAdapter) ValidateConfig(cfg map[string]string) bool {
	return cfg != nil
}
