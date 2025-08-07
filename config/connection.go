package config

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"github.com/syrlramadhan/database-anggota-api/helper"
)

func ConnectToDatabase() (db *sql.DB, err error) {
	var w http.ResponseWriter
	err = godotenv.Load()
	if err != nil {
		helper.WriteJSONError(w, http.StatusInternalServerError, err.Error())
		return nil, err
	}

	dbName := os.Getenv("DB_NAME")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")

	mysql := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbPass, dbHost, dbPort, dbName)
	db, err = sql.Open("mysql", mysql)
	if err != nil {
		helper.WriteJSONError(w, http.StatusInternalServerError, err.Error())
		return db, err
	}

	err = db.Ping()
	if err != nil {
		helper.WriteJSONError(w, http.StatusInternalServerError, err.Error())
		return db, err
	}

	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	return db, nil
}
