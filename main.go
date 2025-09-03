package main

import (
	"log"
	"net/http"
	"os"

	"go-url-shortener/internal/handler"
	"go-url-shortener/internal/service"
	"go-url-shortener/internal/storage"
)

func main() {
	// Get server address from environment or use default
	addr := getEnv("ADDR", ":8080")
	baseURL := getEnv("BASE_URL", "http://localhost:8080")

	// Initialize storage, service and handler
	store := storage.NewInMemoryStorage()
	svc := service.NewService(store)
	h := handler.NewHandler(svc, baseURL)

	// Register routes according to acceptance criteria
	http.HandleFunc("/api/shorten", h.Shorten) // POST /api/shorten for creating short URLs
	http.HandleFunc("/", h.Redirect)           // GET /{code} for redirecting

	log.Printf("Starting URL Shortener server on %s", addr)
	log.Printf("Base URL: %s", baseURL)
	log.Printf("API Endpoints:")
	log.Printf("  POST %s/api/shorten - Create short URL", baseURL)
	log.Printf("  GET  %s/{code} - Redirect to original URL", baseURL)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}