package integrations

// PlatformAdapter is a generated integration seam for deterministic workflows.
type PlatformAdapter struct{}

func (a PlatformAdapter) Name() string {
	return "platform"
}

func (a PlatformAdapter) ValidateConfig(cfg map[string]string) bool {
	return cfg != nil
}
