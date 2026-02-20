package idempotency

import "sync"

type Store struct {
	mu   sync.RWMutex
	data map[string][]byte
}

func NewStore() *Store {
	return &Store{data: map[string][]byte{}}
}

func (s *Store) Read(key string) ([]byte, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.data[key]
	return v, ok
}

func (s *Store) Write(key string, payload []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = payload
}
