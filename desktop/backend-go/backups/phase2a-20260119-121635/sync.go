package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/middleware"
)

// SyncRequest represents a sync request
type SyncRequest struct {
	Since      string   `form:"since"`
	Tables     []string `form:"tables"`
	FullSync   bool     `form:"full_sync"`
}

// SyncableTable defines tables that support sync
var SyncableTables = map[string]string{
	"contexts":            "SELECT * FROM contexts WHERE user_id = $1 AND updated_at > $2 ORDER BY updated_at DESC",
	"conversations":       "SELECT * FROM conversations WHERE user_id = $1 AND updated_at > $2 ORDER BY updated_at DESC",
	"messages":            "SELECT m.* FROM messages m JOIN conversations c ON m.conversation_id = c.id WHERE c.user_id = $1 AND m.created_at > $2 ORDER BY m.created_at DESC",
	"projects":            "SELECT * FROM projects WHERE user_id = $1 AND updated_at > $2 ORDER BY updated_at DESC",
	"artifacts":           "SELECT * FROM artifacts WHERE user_id = $1 AND updated_at > $2 ORDER BY updated_at DESC",
	"nodes":               "SELECT * FROM nodes WHERE user_id = $1 AND updated_at > $2 ORDER BY updated_at DESC",
	"team_members":        "SELECT * FROM team_members WHERE user_id = $1 AND updated_at > $2 ORDER BY updated_at DESC",
	"tasks":               "SELECT * FROM tasks WHERE user_id = $1 AND updated_at > $2 ORDER BY updated_at DESC",
	"focus_items":         "SELECT * FROM focus_items WHERE user_id = $1 AND updated_at > $2 ORDER BY updated_at DESC",
	"daily_logs":          "SELECT * FROM daily_logs WHERE user_id = $1 AND updated_at > $2 ORDER BY updated_at DESC",
	"user_settings":       "SELECT * FROM user_settings WHERE user_id = $1 AND updated_at > $2",
	"clients":             "SELECT * FROM clients WHERE user_id = $1 AND updated_at > $2 ORDER BY updated_at DESC",
	"client_contacts":     "SELECT cc.* FROM client_contacts cc JOIN clients c ON cc.client_id = c.id WHERE c.user_id = $1 AND cc.updated_at > $2 ORDER BY cc.updated_at DESC",
	"client_interactions": "SELECT ci.* FROM client_interactions ci JOIN clients c ON ci.client_id = c.id WHERE c.user_id = $1 AND ci.created_at > $2 ORDER BY ci.created_at DESC",
	"client_deals":        "SELECT cd.* FROM client_deals cd JOIN clients c ON cd.client_id = c.id WHERE c.user_id = $1 AND cd.updated_at > $2 ORDER BY cd.updated_at DESC",
	"calendar_events":     "SELECT * FROM calendar_events WHERE user_id = $1 AND updated_at > $2 ORDER BY updated_at DESC",
}

// GetSyncChanges returns changes since a given timestamp
// GET /api/sync/:table
func (h *Handlers) GetSyncChanges(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := user.ID

	table := c.Param("table")
	since := c.Query("since")

	// Validate table is syncable
	query, ok := SyncableTables[table]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Table not syncable: " + table})
		return
	}

	// Parse since timestamp, default to epoch
	sinceTime := time.Time{}
	if since != "" {
		var err error
		sinceTime, err = time.Parse(time.RFC3339, since)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid since timestamp"})
			return
		}
	}

	// Execute query
	rows, err := h.pool.Query(c, query, userID, sinceTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query changes"})
		return
	}
	defer rows.Close()

	// Collect results
	var results []map[string]interface{}
	fieldDescriptions := rows.FieldDescriptions()

	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			continue
		}

		row := make(map[string]interface{})
		for i, fd := range fieldDescriptions {
			row[string(fd.Name)] = values[i]
		}
		results = append(results, row)
	}

	if results == nil {
		results = []map[string]interface{}{}
	}

	c.JSON(http.StatusOK, results)
}

// FullSync returns all data for initial sync
// GET /api/sync/full
func (h *Handlers) FullSync(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := user.ID

	epoch := time.Time{}
	result := make(map[string][]map[string]interface{})

	for table, query := range SyncableTables {
		rows, err := h.pool.Query(c, query, userID, epoch)
		if err != nil {
			continue
		}

		var records []map[string]interface{}
		fieldDescriptions := rows.FieldDescriptions()

		for rows.Next() {
			values, err := rows.Values()
			if err != nil {
				continue
			}

			row := make(map[string]interface{})
			for i, fd := range fieldDescriptions {
				row[string(fd.Name)] = values[i]
			}
			records = append(records, row)
		}
		rows.Close()

		if records == nil {
			records = []map[string]interface{}{}
		}
		result[table] = records
	}

	c.JSON(http.StatusOK, gin.H{
		"sync_timestamp": time.Now().UTC().Format(time.RFC3339),
		"data":           result,
	})
}

// GetSyncStatus returns sync status and server timestamp
// GET /api/sync/status
func (h *Handlers) GetSyncStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":           "healthy",
		"server_timestamp": time.Now().UTC().Format(time.RFC3339),
		"version":          "1.0.0",
	})
}

// createTableSyncHandler creates a handler for a specific table's sync endpoint
func (h *Handlers) createTableSyncHandler(table string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := middleware.GetCurrentUser(c)
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		userID := user.ID

		since := c.Query("since")

		// Validate table is syncable
		query, ok := SyncableTables[table]
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Table not syncable: " + table})
			return
		}

		// Parse since timestamp, default to epoch
		sinceTime := time.Time{}
		if since != "" {
			var err error
			sinceTime, err = time.Parse(time.RFC3339, since)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid since timestamp"})
				return
			}
		}

		// Execute query
		rows, err := h.pool.Query(c, query, userID, sinceTime)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query changes"})
			return
		}
		defer rows.Close()

		// Collect results
		var results []map[string]interface{}
		fieldDescriptions := rows.FieldDescriptions()

		for rows.Next() {
			values, err := rows.Values()
			if err != nil {
				continue
			}

			row := make(map[string]interface{})
			for i, fd := range fieldDescriptions {
				row[string(fd.Name)] = values[i]
			}
			results = append(results, row)
		}

		if results == nil {
			results = []map[string]interface{}{}
		}

		c.JSON(http.StatusOK, results)
	}
}
