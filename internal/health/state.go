package health

import (
	"sync"
	"time"
)

type State struct {
	mu          sync.RWMutex
	ready       bool
	lastError   string
	lastSuccess time.Time
}

func New() *State { return &State{} }

func (s *State) MarkSuccess() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ready = true
	s.lastError = ""
	s.lastSuccess = time.Now()
}

func (s *State) MarkFailure(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err != nil {
		s.lastError = err.Error()
	}
}

func (s *State) Snapshot() (ready bool, lastError string, lastSuccess time.Time) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.ready, s.lastError, s.lastSuccess
}
