package workers

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/services"
)

// AppGenerationWorker processes the app generation queue in the background
type AppGenerationWorker struct {
	pool     *pgxpool.Pool
	worker   *services.AppGenerationWorker
	interval time.Duration
	stopCh   chan struct{}
	logger   *slog.Logger
}

// NewAppGenerationWorker creates a new app generation background worker
func NewAppGenerationWorker(pool *pgxpool.Pool, logger *slog.Logger) *AppGenerationWorker {
	// Create OSA client adapter
	osaClient := services.NewOSAClientAdapter(pool)

	// Create app deployment service (enabled by default)
	var deployer *services.AppDeploymentService

	// Check if app deployment is explicitly disabled via environment variable
	if os.Getenv("DISABLE_APP_DEPLOYMENT") == "true" {
		logger.Warn("App deployment explicitly DISABLED - unset DISABLE_APP_DEPLOYMENT to enable")
	} else {
		workspacePath := os.Getenv("WORKSPACE_PATH")
		if workspacePath == "" {
			workspacePath = "/tmp/businessos-workspace"
		}
		deployer = services.NewAppDeploymentService(pool, logger, workspacePath)
		logger.Info("App deployment ENABLED", "workspace_path", workspacePath)
	}

	// Create the service worker
	serviceWorker := services.NewAppGenerationWorker(pool, osaClient, deployer, logger)

	return &AppGenerationWorker{
		pool:     pool,
		worker:   serviceWorker,
		interval: 5 * time.Second, // Process queue every 5 seconds (consistent with OSAQueueWorker)
		stopCh:   make(chan struct{}),
		logger:   logger,
	}
}

// Start begins processing the app generation queue
func (w *AppGenerationWorker) Start(ctx context.Context) {
	w.logger.Info("[AppGenerationWorker] Starting app generation queue processor...")

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	// Process immediately on start
	w.processQueue(ctx)

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("[AppGenerationWorker] Shutting down (context cancelled)")
			return
		case <-w.stopCh:
			w.logger.Info("[AppGenerationWorker] Shutting down (stop signal)")
			return
		case <-ticker.C:
			w.processQueue(ctx)
		}
	}
}

// Stop signals the worker to shutdown
func (w *AppGenerationWorker) Stop() {
	close(w.stopCh)
}

// processQueue processes pending items in the generation queue
func (w *AppGenerationWorker) processQueue(ctx context.Context) {
	// Process up to 5 items per tick to avoid long-running operations
	maxItemsPerTick := 5

	for i := 0; i < maxItemsPerTick; i++ {
		err := w.worker.ProcessQueue(ctx)
		if err != nil {
			// ProcessQueue returns nil when no items are pending
			// Only log actual errors, not "no items" conditions
			w.logger.Error("[AppGenerationWorker] Error processing queue",
				"error", err,
				"iteration", i,
			)
			// Stop processing on error
			break
		}
		// nil means item was processed successfully, continue to next
	}
}
