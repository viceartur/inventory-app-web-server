package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func ConnectToDB() (*sql.DB, error) {
	errEnv := godotenv.Load(".env")
	if errEnv != nil {
		log.Fatalf("Error loading .env file")
	}
	var (
		host     = os.Getenv("DB_HOST")
		port     = os.Getenv("DB_PORT")
		user     = os.Getenv("DB_USER")
		password = os.Getenv("DB_PASSWORD")
		dbname   = os.Getenv("DB_NAME")
	)

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, errors.New(err.Error())
	}

	if err = db.Ping(); err != nil {
		return nil, errors.New(err.Error())
	}

	return db, nil
}
