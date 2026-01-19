package handlers

import (
	"encoding/json"
	"net/http"

	"komite-sekolah/database"
	"komite-sekolah/models"

	"golang.org/x/crypto/bcrypt"
)

type CreateStudentRequest struct {
	NIS      	   string `json:"nis"`
	VirtualAccount string `json:"virtual_account"`
	Name    	   string `json:"name"`
	Password 	   string `json:"password"` // Initial password given by admin
}

// CreateStudent creates a new student account (admin only)
func CreateStudent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Check if user is admin
	role, ok := r.Context().Value("user_role").(models.UserRole)
	if !ok || role != models.RoleAdmin {
		respondError(w, http.StatusForbidden, "Admin access required")
		return
	}

	var req CreateStudentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	} 

	if req.NIS == "" || req.VirtualAccount == "" || req.Name == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "NIS, virtual account, name, and password are required")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	user, err := database.CreateStudent(req.NIS, req.VirtualAccount, req.Name, string(hashedPassword))
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create student: "+err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, user)
}

// GetStudents returns all students (admin only)
func GetStudents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	} 

	role, ok := r.Context().Value("user_role").(models.UserRole)
	if !ok || role != models.RoleAdmin {
		respondError(w, http.StatusForbidden, "Admin access required")
		return
	}

	students, err := database.GetAllStudents()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch students")
		return
	}

	respondJSON(w, http.StatusOK, students)
}

type ResetPasswordRequest struct {
	UserID      int64  `json:"user_id"`
	NewPassword string `json:"new_password"`
}

// ResetStudentPassword resets a student's password (admin only)
// After reset, student must change password on next login
func ResetStudentPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	role, ok := r.Context().Value("user_role").(models.UserRole)
	if !ok || role != models.RoleAdmin {
		respondError(w, http.StatusForbidden, "Admin access required")
		return
	}

	var req ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.UserID == 0 || req.NewPassword == "" {
		respondError(w, http.StatusBadRequest, "User ID and new password are required")
		return
	}

	// Verify it's a student
	user, err := database.GetUserByID(req.UserID)
	if err != nil {
		respondError(w, http.StatusNotFound, "User not found")
		return
	} 

	if user.Role != models.RoleStudent {
		respondError(w, http.StatusBadRequest, "Can only reset student passwords")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	// Reset password and set must_change_password to true
	if err := database.ResetPassword(req.UserID, string(hashedPassword)); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to reset password")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Password reset successfully"})
}


