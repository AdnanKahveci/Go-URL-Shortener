package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"go-url-shortener/internal/service"
)

// Handler provides HTTP handlers for URL shortener API
type Handler struct {
	service service.ServiceInterface
	baseURL string
}

func NewHandler(s service.ServiceInterface, baseURL string) *Handler {
	return &Handler{
		service: s,
		baseURL: baseURL,
	}
}

// API request/response types
type ShortenRequest struct {
	URL         string `json:"url"`
	CustomAlias string `json:"custom_alias,omitempty"`
}

type ShortenResponse struct {
	Short string `json:"short"`
	Code  string `json:"code"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

// Shorten handles POST /api/shorten requests
func (h *Handler) Shorten(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validate URL format
	if req.URL == "" || !isValidURL(req.URL) {
		http.Error(w, "Invalid or empty 'url' provided. Must be a valid http(s) URL.", http.StatusBadRequest)
		return
	}

	// Shorten URL with or without custom alias
	var shortCode string
	var err error
	if req.CustomAlias != "" {
		shortCode, err = h.service.ShortenURLWithAlias(req.URL, req.CustomAlias)
	} else {
		shortCode, err = h.service.ShortenURL(req.URL)
	}
	
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			http.Error(w, "Custom alias already taken", http.StatusConflict)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to shorten URL: %v", err), http.StatusInternalServerError)
		return
	}

	shortURL := strings.TrimSuffix(h.baseURL, "/") + "/" + shortCode
	resp := ShortenResponse{
		Short: shortURL,
		Code:  shortCode,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// Redirect handles GET /{code} requests  
func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Skip API endpoints
	if strings.HasPrefix(r.URL.Path, "/shorten") {
		http.NotFound(w, r)
		return
	}

	shortCode := r.URL.Path[1:]
	if shortCode == "" {
		http.Error(w, "Short code not provided in URL path", http.StatusBadRequest)
		return
	}

	longURL, err := h.service.GetLongURL(shortCode)
	if err != nil {
		if err.Error() == "short code not found" {
			http.Error(w, "Short URL not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to retrieve long URL: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("Redirecting %s -> %s", shortCode, longURL)
	http.Redirect(w, r, longURL, http.StatusMovedPermanently)
}

// isValidURL checks if URL uses http/https scheme
func isValidURL(str string) bool {
	u, err := url.Parse(str)
	if err != nil {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}
