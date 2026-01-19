package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"komite-sekolah/config"
	"komite-sekolah/database"
	"komite-sekolah/models"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func getJWTSecret() []byte {
	return []byte(config.AppConfig.JWTSecret)
}

type Claims struct {
	UserID int64           `json:"user_id"`
	Role   models.UserRole `json:"role"`
	jwt.RegisteredClaims
}

// LoginAdmin handles admin login with username & password
func LoginAdmin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	} 

	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Username == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "Username and password are required")
		return
	} 

	user, err := database.GetUserByUsername(req.Username)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	if user.Role != models.RoleAdmin {
		respondError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		respondError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	token, err := generateToken(user)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	respondJSON(w, http.StatusOK, models.LoginResponse{
		Token:              token,
		User:               *user,
		MustChangePassword: user.MustChangePassword,
	})
}

// LoginStudent handles student login with NIS & password
func LoginStudent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	} 

	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.NIS == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "NIS and password are required")
		return
	} 

	user, err := database.GetUserByNIS(req.NIS)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	if user.Role != models.RoleStudent {
		respondError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		respondError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	token, err := generateToken(user)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	respondJSON(w, http.StatusOK, models.LoginResponse{
		Token:              token,
		User:               *user,
		MustChangePassword: user.MustChangePassword,
	})
}

// ChangePassword handles password change for authenticated users
func ChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	} 

	// Get user from context (set by auth middleware)
	userID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	} 

	var req models.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	} 

	if req.OldPassword == "" || req.NewPassword == "" {
		respondError(w, http.StatusBadRequest, "Old password and new password are required")
		return
	} 

	if len(req.NewPassword) < 6 {
		respondError(w, http.StatusBadRequest, "New password must be at least 6 characters")
		return
	} 

	user, err := database.GetUserByID(userID)
	if err != nil {
		respondError(w, http.StatusNotFound, "User not found")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		respondError(w, http.StatusBadRequest, "Old password is incorrect")
		return
	} 

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	if err := database.UpdatePassword(userID, string(hashedPassword)); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update password")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Kata sandi berhasil diubah"})
}

func generateToken(user *models.User) (string, error) {
	claims := &Claims{
		UserID: user.ID,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getJWTSecret())
}


