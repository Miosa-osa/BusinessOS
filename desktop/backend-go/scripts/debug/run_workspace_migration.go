package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Read SQL file
	sqlContent, err := os.ReadFile("internal/database/migrations/026_workspaces_and_roles.sql")
	if err != nil {
		log.Fatal("Failed to read SQL file:", err)
	}

	// Connect to database
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL not set")
	}

	fmt.Println("📡 Connecting to database...")

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer pool.Close()

	fmt.Println("🚀 Running workspace migration...")

	// Execute SQL
	_, err = pool.Exec(context.Background(), string(sqlContent))
	if err != nil {
		log.Fatal("Failed to execute SQL:", err)
	}

	fmt.Println("✅ Workspace tables created successfully!")
	fmt.Println("\n📊 Created tables:")
	fmt.Println("  - workspaces")
	fmt.Println("  - workspace_roles")
	fmt.Println("  - workspace_members")
	fmt.Println("  - user_workspace_profiles")
	fmt.Println("  - workspace_memories")
	fmt.Println("  - project_members")
	fmt.Println("  - role_permissions")
	fmt.Println("\n🔧 Functions:")
	fmt.Println("  - seed_default_workspace_roles()")
	fmt.Println("\n✅ Migration complete!")
}
