package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/database/sqlc"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:Lunivate69420@db.fuqhjbgbjamtxcdphjpp.supabase.co:5432/postgres"
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer pool.Close()

	queries := sqlc.New(pool)

	// Test with the actual user_id from the error
	testUserID := "ZVtQRaictVbO9lN0p-csSA"
	testConversationID := uuid.New()
	testContent := "This is a test thinking trace to verify the fix"
	model := "claude-sonnet-4.5"
	thinkingTokens := int32(50)
	stepNumber := int32(1)

	fmt.Println("Testing thinking trace insertion...")
	fmt.Printf("User ID: %s\n", testUserID)
	fmt.Printf("Conversation ID: %s\n", testConversationID)
	fmt.Println()

	// Create thinking trace
	trace, err := queries.CreateThinkingTrace(ctx, sqlc.CreateThinkingTraceParams{
		UserID:         testUserID,
		ConversationID: pgtype.UUID{Bytes: testConversationID, Valid: true},
		MessageID:      pgtype.UUID{Valid: false},
		ThinkingContent: testContent,
		ThinkingType: sqlc.NullThinkingtype{
			Thinkingtype: sqlc.ThinkingtypeAnalysis,
			Valid:        true,
		},
		StepNumber: &stepNumber,
		StartedAt: pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		},
		ThinkingTokens:      &thinkingTokens,
		ModelUsed:           &model,
		ReasoningTemplateID: pgtype.UUID{Valid: false},
		Metadata:            []byte("{}"),
	})

	if err != nil {
		log.Fatalf("❌ Failed to save thinking trace: %v", err)
	}

	fmt.Println("✅ Successfully saved thinking trace!")
	fmt.Printf("   Trace ID: %s\n", trace.ID)
	fmt.Printf("   User ID: %s\n", trace.UserID)
	fmt.Printf("   Content length: %d chars\n", len(trace.ThinkingContent))
	fmt.Printf("   Tokens: %d\n", *trace.ThinkingTokens)
	fmt.Println()

	// Cleanup - delete the test trace
	err = queries.DeleteThinkingTrace(ctx, sqlc.DeleteThinkingTraceParams{
		ID:     trace.ID,
		UserID: testUserID,
	})
	if err != nil {
		log.Printf("⚠️  Warning: Failed to cleanup test trace: %v", err)
	} else {
		fmt.Println("🧹 Cleaned up test trace")
	}

	fmt.Println("\n✅ Test completed successfully! The UUID type mismatch is fixed.")
}
