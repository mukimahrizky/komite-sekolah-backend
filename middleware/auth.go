package middleware

import (
	"context"
	"net/http"
	"strings"

	"komite-sekolah/config"
	"komite-sekolah/models"

	"github.com/golang-jwt/jwt/v5"
)

func getJWTSecret() []byte {
	return []byte(config.AppConfig.JWTSecret)
}

type Claims struct {
	UserID int64           `json:"user_id"`
	Role   models.UserRole `json:"role"`
	jwt.RegisteredClaims
}

// AuthMiddleware validates JWT token and adds user info to context
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
		http.Error(w, `{"error": "Header Authorization diperlukan"}`, http.StatusUnauthorized)
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		http.Error(w, `{"error": "Format Authorization tidak valid"}`, http.StatusUnauthorized)
		return
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return getJWTSecret(), nil
	})

	if err != nil || !token.Valid {
		http.Error(w, `{"error": "Token tidak valid"}`, http.StatusUnauthorized)
		}

		// Add user info to context
		ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "user_role", claims.Role)

		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// AdminOnly middleware ensures only admin can access the route
func AdminOnly(next http.HandlerFunc) http.HandlerFunc {
	return AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		role, ok := r.Context().Value("user_role").(models.UserRole)
		if !ok || role != models.RoleAdmin {
			http.Error(w, `{"error": "Admin access required"}`, http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// CORS middleware for frontend requests
func CORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get allowed origins from config
		allowedOrigins := config.AppConfig.AllowedOrigins
		environment := config.AppConfig.Environment
		requestOrigin := r.Header.Get("Origin")
		
		// Determine which origin to allow
		var allowedOrigin string
		
		if allowedOrigins == "*" {
			// Reject * in all environments for security
			if environment == "production" {
				http.Error(w, `{"error": "CORS misconfiguration: ALLOWED_ORIGINS cannot be * in production"}`, http.StatusInternalServerError)
				return
			}
			// Even in development, warn but allow (for backward compatibility)
			allowedOrigin = "*"
		} else if allowedOrigins != "" {
			// Check if request origin is in the allowed list
			allowedList := strings.Split(allowedOrigins, ",")
			for _, origin := range allowedList {
				origin = strings.TrimSpace(origin)
				if origin == requestOrigin {
					allowedOrigin = origin
					break
				}
			}
			
			// If no match found and we have a request origin
			if allowedOrigin == "" && requestOrigin != "" {
				// Always reject unauthorized origins (both dev and prod)
				http.Error(w, `{"error": "CORS: Origin not allowed"}`, http.StatusForbidden)
				return
			}
		}
		
		// Set CORS headers
		if allowedOrigin != "" {
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		// Only set credentials if not using wildcard
		if allowedOrigin != "*" {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	}
}


