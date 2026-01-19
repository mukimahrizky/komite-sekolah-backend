package middleware

import (
	"context"
	"log"
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
		cfg := config.AppConfig
		requestOrigin := r.Header.Get("Origin")
		log.Println("Request Origin:", requestOrigin)
		log.Println("Allowed Origins from config:", cfg.AllowedOrigins)

		// Non-browser requests (curl, server-to-server) often have no Origin.
		// CORS is a browser-enforced policy, so we can skip origin checks here.
		if requestOrigin == "" {
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
			return
		}

		allowedOrigins := strings.TrimSpace(cfg.AllowedOrigins)
		if allowedOrigins == "" {
			http.Error(w, `{"error": "CORS misconfiguration: ALLOWED_ORIGINS is empty"}`, http.StatusInternalServerError)
			return
		}

		// Disallow wildcard in production.
		if allowedOrigins == "*" && cfg.Environment == "production" {
			http.Error(w, `{"error": "CORS misconfiguration: ALLOWED_ORIGINS cannot be * in production"}`, http.StatusInternalServerError)
			return
		}

		// Determine allowed origin (must echo back ONE origin, never a comma-separated list)
		allowedOrigin := ""
		if allowedOrigins == "*" {
			allowedOrigin = "*"
		} else {
			for _, origin := range strings.Split(allowedOrigins, ",") {
				origin = strings.TrimSpace(origin)
				if origin != "" && origin == requestOrigin {
					allowedOrigin = origin
					break
				}
			}
		}

		if allowedOrigin == "" {
			http.Error(w, `{"error": "CORS: Origin not allowed"}`, http.StatusForbidden)
			return
		}

		// Set CORS headers
		w.Header().Set("Vary", "Origin")
		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		// Credentials are only valid when not using wildcard
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


