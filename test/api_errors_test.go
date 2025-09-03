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

func TestAPI_InvalidURL_And_NotFound(t *testing.T) {
    h := api.NewHandler(svc.New(storage.NewMemoryStore(), "http://localhost:8080"))
    mux := http.NewServeMux()
    h.Register(mux)

    // invalid URL
    body, _ := json.Marshal(map[string]any{"url": "ftp://bad"})
    req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(body))
    w := httptest.NewRecorder()
    mux.ServeHTTP(w, req)
    if w.Code != http.StatusBadRequest {
        t.Fatalf("expected 400, got %d", w.Code)
    }

    // unknown code -> 404
    req2 := httptest.NewRequest(http.MethodGet, "/does-not-exist", nil)
    w2 := httptest.NewRecorder()
    mux.ServeHTTP(w2, req2)
    if w2.Code != http.StatusNotFound {
        t.Fatalf("expected 404, got %d", w2.Code)
    }
}


