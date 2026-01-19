package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"komite-sekolah/database"
	"komite-sekolah/models"
)

// GetMyPaymentHistory returns payment history for the logged-in user
func GetMyPaymentHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	} 

	// Get user info for virtual account
	user, err := database.GetUserByID(userID)
	if err != nil {
		respondError(w, http.StatusNotFound, "User not found")
		return
	} 

	// Get payments
	payments, err := database.GetPaymentsByUserID(userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch payments")
		return
	} 

	virtualAccount := user.VirtualAccount

	response := models.PaymentHistoryResponse{
		VirtualAccount: virtualAccount,
		Payments:       payments,
		User:           user,
	}

	// Handle nil payments
	if response.Payments == nil {
		response.Payments = []models.Payment{}
	}

	respondJSON(w, http.StatusOK, response)
}

// GetAllPayments returns all payments (admin only)
func GetAllPayments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	role, ok := r.Context().Value("user_role").(models.UserRole)
	if !ok || role != models.RoleAdmin {
		respondError(w, http.StatusForbidden, "Admin access required")
		return
	}

	payments, err := database.GetAllPayments()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch payments")
		return
	}

	// Handle nil payments
	if payments == nil {
		payments = []models.Payment{}
	}

	respondJSON(w, http.StatusOK, payments)
}

// GetPaymentsByNIS returns payments for a specific student identified by NIS (admin only)
func GetPaymentsByNIS(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	role, ok := r.Context().Value("user_role").(models.UserRole)
	if !ok || role != models.RoleAdmin {
		respondError(w, http.StatusForbidden, "Admin access required")
		return
	}

	nis := strings.TrimSpace(r.URL.Query().Get("nis"))
	log.Printf("AdminGetPaymentsByNIS called, nis=%q", nis)
	if nis == "" {
		respondError(w, http.StatusBadRequest, "NIS is required")
		return
	}

	user, err := database.GetUserByNIS(nis)
	if err != nil {
		if err == database.ErrUserNotFound {
			log.Printf("GetPaymentsByNIS: user not found for nis=%q", nis)
			respondError(w, http.StatusNotFound, "User not found")
			return
		}
		log.Printf("GetPaymentsByNIS: error fetching user by nis=%q: %v", nis, err)
		respondError(w, http.StatusInternalServerError, "Failed to fetch user")
		return
	}

	userID := user.ID
	payments, err := database.GetPaymentsByUserIDWithUser(userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch payments")
		return
	}

	// Handle nil payments
	if payments == nil {
		payments = []models.Payment{}
	}

	response := models.PaymentHistoryResponse{
		VirtualAccount: user.VirtualAccount,
		Payments:       payments,
		User:           user,
	}

	respondJSON(w, http.StatusOK, response)
}

// GetPaymentsByUser returns payments for a specific user (admin only)
func GetPaymentsByUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	role, ok := r.Context().Value("user_role").(models.UserRole)
	if !ok || role != models.RoleAdmin {
		respondError(w, http.StatusForbidden, "Admin access required")
		return
	}

	// Get user_id from query parameter
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		respondError(w, http.StatusBadRequest, "user_id is required")
		return
	} 

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user_id")
		return
	} 

	payments, err := database.GetPaymentsByUserIDWithUser(userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch payments")
		return
	}

	// Handle nil payments
	if payments == nil {
		payments = []models.Payment{}
	}

	respondJSON(w, http.StatusOK, payments)
}

// CreatePayment creates a new payment record (admin only)
func CreatePayment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	role, ok := r.Context().Value("user_role").(models.UserRole)
	if !ok || role != models.RoleAdmin {
		respondError(w, http.StatusForbidden, "Admin access required")
		return
	}

	var req models.CreatePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if req.UserID == 0 {
		respondError(w, http.StatusBadRequest, "user_id is required")
		return
	}
	if req.Tanggal == "" {
		respondError(w, http.StatusBadRequest, "Tanggal is required")
		return
	}
	if req.Nominal <= 0 {
		respondError(w, http.StatusBadRequest, "Nominal must be greater than 0")
		return
	}


	// Verify user exists
	_, err := database.GetUserByID(req.UserID)
	if err != nil {
		respondError(w, http.StatusNotFound, "User not found")
		return
	}

	payment, err := database.CreatePayment(req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create payment: "+err.Error())
		return
	} 

	respondJSON(w, http.StatusCreated, payment)
}

// DeletePayment deletes a payment record (admin only)
func DeletePayment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	role, ok := r.Context().Value("user_role").(models.UserRole)
	if !ok || role != models.RoleAdmin {
		respondError(w, http.StatusForbidden, "Admin access required")
		return
	}

	// Get payment_id from query parameter
	paymentIDStr := r.URL.Query().Get("payment_id")
	if paymentIDStr == "" {
		respondError(w, http.StatusBadRequest, "payment_id is required")
		return
	}

	paymentID, err := strconv.ParseInt(paymentIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid payment_id")
		return
	}

	// Verify payment exists
	_, err = database.GetPaymentByID(paymentID)
	if err != nil {
		respondError(w, http.StatusNotFound, "Payment not found")
		return
	}

	if err := database.DeletePayment(paymentID); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to delete payment")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Payment deleted successfully"})
}

// UpdatePayment updates an existing payment (admin only)
func UpdatePayment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	role, ok := r.Context().Value("user_role").(models.UserRole)
	if !ok || role != models.RoleAdmin {
		respondError(w, http.StatusForbidden, "Admin access required")
		return
	}

	var req models.UpdatePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	log.Printf("UpdatePayment called, req=%+v, user_role=%v", req, role)

	if req.ID == 0 {
		respondError(w, http.StatusBadRequest, "payment_id is required")
		return
	}

	if req.Tanggal == nil {
		respondError(w, http.StatusBadRequest, "Tanggal is required")
		return
	}

	if req.Nominal == nil {
		respondError(w, http.StatusBadRequest, "Nominal must be greater than 0")
		return
	}

	// Verify payment exists
	_, err := database.GetPaymentByID(req.ID)
	if err != nil {
		if err == database.ErrPaymentNotFound {
			respondError(w, http.StatusNotFound, "Payment not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to fetch payment")
		return
	}

	// Basic validation: require at least one field to update (tanggal, nominal, or keterangan)
	if req.Tanggal == nil && req.Nominal == nil && req.Keterangan == nil {
		respondError(w, http.StatusBadRequest, "No fields to update")
		return
	}

	updated, err := database.UpdatePayment(req.ID, req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update payment: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, updated)
}
