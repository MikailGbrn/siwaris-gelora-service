package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"siwaris-gelora-service/auth"
	"siwaris-gelora-service/config"
	"siwaris-gelora-service/db"
	"siwaris-gelora-service/handlers/adminhandlers"
	"siwaris-gelora-service/handlers/citizenhandlers"
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
	// Load environment variables from .env file
	config.LoadEnv(".env")

	// Initialize database
	dbPath := "./siwaris.db"
	database := db.InitDB(dbPath)
	defer database.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", HealthHandler)
	
	// Citizen Routes
	mux.HandleFunc("POST /api/apply", citizenhandlers.ApplyHandler)
	mux.HandleFunc("POST /api/apply/revision", citizenhandlers.RevisionHandler)
	mux.HandleFunc("GET /api/track", citizenhandlers.TrackHandler)
	
	// Admin Login Route
	mux.HandleFunc("POST /api/admin/login", adminhandlers.AdminLoginHandler)

	// Admin Protected Routes
	mux.HandleFunc("GET /api/admin/applications", auth.AuthMiddleware(adminhandlers.AdminListApplicationsHandler))
	mux.HandleFunc("GET /api/admin/applications/{id}", auth.AuthMiddleware(adminhandlers.AdminGetApplicationHandler))
	mux.HandleFunc("PUT /api/admin/applications/{id}", auth.AuthMiddleware(adminhandlers.AdminUpdateStatusHandler))
	mux.HandleFunc("GET /api/admin/applications/{id}/pdf", auth.AuthMiddleware(adminhandlers.AdminDownloadPDFHandler))

	// Serve Uploaded Files
	fileHandler := http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads")))
	mux.Handle("/uploads/", fileHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server SIWARIS GELORA starting on port %s...\n", port)
	if err := http.ListenAndServe(":"+port, CORSMiddleware(mux)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
