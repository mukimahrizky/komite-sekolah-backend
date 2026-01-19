package main

import (
	"fmt"
	"log"
	"net/http"

	"komite-sekolah/config"
	"komite-sekolah/database"
	"komite-sekolah/handlers"
	"komite-sekolah/middleware"
)

func main() {
	// Load configuration from .env file
	config.Load()

	// Initialize database
	database.Init()
	defer database.Close()

	// Public routes (no auth required)
	http.HandleFunc("/", middleware.CORS(homeHandler))
	http.HandleFunc("/health", middleware.CORS(healthHandler))

	// Auth routes
	http.HandleFunc("/api/auth/admin/login", middleware.CORS(handlers.LoginAdmin))
	http.HandleFunc("/api/auth/student/login", middleware.CORS(handlers.LoginStudent))
	http.HandleFunc("/api/auth/change-password", middleware.CORS(middleware.AuthMiddleware(handlers.ChangePassword)))

	// Admin routes (protected)
	http.HandleFunc("/api/admin/students", middleware.CORS(middleware.AdminOnly(handleStudents)))
	http.HandleFunc("/api/admin/students/reset-password", middleware.CORS(middleware.AdminOnly(handlers.ResetStudentPassword)))

	// Payment routes (student - own payments)
	http.HandleFunc("/api/payments/my-history", middleware.CORS(middleware.AuthMiddleware(handlers.GetMyPaymentHistory)))

	// Payment routes (admin only)
	http.HandleFunc("/api/admin/payments", middleware.CORS(middleware.AdminOnly(handleAdminPayments)))
	http.HandleFunc("/api/admin/payments/by-user", middleware.CORS(middleware.AdminOnly(handlers.GetPaymentsByUser)))
	http.HandleFunc("/api/admin/payments/by-nis", middleware.CORS(middleware.AdminOnly(handlers.GetPaymentsByNIS)))
	http.HandleFunc("/api/admin/payments/delete", middleware.CORS(middleware.AdminOnly(handlers.DeletePayment)))
	http.HandleFunc("/api/admin/payments/edit", middleware.CORS(middleware.AdminOnly(handlers.UpdatePayment)))

	port := ":" + config.AppConfig.ServerPort
	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"message": "Komite Sekolah API", "version": "1.0.0"}`)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status": "OK"}`)
}

// handleStudents routes GET and POST for /api/admin/students
func handleStudents(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handlers.GetStudents(w, r)
	case http.MethodPost:
		handlers.CreateStudent(w, r)
	default:
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
	} 
}

// handleAdminPayments routes GET and POST for /api/admin/payments
func handleAdminPayments(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handlers.GetAllPayments(w, r)
	case http.MethodPost:
		handlers.CreatePayment(w, r)
	default:
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
	} 
}
