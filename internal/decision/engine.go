package decision

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
	"strings"
	"time"

	"github.com/restarone/violet-deterministic-api/internal/adapters/gorse"
	"github.com/restarone/violet-deterministic-api/internal/adapters/gorules"
)

type Engine struct {
	PolicyVersion string
	DataVersion   string

	gorse  gorse.Client
	policy gorules.Client
}

func NewEngine(policyVersion, dataVersion string, gorseClient gorse.Client, policyClient gorules.Client) *Engine {
	return &Engine{
		PolicyVersion: policyVersion,
		DataVersion:   dataVersion,
		gorse:         gorseClient,
		policy:        policyClient,
	}
}

func (e *Engine) Decide(ctx context.Context, req DecisionRequest) DecisionResponse {
	stages := make([]StageTrace, 0, 3)
	dependencyStatus := "ok"

	gorseIDs := make([]string, 0)
	if e.gorse != nil {
		ids, err := e.gorse.Recommend(ctx, req.UserID, len(req.CandidateItems))
		if err != nil {
			dependencyStatus = "degraded"
			stages = append(stages, StageTrace{Stage: "gorse_recommend", Outcome: "degraded", ErrMessage: err.Error()})
		} else {
			gorseIDs = ids
			stages = append(stages, StageTrace{Stage: "gorse_recommend", Outcome: "ok"})
		}
	} else {
		stages = append(stages, StageTrace{Stage: "gorse_recommend", Outcome: "skipped"})
	}

	blockedTags := map[string]struct{}{}
	if e.policy != nil {
		policyIn := map[string]any{
			"surface":       req.Surface,
			"context":       req.Context,
			"candidate_len": len(req.CandidateItems),
		}
		out, err := e.policy.Evaluate(ctx, req.TenantID, policyIn)
		if err != nil {
			dependencyStatus = "degraded"
			stages = append(stages, StageTrace{Stage: "policy_eval", Outcome: "degraded", ErrMessage: err.Error()})
		} else {
			if denyRaw, ok := out["blocked_tags"]; ok {
				if deny, ok := denyRaw.([]string); ok {
					for _, t := range deny {
						blockedTags[t] = struct{}{}
					}
				}
			}
			stages = append(stages, StageTrace{Stage: "policy_eval", Outcome: "ok"})
		}
	} else {
		stages = append(stages, StageTrace{Stage: "policy_eval", Outcome: "skipped"})
	}

	gorseRank := map[string]int{}
	for i, itemID := range gorseIDs {
		gorseRank[itemID] = i
	}

	scored := make([]RankedItem, 0, len(req.CandidateItems))
	plan := req.Context["plan"]
	for _, c := range req.CandidateItems {
		if c.Blocked || hasBlockedTag(c.Tags, blockedTags) {
			continue
		}
		score := c.BaseScore
		if plan == "enterprise" && hasTag(c.Tags, "enterprise") {
			score += 10
		}
		if rank, ok := gorseRank[c.ItemID]; ok {
			score += float64(len(gorseIDs)-rank) * 0.01
		}
		scored = append(scored, RankedItem{ItemID: c.ItemID, Score: score})
	}

	sort.Slice(scored, func(i, j int) bool {
		if scored[i].Score == scored[j].Score {
			return scored[i].ItemID < scored[j].ItemID
		}
		return scored[i].Score > scored[j].Score
	})
	stages = append(stages, StageTrace{Stage: "rank", Outcome: "ok"})

	h := hashDecision(req, e.PolicyVersion, e.DataVersion, gorseIDs, stages)
	decisionID := "dec_" + h[:16]
	traceID := "trc_" + h[16:28]

	return DecisionResponse{
		DecisionID:       decisionID,
		DecisionHash:     h,
		PolicyVersion:    e.PolicyVersion,
		DataVersion:      e.DataVersion,
		GeneratedAt:      time.Now().UTC(),
		TraceID:          traceID,
		DependencyStatus: dependencyStatus,
		Items:            scored,
		Stages:           stages,
	}
}

type canonicalPair struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type canonicalDecision struct {
	Request        DecisionRequest `json:"request"`
	PolicyVersion  string          `json:"policy_version"`
	DataVersion    string          `json:"data_version"`
	GorseCandidate []string        `json:"gorse_candidate_ids"`
	Stages         []StageTrace    `json:"stages"`
}

func hashDecision(req DecisionRequest, policyVersion, dataVersion string, gorseIDs []string, stages []StageTrace) string {
	normalized := DecisionRequest{
		TenantID:       req.TenantID,
		UserID:         req.UserID,
		Surface:        req.Surface,
		Context:        normalizeContext(req.Context),
		CandidateItems: normalizeCandidates(req.CandidateItems),
	}
	payload := canonicalDecision{
		Request:        normalized,
		PolicyVersion:  policyVersion,
		DataVersion:    dataVersion,
		GorseCandidate: append([]string(nil), gorseIDs...),
		Stages:         append([]StageTrace(nil), stages...),
	}
	b, _ := json.Marshal(payload)
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

func normalizeContext(ctx map[string]string) map[string]string {
	if len(ctx) == 0 {
		return nil
	}
	keys := make([]string, 0, len(ctx))
	for k := range ctx {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make(map[string]string, len(keys))
	for _, k := range keys {
		out[k] = ctx[k]
	}
	return out
}

func normalizeCandidates(in []CandidateItem) []CandidateItem {
	out := append([]CandidateItem(nil), in...)
	for i := range out {
		out[i].Tags = normalizeTags(out[i].Tags)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].ItemID == out[j].ItemID {
			if out[i].BaseScore == out[j].BaseScore {
				return strings.Join(out[i].Tags, ",") < strings.Join(out[j].Tags, ",")
			}
			return out[i].BaseScore < out[j].BaseScore
		}
		return out[i].ItemID < out[j].ItemID
	})
	return out
}

func normalizeTags(tags []string) []string {
	if len(tags) == 0 {
		return nil
	}
	out := append([]string(nil), tags...)
	sort.Strings(out)
	return out
}

func hasTag(tags []string, value string) bool {
	for _, t := range tags {
		if t == value {
			return true
		}
	}
	return false
}

func hasBlockedTag(tags []string, blocked map[string]struct{}) bool {
	for _, t := range tags {
		if _, ok := blocked[t]; ok {
			return true
		}
	}
	return false
}
