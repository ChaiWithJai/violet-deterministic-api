package gorse

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"time"
)

type HTTPClient struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

func NewHTTPClient(baseURL, apiKey string) *HTTPClient {
	return &HTTPClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		http: &http.Client{
			Timeout: 2 * time.Second,
		},
	}
}

func (c *HTTPClient) Recommend(ctx context.Context, userID string, n int) ([]string, error) {
	if c.baseURL == "" || userID == "" || n <= 0 {
		return nil, nil
	}
	u := fmt.Sprintf("%s/api/recommend/%s?n=%d", c.baseURL, url.PathEscape(userID), n)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	if c.apiKey != "" {
		req.Header.Set("X-API-Key", c.apiKey)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("gorse_status_%d", resp.StatusCode)
	}

	var raw any
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	ids := make([]string, 0)
	switch v := raw.(type) {
	case []any:
		for _, x := range v {
			switch t := x.(type) {
			case string:
				ids = append(ids, t)
			case map[string]any:
				if itemID, ok := t["Id"]; ok {
					if s, ok := itemID.(string); ok {
						ids = append(ids, s)
					}
				} else if itemID, ok := t["item_id"]; ok {
					if s, ok := itemID.(string); ok {
						ids = append(ids, s)
					}
				}
			}
		}
	case map[string]any:
		if items, ok := v["items"].([]any); ok {
			for _, x := range items {
				if m, ok := x.(map[string]any); ok {
					if s, ok := m["item_id"].(string); ok {
						ids = append(ids, s)
					}
				}
			}
		}
	}

	// Ensure deterministic tie behavior even if upstream emits duplicates.
	dedup := map[string]struct{}{}
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if _, ok := dedup[id]; ok {
			continue
		}
		dedup[id] = struct{}{}
		out = append(out, id)
	}
	sort.Strings(out)
	return out, nil
}

type NoopClient struct{}

func (NoopClient) Recommend(context.Context, string, int) ([]string, error) {
	return nil, nil
}
