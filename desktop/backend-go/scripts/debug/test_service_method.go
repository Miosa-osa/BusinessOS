package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/services"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	workspaceID := uuid.MustParse("064e8e2a-5d3e-4d00-8492-df3628b1ec96")
	userID := "ZVtQRaictVbO9lN0p-csSA"

	fmt.Println("=== TESTING SERVICE METHOD ===")
	fmt.Printf("Workspace: %s\n", workspaceID)
	fmt.Printf("User: %s\n\n", userID)

	// Create the service
	service := services.NewMemoryHierarchyService(pool)

	// Call GetAccessibleMemories (exactly as the handler does)
	memories, err := service.GetAccessibleMemories(ctx, workspaceID, userID, nil, 20)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
	} else {
		fmt.Printf("Success! Retrieved %d memories:\n", len(memories))
		for i, mem := range memories {
			fmt.Printf("  %d. %s (type=%s, visibility=%s)\n", i+1, mem.Title, mem.MemoryType, mem.Visibility)
		}
	}
}
