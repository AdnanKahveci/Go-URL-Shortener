package api

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"go-url-shortener/internal/service"
)

// Handler wires HTTP endpoints to the ShortenerService.
type Handler struct {
	service *service.ShortenerService
	limiter *rateLimiter
}

func NewHandler(svc *service.ShortenerService) *Handler {
	return &Handler{service: svc, limiter: newRateLimiter(60, time.Minute)} // 60 req/min/IP
}

// Register attaches the HTTP routes to the given mux.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/shorten", h.withRateLimit(h.handleShorten)) // POST create short link
	mux.HandleFunc("/api/stats/", h.withRateLimit(h.handleStats))    // GET stats by code
	mux.HandleFunc("/", h.withRateLimit(h.handleRedirect))           // GET redirect by code
}

type shortenRequest struct {
	URL              string `json:"url"`
	CustomAlias      string `json:"custom_alias"`
	ExpireInSeconds  int    `json:"expire_in_seconds"`
}

type shortenResponse struct {
	Code  string `json:"code"`
	Short string `json:"short"`
}

// handleShorten accepts JSON and creates a short URL
func (h *Handler) handleShorten(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req shortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	resp, err := h.service.CreateShort(service.CreateShortRequest{URL: req.URL, CustomAlias: req.CustomAlias, ExpireInSeconds: req.ExpireInSeconds})
	if err != nil {
		switch err {
		case service.ErrInvalidURL:
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		case service.ErrAliasTaken:
			writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
			return
		default:
			log.Printf("create short error: %v", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
			return
		}
	}
	writeJSON(w, http.StatusCreated, shortenResponse{Code: resp.Code, Short: resp.Short})
}

// handleRedirect looks up the code and responds with a 302 redirect
func (h *Handler) handleRedirect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/")
	if path == "" || path == "api" || strings.HasPrefix(path, "api/") {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	original, ok := h.service.ResolveAndTrack(path)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	http.Redirect(w, r, original, http.StatusFound)
}

// handleStats returns analytics and metadata for a given code
func (h *Handler) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	code := strings.TrimPrefix(r.URL.Path, "/api/stats/")
	if code == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing code"})
		return
	}
	st, ok := h.service.GetStats(code)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, st)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// --- Simple IP rate limiter ---

type rateLimiter struct {
	mu     sync.Mutex
	hits   map[string]clientHits
	limit  int
	window time.Duration
}

type clientHits struct {
	count int
	reset time.Time
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		hits:   make(map[string]clientHits),
		limit:  limit,
		window: window,
	}
}

func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	h, ok := rl.hits[ip]
	if !ok || now.After(h.reset) {
		rl.hits[ip] = clientHits{count: 1, reset: now.Add(rl.window)}
		return true
	}
	if h.count >= rl.limit {
		return false
	}
	h.count++
	rl.hits[ip] = h
	return true
}

func (h *Handler) withRateLimit(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)
		if !h.limiter.allow(ip) {
			writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "rate limit exceeded"})
			return
		}
		next.ServeHTTP(w, r)
	}
}

func clientIP(r *http.Request) string {
	// prefer X-Forwarded-For if present, otherwise RemoteAddr
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}


