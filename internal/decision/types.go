package decision

import "time"

type CandidateItem struct {
	ItemID    string   `json:"item_id"`
	BaseScore float64  `json:"base_score"`
	Tags      []string `json:"tags,omitempty"`
	Blocked   bool     `json:"blocked,omitempty"`
}

type DecisionRequest struct {
	TenantID       string            `json:"tenant_id"`
	UserID         string            `json:"user_id"`
	Surface        string            `json:"surface"`
	Context        map[string]string `json:"context,omitempty"`
	CandidateItems []CandidateItem   `json:"candidate_items"`
}

type RankedItem struct {
	ItemID string  `json:"item_id"`
	Score  float64 `json:"score"`
}

type DecisionResponse struct {
	DecisionID       string       `json:"decision_id"`
	DecisionHash     string       `json:"decision_hash"`
	PolicyVersion    string       `json:"policy_version"`
	DataVersion      string       `json:"data_version"`
	GeneratedAt      time.Time    `json:"generated_at"`
	TraceID          string       `json:"trace_id"`
	DependencyStatus string       `json:"dependency_status"`
	Items            []RankedItem `json:"items"`
	Stages           []StageTrace `json:"stages"`
}

type StageTrace struct {
	Stage      string `json:"stage"`
	Outcome    string `json:"outcome"`
	ErrMessage string `json:"err_message,omitempty"`
}

// FeedbackEvent is intentionally minimal for scaffold stage.
type FeedbackEvent struct {
	DecisionID string `json:"decision_id"`
	ItemID     string `json:"item_id"`
	EventType  string `json:"event_type"`
	ActorID    string `json:"actor_id"`
}
