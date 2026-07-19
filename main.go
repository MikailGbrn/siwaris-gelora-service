package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"siwaris-gelora-service/db"
	"siwaris-gelora-service/handlers/adminhandlers"
)

// CORSMiddleware handles Cross-Origin Resource Sharing (CORS) header injection
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// HealthHandler is a simple status check endpoint
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"message": "SIWARIS GELORA API is running",
	})
}

func main() {
	// Initialize database
	dbPath := "./siwaris.db"
	database := db.InitDB(dbPath)
	defer database.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", HealthHandler)
	
	// Admin Login Route
	mux.HandleFunc("POST /api/admin/login", adminhandlers.AdminLoginHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server SIWARIS GELORA starting on port %s...\n", port)
	if err := http.ListenAndServe(":"+port, CORSMiddleware(mux)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
