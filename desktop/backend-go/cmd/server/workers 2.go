package main

import (
	"context"
	"log"
	"time"

	"github.com/rhl/businessos-backend/internal/services"
)

// conversationSummaryJobConfig holds all parameters for the conversation summary background job.
type conversationSummaryJobConfig struct {
	interval    time.Duration
	batchSize   int
	maxMessages int
}

// startConversationSummaryJob starts the optional background job that keeps
// conversation_summaries fresh for context + semantic search.
// All closure variables are passed as explicit parameters.
func startConversationSummaryJob(ctx context.Context, svc *services.ConversationIntelligenceService, cfg conversationSummaryJobConfig) {
	go func() {
		t := time.NewTicker(cfg.interval)
		defer t.Stop()
		log.Printf("Conversation summary job enabled (interval=%s batch=%d maxMessages=%d)",
			cfg.interval, cfg.batchSize, cfg.maxMessages)
		for {
			select {
			case <-t.C:
				count, err := svc.BackfillStaleSummaries(ctx, cfg.batchSize, cfg.maxMessages, false)
				if err != nil {
					log.Printf("Conversation summary job error: %v", err)
					continue
				}
				if count > 0 {
					log.Printf("Conversation summary job updated %d conversations", count)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

// behaviorPatternsJobConfig holds all parameters for the behavior patterns background job.
type behaviorPatternsJobConfig struct {
	interval  time.Duration
	batchSize int
}

// startBehaviorPatternsJob starts the optional background job that detects behavior
// patterns and stores them as user_facts for explicit confirmation.
// All closure variables are passed as explicit parameters.
func startBehaviorPatternsJob(ctx context.Context, svc *services.LearningService, cfg behaviorPatternsJobConfig) {
	go func() {
		t := time.NewTicker(cfg.interval)
		defer t.Stop()
		log.Printf("Behavior patterns job enabled (interval=%s userBatch=%d)",
			cfg.interval, cfg.batchSize)
		for {
			select {
			case <-t.C:
				usersProcessed, factsUpserted, err := svc.BackfillRecentUsersBehaviorPatterns(ctx, cfg.batchSize)
				if err != nil {
					log.Printf("Behavior patterns job error: %v", err)
					continue
				}
				if usersProcessed > 0 || factsUpserted > 0 {
					log.Printf("Behavior patterns job processed %d users, upserted %d pattern facts",
						usersProcessed, factsUpserted)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

// appProfilerJobConfig holds all parameters for the app profiler sync background job.
type appProfilerJobConfig struct {
	interval  time.Duration
	batchSize int
}

// startAppProfilerSyncJob starts the optional background job that auto-syncs
// application profiles based on git HEAD or filesystem changes.
// All closure variables are passed as explicit parameters.
func startAppProfilerSyncJob(ctx context.Context, svc *services.AppProfilerService, cfg appProfilerJobConfig) {
	go func() {
		t := time.NewTicker(cfg.interval)
		defer t.Stop()
		log.Printf("App profiler auto-sync job enabled (interval=%s batch=%d)",
			cfg.interval, cfg.batchSize)
		for {
			select {
			case <-t.C:
				checked, refreshed, err := svc.SyncAutoProfiles(ctx, cfg.batchSize)
				if err != nil {
					log.Printf("App profiler auto-sync job error: %v", err)
					continue
				}
				if checked > 0 || refreshed > 0 {
					log.Printf("App profiler auto-sync checked %d profiles, refreshed %d", checked, refreshed)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}
