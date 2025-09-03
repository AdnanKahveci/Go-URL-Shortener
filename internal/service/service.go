package service

import (
	"errors"
	"fmt"
	"sync"

	"go-url-shortener/internal/storage"
	"go-url-shortener/pkg/base62"
)

// ServiceInterface defines URL shortener business logic
type ServiceInterface interface {
	ShortenURL(longURL string) (string, error)
	ShortenURLWithAlias(longURL, customAlias string) (string, error)
	GetLongURL(shortCode string) (string, error)
}

// Service implements ServiceInterface
type Service struct {
	storage   storage.Storage
	counterMu sync.Mutex
	counter   int64
}

func NewService(s storage.Storage) *Service {
	return &Service{
		storage: s,
	}
}

func (s *Service) ShortenURL(longURL string) (string, error) {
	return s.ShortenURLWithAlias(longURL, "")
}

func (s *Service) ShortenURLWithAlias(longURL, customAlias string) (string, error) {
	// Check if URL already shortened
	if existingShortCode, ok := s.storage.GetShortCode(longURL); ok {
		return existingShortCode, nil
	}

	var shortCode string
	if customAlias != "" {
		// Check if custom alias is available
		if _, exists := s.storage.Get(customAlias); exists {
			return "", errors.New("custom alias already exists")
		}
		shortCode = customAlias
	} else {
		shortCode = s.GenerateShortCode()
	}

	err := s.storage.Save(shortCode, longURL)
	if err != nil {
		return "", fmt.Errorf("failed to save URL mapping: %w", err)
	}
	return shortCode, nil
}

func (s *Service) GetLongURL(shortCode string) (string, error) {
	longURL, ok := s.storage.Get(shortCode)
	if !ok {
		return "", errors.New("short code not found")
	}
	return longURL, nil
}

// GenerateShortCode creates unique base62 code
func (s *Service) GenerateShortCode() string {
	s.counterMu.Lock()
	defer s.counterMu.Unlock()
	s.counter++
	return base62.ToBase62(s.counter)
}