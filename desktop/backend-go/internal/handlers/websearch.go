package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/services"
)

// WebSearch performs a web search
func (h *Handlers) WebSearch(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'q' is required"})
		return
	}

	maxResults := 10
	if max := c.Query("max"); max != "" {
		if parsed, err := strconv.Atoi(max); err == nil && parsed > 0 && parsed <= 20 {
			maxResults = parsed
		}
	}

	searchService := services.NewWebSearchService()
	results, err := searchService.Search(c.Request.Context(), query, maxResults)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Search failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, results)
}

// WebSearchWithContext performs a web search and returns formatted context
func (h *Handlers) WebSearchWithContext(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'q' is required"})
		return
	}

	maxResults := 5
	if max := c.Query("max"); max != "" {
		if parsed, err := strconv.Atoi(max); err == nil && parsed > 0 && parsed <= 10 {
			maxResults = parsed
		}
	}

	searchService := services.NewWebSearchService()
	results, err := searchService.Search(c.Request.Context(), query, maxResults)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Search failed",
			"details": err.Error(),
		})
		return
	}

	// Format as context for AI
	contextText := searchService.FormatResultsAsContext(results)

	c.JSON(http.StatusOK, gin.H{
		"query":   query,
		"results": results.Results,
		"context": contextText,
		"meta": gin.H{
			"total_results": results.TotalResults,
			"search_time":   results.SearchTime,
			"provider":      results.Provider,
		},
	})
}
