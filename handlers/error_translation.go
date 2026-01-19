package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
)

func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	translations := map[string]string{
		"Method not allowed": "Metode tidak diizinkan",
		"Unauthorized": "Tidak terautentikasi",
		"User not found": "Pengguna tidak ditemukan",
		"Failed to fetch payments": "Gagal mengambil riwayat pembayaran",
		"Failed to fetch payment summary": "Gagal mengambil ringkasan pembayaran",
		"Admin access required": "Akses admin diperlukan",
		"Invalid request body": "Isi permintaan tidak valid",
		"Username and password are required": "Username dan kata sandi diperlukan",
		"Invalid credentials": "Kredensial tidak valid",
		"Failed to generate token": "Gagal membuat token",
		"NIS and password are required": "NIS dan kata sandi diperlukan",
		"Old password and new password are required": "Kata sandi lama dan baru diperlukan",
		"New password must be at least 6 characters": "Kata sandi baru harus minimal 6 karakter",
		"user_id is required": "user_id diperlukan",
		"Invalid user_id": "user_id tidak valid",
		"payment_id is required": "payment_id diperlukan",
		"Invalid payment_id": "payment_id tidak valid",
		"Payment not found": "Pembayaran tidak ditemukan",
		"Failed to delete payment": "Gagal menghapus pembayaran",
		"nis is required": "NIS diperlukan",
		"Failed to fetch user": "Gagal mengambil data pengguna",
		"Failed to create payment: ": "Gagal membuat pembayaran: ",
		"Failed to update payment: ": "Gagal memperbarui pembayaran: ",
		"No fields to update": "Tidak ada field untuk diperbarui",
		"Old password is incorrect": "Kata sandi lama salah",
		"Failed to hash password": "Gagal mengenkripsi kata sandi",
		"Failed to update password": "Gagal memperbarui kata sandi",
		"Password changed successfully": "Kata sandi berhasil diubah",
		"Nominal must be greater than 0": "Nominal harus lebih besar dari 0",
		"Tanggal is required": "Tanggal diperlukan",
	}

	// Exact match translation
	if t, ok := translations[message]; ok {
		message = t
	} else {
		// Prefix replacement for messages like "Failed to create payment: <err>"
		for eng, indo := range translations {
			if strings.HasSuffix(eng, ": ") && strings.HasPrefix(message, eng) {
				message = strings.Replace(message, eng, indo, 1)
				break
			}
		}
	}

	respondJSON(w, status, map[string]string{"error": message})
}