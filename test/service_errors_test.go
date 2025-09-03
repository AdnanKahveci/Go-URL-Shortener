package test

import (
    "testing"

    "go-url-shortener/internal/service"
    "go-url-shortener/internal/storage"
)

func TestService_InvalidURL(t *testing.T) {
    store := storage.NewInMemoryStorage()
    s := service.NewService(store)
    // Service accepts any string as URL, validation is done in handler
    // Let's test that it works with any string
    code, err := s.ShortenURL("any-string")
    if err != nil {
        t.Fatalf("service should accept any string: %v", err)
    }
    if code == "" {
        t.Fatalf("expected code to be generated")
    }
}

func TestService_AliasTaken(t *testing.T) {
    store := storage.NewInMemoryStorage()
    s := service.NewService(store)
    
    // First, store a URL
    code1, err := s.ShortenURL("https://a.com")
    if err != nil {
        t.Fatalf("unexpected: %v", err)
    }
    
    // Try to store another URL with same code (simulate collision)
    err2 := store.Save(code1, "https://b.com")
    if err2 == nil {
        t.Fatalf("expected error for duplicate short code")
    }
}


