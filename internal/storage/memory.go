package storage

import (
	"errors"
	"sync"
	"time"
)

// Entry represents a stored short-code mapping to an original URL.
type Entry struct {
	OriginalURL string
	CreatedAt   time.Time
	ExpireAt    *time.Time
	Clicks      int64
	LastAccess  *time.Time
}

// Store defines the behaviors required by the URL shortener service.
type Store interface {
	// Save stores the mapping if the code does not already exist.
	Save(code string, url string) error
	// Get retrieves the original URL associated with a code.
	Get(code string) (string, bool)
	// GetEntry returns the full entry including analytics/expiry.
	GetEntry(code string) (Entry, bool)
	// SetExpire sets the expiration time for a code.
	SetExpire(code string, expireAt time.Time) bool
	// TrackHit increments click count and updates last access time.
	TrackHit(code string) bool
	// Exists returns true if the code is already stored.
	Exists(code string) bool
}

var (
	// ErrCodeExists is returned when attempting to save using a code that is already taken.
	ErrCodeExists = errors.New("code already exists")
)

// MemoryStore is a simple in-memory Store implementation using a map with a RWMutex.
type MemoryStore struct {
	mu          sync.RWMutex
	codeToEntry map[string]Entry
}

// NewMemoryStore creates a new MemoryStore.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{codeToEntry: make(map[string]Entry)}
}

// Save stores a new code->URL mapping, returning ErrCodeExists if present.
func (s *MemoryStore) Save(code string, url string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.codeToEntry[code]; exists {
		return ErrCodeExists
	}
	s.codeToEntry[code] = Entry{OriginalURL: url, CreatedAt: time.Now()}
	return nil
}

// Get returns the original URL for a given code.
func (s *MemoryStore) Get(code string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.codeToEntry[code]
	if !ok {
		return "", false
	}
	return e.OriginalURL, true
}

// GetEntry returns the stored entry for a code.
func (s *MemoryStore) GetEntry(code string) (Entry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.codeToEntry[code]
	return e, ok
}

// SetExpire sets the expiration time for an existing code.
func (s *MemoryStore) SetExpire(code string, expireAt time.Time) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.codeToEntry[code]
	if !ok {
		return false
	}
	e.ExpireAt = &expireAt
	s.codeToEntry[code] = e
	return true
}

// TrackHit updates analytics for a code.
func (s *MemoryStore) TrackHit(code string) bool {
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.codeToEntry[code]
	if !ok {
		return false
	}
	e.Clicks++
	e.LastAccess = &now
	s.codeToEntry[code] = e
	return true
}

// Exists indicates whether the code is already stored.
func (s *MemoryStore) Exists(code string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.codeToEntry[code]
	return ok
}


