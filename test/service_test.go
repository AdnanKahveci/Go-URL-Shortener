package test

import (
    "testing"

    "go-url-shortener/internal/service"
    "go-url-shortener/internal/storage"
)

func TestService_CreateAndResolve_Happy(t *testing.T) {
    store := storage.NewInMemoryStorage()
    s := service.NewService(store)

    shortCode, err := s.ShortenURL("https://example.com")
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if shortCode == "" {
        t.Fatalf("expected shortCode, got empty")
    }

    url, err2 := s.GetLongURL(shortCode)
    if err2 != nil || url != "https://example.com" {
        t.Fatalf("resolve failed: %v %s", err2, url)
    }
}


