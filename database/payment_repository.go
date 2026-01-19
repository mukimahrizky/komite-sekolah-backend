package database

import (
	"database/sql"
	"errors"
	"log"
	"strings"

	"komite-sekolah/models"
)

var (
	ErrPaymentNotFound = errors.New("Payment not found")
)

// CreatePayment creates a new payment record
func CreatePayment(req models.CreatePaymentRequest) (*models.Payment, error) {
	result, err := DB.Exec(`
		INSERT INTO payments (user_id, tanggal, nominal, keterangan)
		VALUES (?, ?, ?, ?)
	`, req.UserID, req.Tanggal, req.Nominal, req.Keterangan)

	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()
	return GetPaymentByID(id)
}

// GetPaymentByID retrieves a payment by ID
func GetPaymentByID(id int64) (*models.Payment, error) {
	payment := &models.Payment{}
	err := DB.QueryRow(`
		SELECT id, user_id, tanggal, nominal, COALESCE(keterangan, ''), created_at, updated_at
		FROM payments WHERE id = ?
	`, id).Scan(
		&payment.ID, &payment.UserID, &payment.Tanggal, &payment.Nominal,
		&payment.Keterangan,
		&payment.CreatedAt, &payment.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrPaymentNotFound
	}
	if err != nil {
		return nil, err
	}
	return payment, nil
}

// GetPaymentsByUserID retrieves all payments for a specific user
func GetPaymentsByUserID(userID int64) ([]models.Payment, error) {
	rows, err := DB.Query(`
		SELECT p.id, p.user_id, p.tanggal, p.nominal, COALESCE(p.keterangan, ''), p.created_at, p.updated_at
		FROM payments p
		WHERE p.user_id = ?
		ORDER BY p.tanggal DESC, p.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []models.Payment
	for rows.Next() {
		var payment models.Payment
		err := rows.Scan(
			&payment.ID, &payment.UserID, &payment.Tanggal, &payment.Nominal,
			&payment.Keterangan,
			&payment.CreatedAt, &payment.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		payments = append(payments, payment)
	}
	return payments, nil
}

// GetPaymentsByUserIDWithUser retrieves all payments for a user with user info
func GetPaymentsByUserIDWithUser(userID int64) ([]models.Payment, error) {
	rows, err := DB.Query(`
		SELECT p.id, p.user_id, p.tanggal, p.nominal, COALESCE(p.keterangan, ''), p.created_at, p.updated_at,
			   u.id, COALESCE(u.username, ''), COALESCE(u.nis, ''), u.name, u.role
		FROM payments p
		JOIN users u ON p.user_id = u.id
		WHERE p.user_id = ?
		ORDER BY p.tanggal DESC, p.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []models.Payment
	count := 0
	for rows.Next() {
		var payment models.Payment
		var user models.User
		err := rows.Scan(
			&payment.ID, &payment.UserID, &payment.Tanggal, &payment.Nominal,
			&payment.Keterangan,
			&payment.CreatedAt, &payment.UpdatedAt,
			&user.ID, &user.Username, &user.NIS, &user.Name, &user.Role,
		)
		if err != nil {
			return nil, err
		}
		payment.User = &user
		payments = append(payments, payment)
		count++
	}
	log.Printf("GetPaymentsByUserIDWithUser: userID=%d returns %d payments", userID, count)
	return payments, nil
}

// GetAllPayments retrieves all payments (admin only)
func GetAllPayments() ([]models.Payment, error) {
	rows, err := DB.Query(`
		SELECT p.id, p.user_id, p.tanggal, p.nominal, COALESCE(p.keterangan, ''), p.created_at, p.updated_at,
			   u.id, COALESCE(u.username, ''), COALESCE(u.nis, ''), u.name, u.role
		FROM payments p
		JOIN users u ON p.user_id = u.id
		ORDER BY p.tanggal DESC, p.created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []models.Payment
	for rows.Next() {
		var payment models.Payment
		var user models.User
		err := rows.Scan(
			&payment.ID, &payment.UserID, &payment.Tanggal, &payment.Nominal,
			&payment.Keterangan,
			&payment.CreatedAt, &payment.UpdatedAt,
			&user.ID, &user.Username, &user.NIS, &user.Name, &user.Role,
		)
		if err != nil {
			return nil, err
		}
		payment.User = &user
		payments = append(payments, payment)
	}
	return payments, nil
}

// DeletePayment deletes a payment record
func DeletePayment(paymentID int64) error {
	_, err := DB.Exec(`DELETE FROM payments WHERE id = ?`, paymentID)
	return err
}

// UpdatePayment updates a payment record
func UpdatePayment(paymentID int64, req models.UpdatePaymentRequest) (*models.Payment, error) {
	var sets []string
	var args []interface{}

	if req.Tanggal != nil {
		sets = append(sets, "tanggal = ?")
		args = append(args, *req.Tanggal)
	}
	if req.Nominal != nil {
		sets = append(sets, "nominal = ?")
		args = append(args, *req.Nominal)
	}
	if req.Keterangan != nil {
		sets = append(sets, "keterangan = ?")
		args = append(args, *req.Keterangan)
	}

	if len(sets) == 0 {
		return nil, errors.New("No fields to update")
	}

	query := "UPDATE payments SET " + strings.Join(sets, ", ") + ", updated_at = CURRENT_TIMESTAMP WHERE id = ?"
	args = append(args, paymentID)

	log.Printf("UpdatePayment query=%q args=%v", query, args)
	_, err := DB.Exec(query, args...)
	if err != nil {
		log.Printf("UpdatePayment: exec error: %v", err)
		return nil, err
	}
	return GetPaymentByID(paymentID)
}

// GetPaymentSummaryByUserID calculates payment summary for a user
func GetPaymentSummaryByUserID(userID int64, totalTagihan int64) (*models.PaymentSummary, error) {
	var totalPembayaran int64
	var jumlahTransaksi int

	// Get total payments
	err := DB.QueryRow(`
		SELECT COALESCE(SUM(nominal), 0), COUNT(*)
		FROM payments 
		WHERE user_id = ?
	`, userID).Scan(&totalPembayaran, &jumlahTransaksi)
	if err != nil {
		return nil, err
	}

	return &models.PaymentSummary{
		TotalTagihan:    totalTagihan,
		TotalPembayaran: totalPembayaran,
		SisaTagihan:     totalTagihan - totalPembayaran,
		JumlahTransaksi: jumlahTransaksi,
	}, nil
}

// GetTotalPaymentsByUserID gets total of payments for a user
func GetTotalPaymentsByUserID(userID int64) (int64, error) {
	var total int64
	err := DB.QueryRow(`
		SELECT COALESCE(SUM(nominal), 0)
		FROM payments 
		WHERE user_id = ?
	`, userID).Scan(&total)
	return total, err
}

