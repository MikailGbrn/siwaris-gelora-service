package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"siwaris-gelora-service/db"
)

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

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server SIWARIS GELORA starting on port %s...\n", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
