package test

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "go-url-shortener/internal/api"
    svc "go-url-shortener/internal/service"
    "go-url-shortener/internal/storage"
)

func TestAPI_Shorten_And_Redirect(t *testing.T) {
    s := svc.New(storage.NewMemoryStore(), "http://localhost:8080")
    h := api.NewHandler(s)
    mux := http.NewServeMux()
    h.Register(mux)

    // create short
    body, _ := json.Marshal(map[string]any{"url": "https://example.com"})
    req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(body))
    w := httptest.NewRecorder()
    mux.ServeHTTP(w, req)
    if w.Code != http.StatusCreated {
        t.Fatalf("expected 201, got %d", w.Code)
    }
    var resp struct{ Code string }
    _ = json.Unmarshal(w.Body.Bytes(), &resp)
    if resp.Code == "" {
        t.Fatalf("missing code")
    }

    // redirect
    req2 := httptest.NewRequest(http.MethodGet, "/"+resp.Code, nil)
    w2 := httptest.NewRecorder()
    mux.ServeHTTP(w2, req2)
    if w2.Code != http.StatusFound {
        t.Fatalf("expected 302, got %d", w2.Code)
    }
    if loc := w2.Header().Get("Location"); loc != "https://example.com" {
        t.Fatalf("unexpected location: %s", loc)
    }
}


