package services

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/config"
	"github.com/rhl/businessos-backend/internal/integrations/osa"
)

// OSASyncService handles synchronization between BusinessOS and OSA-5
// This is a stub implementation - full sync functionality will be added later
type OSASyncService struct {
	pool      *pgxpool.Pool
	osaClient *osa.Client
}

// NewOSASyncService creates a new OSA sync service (stub)
func NewOSASyncService(pool *pgxpool.Pool, cfg *config.Config) (*OSASyncService, error) {
	return &OSASyncService{
		pool: pool,
	}, nil
}

// TODO: Implement sync methods when database schema is ready:
// - SyncUser(ctx context.Context, userID uuid.UUID) error
// - SyncWorkspace(ctx context.Context, workspaceID uuid.UUID) error
// - ProcessOutbox(ctx context.Context) error
