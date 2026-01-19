package database

import (
	"database/sql"
	"fmt"
	"log"

	"komite-sekolah/config"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

var DB *sql.DB

func Init() {
	// Get MySQL connection details from config
	cfg := config.AppConfig

	// MySQL connection string format: user:password@tcp(host:port)/dbname
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)

	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	createTables()
	seedAdmin()
	log.Println("Database initialized successfully")
}

func createTables() {
	// Create tables first
	tableQueries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INT AUTO_INCREMENT PRIMARY KEY,
			username VARCHAR(255) UNIQUE,
			nis VARCHAR(255) UNIQUE,
			virtual_account VARCHAR(255) UNIQUE,
			name VARCHAR(255) NOT NULL,
			password VARCHAR(255) NOT NULL,
			role ENUM('admin', 'student') NOT NULL,
			must_change_password TINYINT(1) DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_users_username (username),
			INDEX idx_users_nis (nis)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`,
		`CREATE TABLE IF NOT EXISTS payments (
			id INT AUTO_INCREMENT PRIMARY KEY,
			user_id INT NOT NULL,
			tanggal DATE NOT NULL,
			nominal BIGINT NOT NULL,
			keterangan TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			INDEX idx_payments_user_id (user_id),
			INDEX idx_payments_tanggal (tanggal)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`,
	}

	for _, query := range tableQueries {
		_, err := DB.Exec(query)
		if err != nil {
			log.Fatal("Failed to create tables:", err)
		}
	}
}

// For development, recreate the database if you need schema changes (delete `komite_sekolah.db` and restart).

func seedAdmin() {
	// Check if admin exists
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'admin'").Scan(&count)
	if err != nil {
		log.Fatal("Failed to check admin:", err)
	}

	if count == 0 {
		// Create default admin (password: admin123)
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		if err != nil {
			log.Fatal("Failed to hash admin password:", err)
		}
		_, err = DB.Exec(`
			INSERT INTO users (username, virtual_account, name, password, role, must_change_password)
			VALUES (?, ?, ?, ?, ?, ?)
		`, "admin", "", "Administrator", string(hashedPassword), "admin", 0)
		if err != nil {
			log.Fatal("Failed to seed admin:", err)
		}
		log.Println("Default admin created (username: admin, password: admin123)")
	}
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}

