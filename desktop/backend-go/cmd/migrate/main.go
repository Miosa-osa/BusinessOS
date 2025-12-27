package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env
	godotenv.Load()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL not set")
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close(ctx)

	// Add share_calendar column
	_, err = conn.Exec(ctx, `
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'team_members' AND column_name = 'share_calendar'
			) THEN
				ALTER TABLE team_members ADD COLUMN share_calendar BOOLEAN DEFAULT FALSE;
			END IF;
		END $$;
	`)
	if err != nil {
		log.Printf("Warning adding share_calendar: %v", err)
	} else {
		fmt.Println("✓ share_calendar column OK")
	}

	// Add calendar_user_id column
	_, err = conn.Exec(ctx, `
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'team_members' AND column_name = 'calendar_user_id'
			) THEN
				ALTER TABLE team_members ADD COLUMN calendar_user_id VARCHAR(255);
			END IF;
		END $$;
	`)
	if err != nil {
		log.Printf("Warning adding calendar_user_id: %v", err)
	} else {
		fmt.Println("✓ calendar_user_id column OK")
	}

	fmt.Println("Migration complete!")
}
