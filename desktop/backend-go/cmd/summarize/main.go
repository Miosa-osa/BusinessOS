package main

import (
	"context"
	"flag"
	"log"

	"github.com/rhl/businessos-backend/internal/config"
	"github.com/rhl/businessos-backend/internal/database"
	"github.com/rhl/businessos-backend/internal/services"
)

func main() {
	var (
		limit       = flag.Int("limit", 50, "Max conversations to process")
		maxMessages = flag.Int("max-messages", 200, "Max messages per conversation to include (most recent N)")
		force       = flag.Bool("force", false, "Process conversations even if summaries look up-to-date")
	)
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config load: %v", err)
	}

	pool, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer database.Close()

	embed := services.NewEmbeddingService(pool, cfg.OllamaLocalURL)
	intel := services.NewConversationIntelligenceService(pool, embed)

	ctx := context.Background()
	count, err := intel.BackfillStaleSummaries(ctx, *limit, *maxMessages, *force)
	if err != nil {
		log.Fatalf("backfill: %v", err)
	}
	log.Printf("done: %d conversations analyzed", count)
}
