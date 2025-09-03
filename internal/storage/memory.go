package storage

import (
	"errors"
	"sync"
)

// Storage interface for URL mappings
type Storage interface {
	Save(shortCode, longURL string) error
	Get(shortCode string) (string, bool)
	GetShortCode(longURL string) (string, bool)
}

// InMemoryStorage implements Storage using in-memory maps
type InMemoryStorage struct {
	mu          sync.RWMutex
	shortToLong map[string]string
	longToShort map[string]string
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		shortToLong: make(map[string]string),
		longToShort: make(map[string]string),
	}
}

func NewMemoryStore() *InMemoryStorage {
	return NewInMemoryStorage()
}

func (s *InMemoryStorage) Save(shortCode, longURL string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.shortToLong[shortCode]; exists {
		return errors.New("short code already exists")
	}
	if _, exists := s.longToShort[longURL]; exists {
		return errors.New("long URL already shortened")
	}

	s.shortToLong[shortCode] = longURL
	s.longToShort[longURL] = shortCode
	return nil
}

func (s *InMemoryStorage) Get(shortCode string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	longURL, ok := s.shortToLong[shortCode]
	return longURL, ok
}

func (s *InMemoryStorage) GetShortCode(longURL string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	shortCode, ok := s.longToShort[longURL]
	return shortCode, ok
}