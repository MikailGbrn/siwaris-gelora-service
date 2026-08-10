package adminhandlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"siwaris-gelora-service/auth"
	"siwaris-gelora-service/db"

	"golang.org/x/crypto/bcrypt"
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

// AdminLoginHandler verifies admin credentials against database and returns a JWT token
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

	// Query hashed password from database
	var passwordHash string
	err = db.DB.QueryRow(db.Rebind("SELECT password_hash FROM admins WHERE email = ?"), creds.Email).Scan(&passwordHash)
	if err == sql.ErrNoRows {
		http.Error(w, "Email atau password salah", http.StatusUnauthorized)
		return
	} else if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Compare password with hashed value using bcrypt
	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(creds.Password))
	if err != nil {
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
