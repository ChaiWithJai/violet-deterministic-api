package decision

import "sync"

type ReplayStore struct {
	mu   sync.RWMutex
	data map[string]DecisionResponse
}

func NewReplayStore() *ReplayStore {
	return &ReplayStore{data: map[string]DecisionResponse{}}
}

func (s *ReplayStore) Put(resp DecisionResponse) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[resp.DecisionID] = resp
}

func (s *ReplayStore) Get(decisionID string) (DecisionResponse, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	resp, ok := s.data[decisionID]
	return resp, ok
}
