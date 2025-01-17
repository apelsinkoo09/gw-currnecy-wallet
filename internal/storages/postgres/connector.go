package postgres

import (
	"database/sql"
	"fmt"
	"gw-currncy-wallet/pkg/env"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func Connection() (*sql.DB, error) {
	err := env.LoadConfig("../config.env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	USERNAME_DB := os.Getenv("USERNAME_DB")
	PASSWORD_DB := os.Getenv("PASSWORD_DB")
	DATABASE := os.Getenv("DATABASE")
	SSL := os.Getenv("SSL")
	HOST_DB := os.Getenv("HOST_DB")
	PORT_DB := os.Getenv("PORT_DB")

	if USERNAME_DB == "" || PASSWORD_DB == "" || DATABASE == "" || SSL == "" || HOST_DB == "" || PORT_DB == "" {
		return nil, fmt.Errorf("missing required environment variables")
	}

	conString := fmt.Sprintf(
		"host=%s port=%s user=%s sslmode=%s password=%s dbname=%s",
		HOST_DB, PORT_DB, USERNAME_DB, SSL,
		PASSWORD_DB, DATABASE,
	)

	db, err := sql.Open("postgres", conString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("database is unreachable: %v", err)
	}
	fmt.Println("Successfully connected to PostgreSQL!")
	return db, nil
}
