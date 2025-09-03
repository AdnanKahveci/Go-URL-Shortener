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

// Handler provides HTTP handlers for the shortener service's API endpoints.
type Handler struct {
	service service.ServiceInterface // The service containing the business logic
	baseURL string                   // The base URL of the shortener service (e.g., "http://localhost:8080/")
}

// NewHandler creates and returns a new Handler instance.
func NewHandler(s service.ServiceInterface, baseURL string) *Handler {
	return &Handler{
		service: s,
		baseURL: baseURL,
	}
}

// ShortenRequest defines the structure for the request body when shortening a URL.
type ShortenRequest struct {
	URL         string `json:"url"`                    // The original long URL to be shortened
	CustomAlias string `json:"custom_alias,omitempty"` // Optional custom alias
}

// ShortenResponse defines the structure for the response body after shortening a URL.
type ShortenResponse struct {
	Short string `json:"short"` // The newly generated short URL
	Code  string `json:"code"`  // Just the short code part
}

// ErrorResponse defines a generic error response structure for API errors.
type ErrorResponse struct {
	Message string `json:"message"` // A descriptive error message
}

// Shorten handles POST requests to create a short URL.
// It expects a JSON request body like: `{"url": "https://example.com/very/long/url"}`
// It responds with JSON like: `{"short": "http://localhost:8080/abc123", "code": "abc123"}`
func (h *Handler) Shorten(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests for this endpoint.
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ShortenRequest
	// Decode the JSON request body into the ShortenRequest struct.
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validate the provided long URL.
	if req.URL == "" || !isValidURL(req.URL) {
		http.Error(w, "Invalid or empty 'url' provided. Must be a valid http(s) URL.", http.StatusBadRequest)
		return
	}

	// Call the core service to shorten the URL.
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

	// Construct the full short URL using the base URL and the generated short code.
	shortURL := strings.TrimSuffix(h.baseURL, "/") + "/" + shortCode
	resp := ShortenResponse{
		Short: shortURL,
		Code:  shortCode,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// Redirect handles GET requests to redirect to the original long URL.
// It expects the short code in the URL path, e.g., `/abc123`.
func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	// Only allow GET requests for this endpoint.
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Skip API endpoints
	if strings.HasPrefix(r.URL.Path, "/shorten") {
		http.NotFound(w, r)
		return
	}

	// Extract the short code from the URL path.
	// r.URL.Path will be like "/abc123", so we slice it to remove the leading "/".
	shortCode := r.URL.Path[1:]

	// If no short code is provided (e.g., request to the root "/"), return a bad request error.
	if shortCode == "" {
		http.Error(w, "Short code not provided in URL path", http.StatusBadRequest)
		return
	}

	// Call the core service to retrieve the original long URL.
	longURL, err := h.service.GetLongURL(shortCode)
	if err != nil {
		// If the short code is not found, return a 404 Not Found error.
		if err.Error() == "short code not found" {
			http.Error(w, "Short URL not found", http.StatusNotFound)
			return
		}
		// For other service errors, return a 500 Internal Server Error.
		http.Error(w, fmt.Sprintf("Failed to retrieve long URL: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("Redirecting %s -> %s", shortCode, longURL)

	// Perform a permanent redirect (HTTP 301 Moved Permanently) to the original long URL.
	http.Redirect(w, r, longURL, http.StatusMovedPermanently)
}

// isValidURL validates that the URL is a valid HTTP or HTTPS URL
func isValidURL(str string) bool {
	u, err := url.Parse(str)
	if err != nil {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}
