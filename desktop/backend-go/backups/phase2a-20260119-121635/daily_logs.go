package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rhl/businessos-backend/internal/database/sqlc"
	"github.com/rhl/businessos-backend/internal/middleware"
)

// ListDailyLogs returns daily logs for the current user
func (h *Handlers) ListDailyLogs(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	queries := sqlc.New(h.pool)

	// Parse pagination params
	limit := int32(30)
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = int32(parsed)
		}
	}

	offset := int32(0)
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = int32(parsed)
		}
	}

	logs, err := queries.ListDailyLogs(c.Request.Context(), sqlc.ListDailyLogsParams{
		UserID: user.ID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list daily logs"})
		return
	}

	c.JSON(http.StatusOK, logs)
}

// GetTodayLog returns today's daily log
func (h *Handlers) GetTodayLog(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	queries := sqlc.New(h.pool)
	log, err := queries.GetTodayLog(c.Request.Context(), user.ID)
	if err != nil {
		// Return null instead of 404 for missing today log
		c.JSON(http.StatusOK, nil)
		return
	}

	c.JSON(http.StatusOK, transformDailyLog(log))
}

// transformDailyLog converts sqlc DailyLog to JSON-friendly format
func transformDailyLog(log sqlc.DailyLog) map[string]interface{} {
	result := map[string]interface{}{
		"id":                   uuidToString(log.ID),
		"user_id":              log.UserID,
		"date":                 log.Date.Time.Format("2006-01-02"),
		"content":              log.Content,
		"transcription_source": log.TranscriptionSource,
		"energy_level":         log.EnergyLevel,
	}

	// Handle JSONB fields
	if log.ExtractedActions != nil {
		var actions interface{}
		if err := json.Unmarshal(log.ExtractedActions, &actions); err == nil {
			result["extracted_actions"] = actions
		} else {
			result["extracted_actions"] = nil
		}
	} else {
		result["extracted_actions"] = nil
	}

	if log.ExtractedPatterns != nil {
		var patterns interface{}
		if err := json.Unmarshal(log.ExtractedPatterns, &patterns); err == nil {
			result["extracted_patterns"] = patterns
		} else {
			result["extracted_patterns"] = nil
		}
	} else {
		result["extracted_patterns"] = nil
	}

	if log.CreatedAt.Valid {
		result["created_at"] = log.CreatedAt.Time.Format(time.RFC3339)
	} else {
		result["created_at"] = nil
	}

	if log.UpdatedAt.Valid {
		result["updated_at"] = log.UpdatedAt.Time.Format(time.RFC3339)
	} else {
		result["updated_at"] = nil
	}

	return result
}

// GetDailyLogByDate returns a daily log for a specific date
func (h *Handlers) GetDailyLogByDate(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	dateStr := c.Param("date")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Use YYYY-MM-DD"})
		return
	}

	queries := sqlc.New(h.pool)
	log, err := queries.GetDailyLogByDate(c.Request.Context(), sqlc.GetDailyLogByDateParams{
		UserID: user.ID,
		Date:   pgtype.Date{Time: date, Valid: true},
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No log found for this date"})
		return
	}

	c.JSON(http.StatusOK, log)
}

// CreateDailyLog creates a new daily log
func (h *Handlers) CreateDailyLog(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var req struct {
		Date                string   `json:"date"`
		Content             string   `json:"content" binding:"required"`
		TranscriptionSource *string  `json:"transcription_source"`
		ExtractedActions    []string `json:"extracted_actions"`
		ExtractedPatterns   []string `json:"extracted_patterns"`
		EnergyLevel         *int32   `json:"energy_level"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	queries := sqlc.New(h.pool)

	// Parse date, default to today
	logDate := time.Now()
	if req.Date != "" {
		if t, err := time.Parse("2006-01-02", req.Date); err == nil {
			logDate = t
		}
	}

	// Handle JSONB arrays
	extractedActions := []byte("[]")
	if req.ExtractedActions != nil && len(req.ExtractedActions) > 0 {
		if actionsJSON, err := json.Marshal(req.ExtractedActions); err == nil {
			extractedActions = actionsJSON
		}
	}

	extractedPatterns := []byte("[]")
	if req.ExtractedPatterns != nil && len(req.ExtractedPatterns) > 0 {
		if patternsJSON, err := json.Marshal(req.ExtractedPatterns); err == nil {
			extractedPatterns = patternsJSON
		}
	}

	log, err := queries.CreateDailyLog(c.Request.Context(), sqlc.CreateDailyLogParams{
		UserID:              user.ID,
		Date:                pgtype.Date{Time: logDate, Valid: true},
		Content:             req.Content,
		TranscriptionSource: req.TranscriptionSource,
		ExtractedActions:    extractedActions,
		ExtractedPatterns:   extractedPatterns,
		EnergyLevel:         req.EnergyLevel,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create daily log"})
		return
	}

	c.JSON(http.StatusCreated, log)
}

// CreateOrUpdateDailyLog creates or updates a daily log for a date
func (h *Handlers) CreateOrUpdateDailyLog(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var req struct {
		Date                string   `json:"date"`
		Content             string   `json:"content" binding:"required"`
		TranscriptionSource *string  `json:"transcription_source"`
		ExtractedActions    []string `json:"extracted_actions"`
		ExtractedPatterns   []string `json:"extracted_patterns"`
		EnergyLevel         *int32   `json:"energy_level"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	queries := sqlc.New(h.pool)

	// Parse date, default to today
	logDate := time.Now()
	if req.Date != "" {
		if t, err := time.Parse("2006-01-02", req.Date); err == nil {
			logDate = t
		}
	}

	// Handle JSONB arrays
	extractedActions := []byte("[]")
	if req.ExtractedActions != nil && len(req.ExtractedActions) > 0 {
		if actionsJSON, err := json.Marshal(req.ExtractedActions); err == nil {
			extractedActions = actionsJSON
		}
	}

	extractedPatterns := []byte("[]")
	if req.ExtractedPatterns != nil && len(req.ExtractedPatterns) > 0 {
		if patternsJSON, err := json.Marshal(req.ExtractedPatterns); err == nil {
			extractedPatterns = patternsJSON
		}
	}

	log, err := queries.UpsertDailyLog(c.Request.Context(), sqlc.UpsertDailyLogParams{
		UserID:              user.ID,
		Date:                pgtype.Date{Time: logDate, Valid: true},
		Content:             req.Content,
		TranscriptionSource: req.TranscriptionSource,
		ExtractedActions:    extractedActions,
		ExtractedPatterns:   extractedPatterns,
		EnergyLevel:         req.EnergyLevel,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upsert daily log"})
		return
	}

	c.JSON(http.StatusOK, log)
}

// UpdateDailyLog updates an existing daily log
func (h *Handlers) UpdateDailyLog(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid daily log ID"})
		return
	}

	var req struct {
		Content             string   `json:"content" binding:"required"`
		TranscriptionSource *string  `json:"transcription_source"`
		ExtractedActions    []string `json:"extracted_actions"`
		ExtractedPatterns   []string `json:"extracted_patterns"`
		EnergyLevel         *int32   `json:"energy_level"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	queries := sqlc.New(h.pool)

	// Handle JSONB arrays
	extractedActions := []byte("[]")
	if req.ExtractedActions != nil && len(req.ExtractedActions) > 0 {
		if actionsJSON, err := json.Marshal(req.ExtractedActions); err == nil {
			extractedActions = actionsJSON
		}
	}

	extractedPatterns := []byte("[]")
	if req.ExtractedPatterns != nil && len(req.ExtractedPatterns) > 0 {
		if patternsJSON, err := json.Marshal(req.ExtractedPatterns); err == nil {
			extractedPatterns = patternsJSON
		}
	}

	log, err := queries.UpdateDailyLog(c.Request.Context(), sqlc.UpdateDailyLogParams{
		ID:                  pgtype.UUID{Bytes: id, Valid: true},
		Content:             req.Content,
		TranscriptionSource: req.TranscriptionSource,
		ExtractedActions:    extractedActions,
		ExtractedPatterns:   extractedPatterns,
		EnergyLevel:         req.EnergyLevel,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update daily log"})
		return
	}

	c.JSON(http.StatusOK, log)
}

// AppendToDailyLog appends content to today's daily log (or creates it)
func (h *Handlers) AppendToDailyLog(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var req struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	queries := sqlc.New(h.pool)

	// Try to get today's log
	existingLog, err := queries.GetTodayLog(c.Request.Context(), user.ID)
	if err != nil {
		// Create new log for today
		log, err := queries.CreateDailyLog(c.Request.Context(), sqlc.CreateDailyLogParams{
			UserID:            user.ID,
			Date:              pgtype.Date{Time: time.Now(), Valid: true},
			Content:           req.Content,
			ExtractedActions:  []byte("[]"),
			ExtractedPatterns: []byte("[]"),
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create daily log"})
			return
		}
		c.JSON(http.StatusCreated, log)
		return
	}

	// Append to existing log
	newContent := existingLog.Content + "\n\n" + req.Content
	log, err := queries.UpdateDailyLog(c.Request.Context(), sqlc.UpdateDailyLogParams{
		ID:                  existingLog.ID,
		Content:             newContent,
		TranscriptionSource: existingLog.TranscriptionSource,
		ExtractedActions:    existingLog.ExtractedActions,
		ExtractedPatterns:   existingLog.ExtractedPatterns,
		EnergyLevel:         existingLog.EnergyLevel,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to append to daily log"})
		return
	}

	c.JSON(http.StatusOK, log)
}

// DeleteDailyLog deletes a daily log
func (h *Handlers) DeleteDailyLog(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid daily log ID"})
		return
	}

	queries := sqlc.New(h.pool)
	err = queries.DeleteDailyLog(c.Request.Context(), sqlc.DeleteDailyLogParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete daily log"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Daily log deleted"})
}
