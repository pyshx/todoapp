package idempotency

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"
)

// Response represents a cached idempotent response
type Response struct {
	StatusCode int
	Body       []byte
	CreatedAt  time.Time
}

// Store provides idempotency key storage
type Store interface {
	Get(ctx context.Context, key string) (*Response, bool)
	Set(ctx context.Context, key string, resp *Response) error
	Delete(ctx context.Context, key string) error
}

// InMemoryStore is a simple in-memory implementation for development
type InMemoryStore struct {
	mu      sync.RWMutex
	entries map[string]*Response
	ttl     time.Duration
}

// NewInMemoryStore creates a new in-memory idempotency store
func NewInMemoryStore(ttl time.Duration) *InMemoryStore {
	store := &InMemoryStore{
		entries: make(map[string]*Response),
		ttl:     ttl,
	}
	go store.cleanup()
	return store
}

func (s *InMemoryStore) Get(ctx context.Context, key string) (*Response, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	resp, ok := s.entries[key]
	if !ok {
		return nil, false
	}

	if time.Since(resp.CreatedAt) > s.ttl {
		return nil, false
	}

	return resp, true
}

func (s *InMemoryStore) Set(ctx context.Context, key string, resp *Response) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	resp.CreatedAt = time.Now()
	s.entries[key] = resp
	return nil
}

func (s *InMemoryStore) Delete(ctx context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.entries, key)
	return nil
}

func (s *InMemoryStore) cleanup() {
	ticker := time.NewTicker(s.ttl / 2)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for key, resp := range s.entries {
			if now.Sub(resp.CreatedAt) > s.ttl {
				delete(s.entries, key)
			}
		}
		s.mu.Unlock()
	}
}

// GenerateKey creates a deterministic key from request data
func GenerateKey(userID, method string, body []byte) string {
	h := sha256.New()
	h.Write([]byte(userID))
	h.Write([]byte(method))
	h.Write(body)
	return hex.EncodeToString(h.Sum(nil))
}
