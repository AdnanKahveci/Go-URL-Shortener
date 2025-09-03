package test

import (
    "testing"

    svc "go-url-shortener/internal/service"
    "go-url-shortener/internal/storage"
)

func TestService_CreateAndResolve_Happy(t *testing.T) {
    store := storage.NewMemoryStore()
    s := svc.New(store, "http://localhost:8080")

    resp, err := s.CreateShort(svc.CreateShortRequest{URL: "https://example.com"})
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if resp.Code == "" || resp.Short == "" {
        t.Fatalf("expected code and short, got %+v", resp)
    }

    url, ok := s.Resolve(resp.Code)
    if !ok || url != "https://example.com" {
        t.Fatalf("resolve failed: %v %s", ok, url)
    }
}


