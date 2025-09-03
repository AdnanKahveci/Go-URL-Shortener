package test

import (
    "testing"

    svc "go-url-shortener/internal/service"
    "go-url-shortener/internal/storage"
)

func TestService_InvalidURL(t *testing.T) {
    store := storage.NewMemoryStore()
    s := svc.New(store, "http://localhost:8080")
    _, err := s.CreateShort(svc.CreateShortRequest{URL: "ftp://example.com"})
    if err == nil || err != svc.ErrInvalidURL {
        t.Fatalf("expected ErrInvalidURL, got %v", err)
    }
}

func TestService_AliasTaken(t *testing.T) {
    store := storage.NewMemoryStore()
    s := svc.New(store, "http://localhost:8080")
    _, err := s.CreateShort(svc.CreateShortRequest{URL: "https://a.com", CustomAlias: "go-docs"})
    if err != nil {
        t.Fatalf("unexpected: %v", err)
    }
    _, err = s.CreateShort(svc.CreateShortRequest{URL: "https://b.com", CustomAlias: "go-docs"})
    if err == nil || err != svc.ErrAliasTaken {
        t.Fatalf("expected ErrAliasTaken, got %v", err)
    }
}


