package database

import (
	"database/sql"
	"errors"
	"log"
	"time"

	"komite-sekolah/models"
)

var (
	ErrUserNotFound = errors.New("User not found")
	ErrDuplicateUser = errors.New("user already exists")
)

func GetUserByUsername(username string) (*models.User, error) {
	user := &models.User{}
	err := DB.QueryRow(`
		SELECT id, username, COALESCE(nis, ''), COALESCE(virtual_account, ''), name, password, role, must_change_password, created_at, updated_at
		FROM users WHERE username = ?
	`, username).Scan(
		&user.ID, &user.Username, &user.NIS, &user.VirtualAccount, &user.Name, &user.Password,
		&user.Role, &user.MustChangePassword, &user.CreatedAt, &user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func GetUserByNIS(nis string) (*models.User, error) {
	user := &models.User{}
	err := DB.QueryRow(`
		SELECT id, COALESCE(username, ''), nis, COALESCE(virtual_account, ''), name, password, role, must_change_password, created_at, updated_at
		FROM users WHERE nis = ?
	`, nis).Scan(
		&user.ID, &user.Username, &user.NIS, &user.VirtualAccount, &user.Name, &user.Password,
		&user.Role, &user.MustChangePassword, &user.CreatedAt, &user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		log.Printf("GetUserByNIS: no rows for nis=%q", nis)
		return nil, ErrUserNotFound
	}
	if err != nil {
		log.Printf("GetUserByNIS: query error for nis=%q: %v", nis, err)
		return nil, err
	}
	log.Printf("GetUserByNIS: found user id=%d nis=%q name=%q", user.ID, user.NIS, user.Name)
	return user, nil
}

func GetUserByID(id int64) (*models.User, error) {
	user := &models.User{}
	err := DB.QueryRow(`
		SELECT id, COALESCE(username, ''), COALESCE(nis, ''), COALESCE(virtual_account, ''), name, password, role, must_change_password, created_at, updated_at
		FROM users WHERE id = ?
	`, id).Scan(
		&user.ID, &user.Username, &user.NIS, &user.VirtualAccount, &user.Name, &user.Password,
		&user.Role, &user.MustChangePassword, &user.CreatedAt, &user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func CreateStudent(nis, virtual_account, name, hashedPassword string) (*models.User, error) {
	result, err := DB.Exec(`
		INSERT INTO users (nis, virtual_account, name, password, role, must_change_password)
		VALUES (?, ?, ?, ?, ?, ?)
	`, nis, virtual_account, name, hashedPassword, models.RoleStudent, 1)

	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()
	return GetUserByID(id)
}

func UpdatePassword(userID int64, hashedPassword string) error {
	_, err := DB.Exec(`
		UPDATE users 
		SET password = ?, must_change_password = 0, updated_at = ?
		WHERE id = ?
	`, hashedPassword, time.Now(), userID)
	return err
}

func ResetPassword(userID int64, hashedPassword string) error {
	_, err := DB.Exec(`
		UPDATE users 
		SET password = ?, must_change_password = 1, updated_at = ?
		WHERE id = ?
	`, hashedPassword, time.Now(), userID)
	return err
}

func GetAllStudents() ([]models.User, error) {
	rows, err := DB.Query(`
		SELECT id, COALESCE(username, ''), nis, COALESCE(virtual_account, ''), name, role, must_change_password, created_at, updated_at
		FROM users WHERE role = 'student'
		ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID, &user.Username, &user.NIS, &user.VirtualAccount, &user.Name,
			&user.Role, &user.MustChangePassword, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}


