package auth

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func getEnvOrDefault(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

// Configurable credentials and signing key
var AdminEmail = getEnvOrDefault("ADMIN_EMAIL", "admin@gelora.go.id")
var AdminPassword = getEnvOrDefault("ADMIN_PASSWORD", "password123")
var JWTSecret = []byte(getEnvOrDefault("JWT_SECRET", "siwaris_gelora_secret_key_2026"))

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// GenerateJWT generates a token for the authenticated user
func GenerateJWT(username string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(JWTSecret)
}

// VerifyJWT parses and validates the token
func VerifyJWT(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return JWTSecret, nil
	})

	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

type contextKey string
const ContextUserKey contextKey = "user"

// AuthMiddleware protects admin routes
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Allow Preflight CORS requests
		if r.Method == http.MethodOptions {
			next(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized: No token provided", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Unauthorized: Invalid token format", http.StatusUnauthorized)
			return
		}

		claims, err := VerifyJWT(parts[1])
		if err != nil {
			http.Error(w, "Unauthorized: Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Inject claims into context
		ctx := context.WithValue(r.Context(), ContextUserKey, claims)
		next(w, r.WithContext(ctx))
	}
}
