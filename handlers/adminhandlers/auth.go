package adminhandlers

import (
	"encoding/json"
	"net/http"
	"siwaris-gelora-service/auth"
)

// LoginRequest defines credentials payload
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse returns token
type LoginResponse struct {
	Token string `json:"token"`
}

// AdminLoginHandler verifies admin credentials and returns a JWT token
func AdminLoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var creds LoginRequest
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, "Payload tidak valid", http.StatusBadRequest)
		return
	}

	if creds.Email != auth.AdminEmail || creds.Password != auth.AdminPassword {
		http.Error(w, "Email atau password salah", http.StatusUnauthorized)
		return
	}

	token, err := auth.GenerateJWT(creds.Email)
	if err != nil {
		http.Error(w, "Gagal membuat token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(LoginResponse{Token: token})
}
