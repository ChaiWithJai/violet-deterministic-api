package decision

import (
	"context"
	"testing"
)

type stubGorse struct {
	ids []string
}

func (s stubGorse) Recommend(context.Context, string, int) ([]string, error) {
	return s.ids, nil
}

type stubPolicy struct {
	blocked []string
}

func (s stubPolicy) Evaluate(context.Context, string, map[string]any) (map[string]any, error) {
	return map[string]any{"allowed": true, "blocked_tags": s.blocked}, nil
}

func TestDecideDeterministicOrdering(t *testing.T) {
	engine := NewEngine("policy-v1", "data-v1", stubGorse{}, stubPolicy{})
	resp := engine.Decide(context.Background(), DecisionRequest{
		TenantID: "t",
		UserID:   "u",
		Surface:  "test",
		Context:  map[string]string{"plan": "enterprise"},
		CandidateItems: []CandidateItem{
			{ItemID: "b", BaseScore: 100},
			{ItemID: "a", BaseScore: 100},
		},
	})

	if len(resp.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(resp.Items))
	}
	if resp.Items[0].ItemID != "a" {
		t.Fatalf("expected a first for tie-breaker, got %s", resp.Items[0].ItemID)
	}
}

func TestHashStableAcrossInputOrder(t *testing.T) {
	engine := NewEngine("policy-v1", "data-v1", stubGorse{}, stubPolicy{})
	base := DecisionRequest{
		TenantID: "t",
		UserID:   "u",
		Surface:  "s",
		Context:  map[string]string{"b": "2", "a": "1"},
		CandidateItems: []CandidateItem{
			{ItemID: "x", BaseScore: 1, Tags: []string{"b", "a"}},
			{ItemID: "y", BaseScore: 2},
		},
	}
	alt := DecisionRequest{
		TenantID: "t",
		UserID:   "u",
		Surface:  "s",
		Context:  map[string]string{"a": "1", "b": "2"},
		CandidateItems: []CandidateItem{
			{ItemID: "y", BaseScore: 2},
			{ItemID: "x", BaseScore: 1, Tags: []string{"a", "b"}},
		},
	}

	r1 := engine.Decide(context.Background(), base)
	r2 := engine.Decide(context.Background(), alt)
	if r1.DecisionHash != r2.DecisionHash {
		t.Fatalf("expected stable hash, got %s vs %s", r1.DecisionHash, r2.DecisionHash)
	}
}

func TestPolicyBlockedTag(t *testing.T) {
	engine := NewEngine("policy-v1", "data-v1", stubGorse{}, stubPolicy{blocked: []string{"blocked"}})
	resp := engine.Decide(context.Background(), DecisionRequest{
		TenantID: "t",
		UserID:   "u",
		Surface:  "s",
		CandidateItems: []CandidateItem{
			{ItemID: "good", BaseScore: 2},
			{ItemID: "bad", BaseScore: 99, Tags: []string{"blocked"}},
		},
	})
	if len(resp.Items) != 1 || resp.Items[0].ItemID != "good" {
		t.Fatalf("expected blocked item filtered, got %#v", resp.Items)
	}
}
