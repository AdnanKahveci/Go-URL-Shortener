package main

import (
	"log"
	"net/http"
	"os"

	"go-url-shortener/internal/api"
	"go-url-shortener/internal/service"
	"go-url-shortener/internal/storage"
)

func main() {
	addr := getEnv("ADDR", ":8080")
	domain := getEnv("DOMAIN", "http://localhost:8080")

	store := storage.NewMemoryStore()
	svc := service.New(store, domain)
	h := api.NewHandler(svc)

	mux := http.NewServeMux()
	h.Register(mux)

	log.Printf("starting server on %s with domain %s", addr, domain)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}


