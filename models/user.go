package models

import "time"

type UserRole string

const (
	RoleAdmin    UserRole = "admin"
	RoleStudent  UserRole = "student"
)

type User struct {
	ID                int64     `json:"id"`
	Username          string    `json:"username,omitempty"`  // For admin login
	NIS               string    `json:"nis,omitempty"`       // For student login (Nomor Induk Siswa)
	VirtualAccount    string    `json:"virtual_account,omitempty"`
	Name              string    `json:"name"`
	Password          string    `json:"-"`                   // Never expose in JSON
	Role              UserRole  `json:"role"`
	MustChangePassword bool     `json:"must_change_password"` // True for first login
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type LoginRequest struct {
	Username string `json:"username,omitempty"` // For admin
	NIS      string `json:"nis,omitempty"`      // For student
	Password string `json:"password"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type LoginResponse struct {
	Token              string `json:"token"`
	User               User   `json:"user"`
	MustChangePassword bool   `json:"must_change_password"`
}


