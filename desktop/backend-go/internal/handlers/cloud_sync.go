package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// CloudSyncHandler handles cloud-mode synchronisation endpoints between a local
// BusinessOS desktop client and a MIOSA Firecracker VM instance.
//
// Push and Pull are intentional stubs: they accept / return valid payloads and
// log the request, but perform no database operations yet. The persistence layer
// will be wired in a follow-up sprint.
type CloudSyncHandler struct{}

// NewCloudSyncHandler constructs a CloudSyncHandler.
func NewCloudSyncHandler() *CloudSyncHandler {
	return &CloudSyncHandler{}
}

// SyncPushRequest is the payload sent by a local BusinessOS client.
type SyncPushRequest struct {
	DeviceID  string       `json:"device_id" binding:"required"`
	Timestamp time.Time    `json:"timestamp" binding:"required"`
	Changes   []SyncChange `json:"changes" binding:"required"`
}

// SyncChange describes a single record-level change to be applied.
type SyncChange struct {
	Table     string          `json:"table"` // e.g. "projects", "tasks", "clients"
	RecordID  string          `json:"record_id"`
	Action    string          `json:"action"` // "create", "update", "delete"
	Data      json.RawMessage `json:"data"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// SyncPullResponse is the payload returned to a client requesting incremental changes.
type SyncPullResponse struct {
	Changes         []SyncChange `json:"changes"`
	ServerTimestamp time.Time    `json:"server_timestamp"`
	HasMore         bool         `json:"has_more"`
}

// Push handles POST /api/sync/push.
//
// Receives a batch of changes from a local BusinessOS client, validates the
// payload, logs the event, and returns 202 Accepted. Actual change application
// is a placeholder pending the persistence implementation.
func (h *CloudSyncHandler) Push(c *gin.Context) {
	var req SyncPushRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "invalid request body",
			"detail": err.Error(),
		})
		return
	}

	slog.InfoContext(c.Request.Context(), "cloud_sync: push received",
		"device_id", req.DeviceID,
		"timestamp", req.Timestamp,
		"change_count", len(req.Changes),
	)

	// TODO(sprint-N): persist req.Changes using the sync service.
	c.JSON(http.StatusAccepted, gin.H{
		"status":           "accepted",
		"received_changes": len(req.Changes),
		"server_timestamp": time.Now().UTC(),
	})
}

// Pull handles GET /api/sync/pull.
//
// Returns changes that occurred after the given ?since timestamp for the
// requesting device. Currently returns an empty changeset as a placeholder
// until the persistence layer is implemented.
//
// Query params:
//   - since      RFC3339 timestamp (required)
//   - device_id  caller device identifier (required)
func (h *CloudSyncHandler) Pull(c *gin.Context) {
	since := c.Query("since")
	deviceID := c.Query("device_id")

	if since == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing required query parameter: since"})
		return
	}
	if deviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing required query parameter: device_id"})
		return
	}

	sinceTime, err := time.Parse(time.RFC3339, since)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "invalid since timestamp; expected RFC3339 format",
			"detail": err.Error(),
		})
		return
	}

	slog.InfoContext(c.Request.Context(), "cloud_sync: pull requested",
		"device_id", deviceID,
		"since", sinceTime,
	)

	// TODO(sprint-N): query the sync log for changes after sinceTime for this device.
	resp := SyncPullResponse{
		Changes:         []SyncChange{},
		ServerTimestamp: time.Now().UTC(),
		HasMore:         false,
	}
	c.JSON(http.StatusOK, resp)
}
