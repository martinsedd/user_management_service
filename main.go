package main

import (
	"github.com/joho/godotenv"
	"log"
	"user-management-microservice/internal/db"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	dbConn, err := db.Connect()
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	db.RunMigrations(dbConn)
}
