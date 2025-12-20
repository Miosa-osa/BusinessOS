package middleware

import (
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/config"
)

// CORSMiddleware returns CORS middleware configured for Better Auth cookies
func CORSMiddleware(cfg *config.Config) gin.HandlerFunc {
	// Filter origins to only include valid http/https origins
	var validOrigins []string
	for _, origin := range cfg.AllowedOrigins {
		origin = strings.TrimSpace(origin)
		if strings.HasPrefix(origin, "http://") || strings.HasPrefix(origin, "https://") {
			validOrigins = append(validOrigins, origin)
		}
	}

	// If no valid origins, allow all in production (will be secured via other means)
	// or use localhost defaults in development
	if len(validOrigins) == 0 {
		if cfg.IsProduction() {
			// In production with no origins specified, allow all
			validOrigins = []string{"*"}
		} else {
			validOrigins = []string{"http://localhost:5173", "http://localhost:5174", "http://localhost:3000"}
		}
	}

	corsConfig := cors.Config{
		AllowOrigins:     validOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "Cookie"},
		ExposeHeaders:    []string{"X-Conversation-Id"},
		AllowCredentials: len(validOrigins) == 1 && validOrigins[0] != "*", // Can't use credentials with wildcard
		MaxAge:           86400,                                            // 24 hours
	}

	return cors.New(corsConfig)
}
