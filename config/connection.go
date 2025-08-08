package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func ConnectToDatabase() (*sql.DB, error) {
	// Load .env
	err := godotenv.Load()
	if err != nil {
		log.Println("Gagal load .env:", err)
		return nil, err
	}

	// Ambil konfigurasi dari environment
	dbName := os.Getenv("DB_NAME")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")

	// Sertakan parseTime=true agar DATE/DATEIME terbaca sebagai time.Time
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", dbUser, dbPass, dbHost, dbPort, dbName)

	// Buka koneksi
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Println("Gagal buka koneksi DB:", err)
		return nil, err
	}

	// Tes koneksi
	err = db.Ping()
	if err != nil {
		log.Println("DB Ping gagal:", err)
		return nil, err
	}

	// Set konfigurasi koneksi pool
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	return db, nil
}
