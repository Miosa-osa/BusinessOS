package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/cache"
	"github.com/rhl/businessos-backend/internal/database/sqlc"
	"github.com/rhl/businessos-backend/internal/middleware"
)

// ContextHandler handles context management operations
type ContextHandler struct {
	pool       *pgxpool.Pool
	queryCache *cache.QueryCache
}

// NewContextHandler creates a new ContextHandler
func NewContextHandler(pool *pgxpool.Pool, queryCache *cache.QueryCache) *ContextHandler {
	return &ContextHandler{pool: pool, queryCache: queryCache}
}

// RegisterContextRoutes registers all context routes on the given router group
func RegisterContextRoutes(api *gin.RouterGroup, h *ContextHandler, auth gin.HandlerFunc) {
	contexts := api.Group("/contexts")
	{
		// Public route (no auth)
		contexts.GET("/public/:shareId", h.GetPublicContext)

		// Protected routes
		protected := contexts.Group("")
		protected.Use(auth, middleware.RequireAuth())
		{
			protected.GET("", h.ListContexts)
			protected.POST("", h.CreateContext)
			protected.GET("/:id", h.GetContext)
			protected.PUT("/:id", h.UpdateContext)
			protected.PATCH("/:id/blocks", h.UpdateContextBlocks)
			protected.POST("/:id/share", h.ShareContext)
			protected.DELETE("/:id/share", h.UnshareContext)
			protected.POST("/:id/duplicate", h.DuplicateContext)
			protected.PATCH("/:id/archive", h.ArchiveContext)
			protected.PATCH("/:id/unarchive", h.UnarchiveContext)
			protected.DELETE("/:id", h.DeleteContext)
			protected.POST("/aggregate", h.AggregateContext)
		}
	}
}

// invalidateContextsCachePattern invalidates all cache entries for a user's contexts
func (h *ContextHandler) invalidateContextsCachePattern(ctx context.Context, userID string) {
	if h.queryCache == nil {
		return
	}

	pattern := fmt.Sprintf("contexts:user:%s:*", userID)
	if _, err := h.queryCache.DeleteByPattern(ctx, pattern); err != nil {
		slog.Warn("Failed to invalidate contexts cache",
			"user_id", userID,
			"pattern", pattern,
			"error", err)
	}
}

// stringToContextType converts a string to sqlc.Contexttype
func stringToContextType(t string) sqlc.Contexttype {
	typeMap := map[string]sqlc.Contexttype{
		"person":   sqlc.ContexttypePERSON,
		"business": sqlc.ContexttypeBUSINESS,
		"project":  sqlc.ContexttypePROJECT,
		"document": sqlc.ContexttypeDocument, // Use lowercase version for DB compatibility
		"custom":   sqlc.ContexttypeCUSTOM,
	}
	if enum, ok := typeMap[strings.ToLower(t)]; ok {
		return enum
	}
	return sqlc.ContexttypeCUSTOM
}

// generateShareID generates a random share ID
func generateShareID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
