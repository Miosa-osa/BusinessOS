package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/rhl/businessos-backend/internal/database/sqlc"
	"github.com/rhl/businessos-backend/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

// CalendarStatsResponse is the API response for calendar stats
type CalendarStatsResponse struct {
	TotalEvents  int64           `json:"total_events"`
	GoogleEvents int64           `json:"google_events"`
	LocalEvents  int64           `json:"local_events"`
	DateRange    *DateRangeStats `json:"date_range,omitempty"`
	LastSync     *time.Time      `json:"last_sync,omitempty"`
}

type DateRangeStats struct {
	From *time.Time `json:"from,omitempty"`
	To   *time.Time `json:"to,omitempty"`
}

// GetCalendarStats returns calendar statistics for the user
func (h *Handlers) GetCalendarStats(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	queries := sqlc.New(h.pool)
	stats, err := queries.GetCalendarStats(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get calendar stats"})
		return
	}

	response := CalendarStatsResponse{
		TotalEvents:  stats.TotalEvents,
		GoogleEvents: stats.GoogleEvents,
		LocalEvents:  stats.LocalEvents,
	}

	// Handle date range
	if stats.DateRangeFrom != nil || stats.DateRangeTo != nil {
		response.DateRange = &DateRangeStats{}
		if stats.DateRangeFrom != nil {
			if t, ok := stats.DateRangeFrom.(time.Time); ok {
				response.DateRange.From = &t
			} else if ts, ok := stats.DateRangeFrom.(pgtype.Timestamptz); ok && ts.Valid {
				response.DateRange.From = &ts.Time
			}
		}
		if stats.DateRangeTo != nil {
			if t, ok := stats.DateRangeTo.(time.Time); ok {
				response.DateRange.To = &t
			} else if ts, ok := stats.DateRangeTo.(pgtype.Timestamptz); ok && ts.Valid {
				response.DateRange.To = &ts.Time
			}
		}
	}

	// Handle last sync
	if stats.LastSync != nil {
		if t, ok := stats.LastSync.(time.Time); ok {
			response.LastSync = &t
		} else if ts, ok := stats.LastSync.(pgtype.Timestamptz); ok && ts.Valid {
			response.LastSync = &ts.Time
		}
	}

	c.JSON(http.StatusOK, response)
}

// GetUpcomingCalendarEvents returns upcoming calendar events
func (h *Handlers) GetUpcomingCalendarEvents(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	// Parse limit from query params, default to 10
	limit := int32(10)
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = int32(l)
		}
	}

	queries := sqlc.New(h.pool)
	events, err := queries.GetUpcomingEvents(c.Request.Context(), sqlc.GetUpcomingEventsParams{
		UserID: user.ID,
		Limit:  limit,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get upcoming events"})
		return
	}

	response := make([]CalendarEventResponse, len(events))
	for i, e := range events {
		response[i] = TransformCalendarEvent(e)
	}

	c.JSON(http.StatusOK, response)
}

// GetTodayCalendarEvents returns today's calendar events
func (h *Handlers) GetTodayCalendarEvents(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	queries := sqlc.New(h.pool)
	events, err := queries.GetTodayEvents(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get today's events"})
		return
	}

	response := make([]CalendarEventResponse, len(events))
	for i, e := range events {
		response[i] = TransformCalendarEvent(e)
	}

	c.JSON(http.StatusOK, response)
}
