package integrations

// StripeAdapter is a generated integration seam for deterministic workflows.
type StripeAdapter struct{}

func (a StripeAdapter) Name() string {
	return "stripe"
}

func (a StripeAdapter) ValidateConfig(cfg map[string]string) bool {
	return cfg != nil
}
