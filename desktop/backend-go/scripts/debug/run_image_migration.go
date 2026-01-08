package main

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// Database connection string from .env
	dbURL := "postgres://postgres:Lunivate69420@db.fuqhjbgbjamtxcdphjpp.supabase.co:5432/postgres?connect_timeout=30"

	// Connect to database
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer pool.Close()

	log.Println("Connected to database successfully!")

	// Check if table already exists
	var exists bool
	err = pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT FROM pg_tables
			WHERE schemaname = 'public'
			AND tablename = 'image_embeddings'
		);
	`).Scan(&exists)

	if err != nil {
		log.Fatalf("Error checking table existence: %v\n", err)
	}

	if exists {
		log.Println("✅ Table 'image_embeddings' already exists - skipping migration")
		return
	}

	log.Println("Table 'image_embeddings' does not exist - running migration...")

	// Read migration file
	migrationSQL, err := os.ReadFile("internal/database/migrations/025_image_embeddings.sql")
	if err != nil {
		log.Fatalf("Error reading migration file: %v\n", err)
	}

	// Execute migration
	log.Println("Executing migration 025_image_embeddings.sql...")
	_, err = pool.Exec(ctx, string(migrationSQL))
	if err != nil {
		log.Fatalf("Error executing migration: %v\n", err)
	}

	log.Println("✅ Migration completed successfully!")

	// Verify table was created
	var count int
	err = pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM pg_tables
		WHERE schemaname = 'public'
		AND tablename LIKE 'image_%';
	`).Scan(&count)

	if err != nil {
		log.Printf("Warning: Could not verify table creation: %v\n", err)
	} else {
		log.Printf("✅ Created %d image-related tables", count)
	}
}
