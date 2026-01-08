package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// Read SQL file
	sqlContent, err := os.ReadFile("../../create_auth_tables.sql")
	if err != nil {
		log.Fatal("Failed to read SQL file:", err)
	}

	// Connect to database
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL not set")
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer pool.Close()

	// Execute SQL
	_, err = pool.Exec(context.Background(), string(sqlContent))
	if err != nil {
		log.Fatal("Failed to execute SQL:", err)
	}

	fmt.Println("✅ Auth tables created successfully!")
}
