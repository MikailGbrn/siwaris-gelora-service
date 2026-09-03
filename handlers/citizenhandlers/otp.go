package citizenhandlers

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"siwaris-gelora-service/email"
)

type OTPItem struct {
	Code      string
	ExpiresAt time.Time
}

var (
	otpStore   = make(map[string]OTPItem)
	otpMutex   sync.RWMutex
	verifStore = make(map[string]bool)
)

func generateOTPCode() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(1000000))
	return fmt.Sprintf("%06d", n.Int64())
}

// SendOTPHandler sends an OTP code to the requested email
func SendOTPHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Email) == "" {
		http.Error(w, "Email tidak valid", http.StatusBadRequest)
		return
	}

	targetEmail := strings.ToLower(strings.TrimSpace(req.Email))
	code := generateOTPCode()

	otpMutex.Lock()
	otpStore[targetEmail] = OTPItem{
		Code:      code,
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}
	delete(verifStore, targetEmail)
	otpMutex.Unlock()

	go email.SendOTPEmail(targetEmail, code)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Kode OTP berhasil dikirim ke email Anda.",
	})
}

// VerifyOTPHandler verifies the 6-digit OTP code
func VerifyOTPHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email string `json:"email"`
		Code  string `json:"code"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Payload tidak valid", http.StatusBadRequest)
		return
	}

	targetEmail := strings.ToLower(strings.TrimSpace(req.Email))
	inputCode := strings.TrimSpace(req.Code)

	otpMutex.Lock()
	defer otpMutex.Unlock()

	item, exists := otpStore[targetEmail]
	if !exists {
		http.Error(w, "Kode OTP belum dikirim atau telah kedaluwarsa.", http.StatusBadRequest)
		return
	}

	if time.Now().After(item.ExpiresAt) {
		delete(otpStore, targetEmail)
		http.Error(w, "Kode OTP telah kedaluwarsa. Silakan minta kode baru.", http.StatusBadRequest)
		return
	}

	if item.Code != inputCode {
		http.Error(w, "Kode OTP yang Anda masukkan salah.", http.StatusBadRequest)
		return
	}

	verifStore[targetEmail] = true
	delete(otpStore, targetEmail)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Email berhasil diverifikasi!",
	})
}

// IsEmailVerified checks if an email has completed OTP verification
func IsEmailVerified(email string) bool {
	targetEmail := strings.ToLower(strings.TrimSpace(email))
	otpMutex.RLock()
	defer otpMutex.RUnlock()
	return verifStore[targetEmail]
}
