package models

import "time"

type Payment struct {
	ID         int64    `json:"id"`
	UserID     int64    `json:"user_id"`
	User       *User    `json:"user,omitempty"`       // Populated when joining with users table
	Tanggal    string   `json:"tanggal"`              // Payment date (YYYY-MM-DD)
	Nominal    int64    `json:"nominal"`              // Amount in Rupiah
	Keterangan string   `json:"keterangan,omitempty"` // Description/notes
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type CreatePaymentRequest struct {
	UserID     int64  `json:"user_id"`
	Tanggal    string `json:"tanggal"`
	Nominal    int64  `json:"nominal"`
	Keterangan string `json:"keterangan,omitempty"`
}

type UpdatePaymentRequest struct {
	ID         int64   `json:"payment_id"`
	Tanggal    *string `json:"tanggal"`
	Nominal    *int64  `json:"nominal"`
	Keterangan *string `json:"keterangan,omitempty"`
}


type PaymentSummary struct {
	TotalTagihan    int64 `json:"total_tagihan"`
	TotalPembayaran int64 `json:"total_pembayaran"`
	SisaTagihan     int64 `json:"sisa_tagihan"`
	JumlahTransaksi int   `json:"jumlah_transaksi"`
}

type PaymentHistoryResponse struct {
	VirtualAccount string          `json:"virtual_account"`
	Summary        PaymentSummary  `json:"summary"`
	Payments       []Payment       `json:"payments"`
	User           *User           `json:"user,omitempty"`
}

