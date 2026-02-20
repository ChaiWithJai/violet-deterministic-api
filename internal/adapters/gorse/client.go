package gorse

import "context"

// Client defines the minimal integration seam for Gorse retrieval/ranking.
// Implementations should include deterministic tie-breaker handling before returning.
type Client interface {
	Recommend(ctx context.Context, userID string, n int) ([]string, error)
}
