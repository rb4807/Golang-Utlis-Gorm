package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
	"os"
	"github.com/rb4807/Golang-Utlis-Postgresql/auth"
	"github.com/rb4807/Golang-Utlis-Postgresql/db"
	"github.com/rb4807/Golang-Utlis-Postgresql/router"
)

func main() {
	// Initialize database
	database := db.InitDB()
	
	// Initialize auth tables
	if err := auth.InitDB(database); err != nil {
		log.Fatalf("Failed to initialize auth tables: %v", err)
	}

	// Get JWT secret from environment
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}

	// Initialize auth service
	authService, err := auth.NewService(auth.Config{
		JWTSecret:     jwtSecret,
		TokenDuration: 24 * time.Hour,
		DB:            database, // Use the correct field name (DB instead of DBConnection)
	})
	if err != nil {
		log.Fatalf("Failed to initialize auth service: %v", err)
	}


	// Set up routes with all middleware applied
	r := router.InitRoutes(authService)

	// Start server
	fmt.Println("Server starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", r))
}