package handlers

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/services"
)

// =====================================================================
// HANDLER
// =====================================================================

type BackgroundJobsHandler struct {
	service   *services.BackgroundJobsService
	scheduler *services.JobScheduler
}

func NewBackgroundJobsHandler(pool *pgxpool.Pool) *BackgroundJobsHandler {
	service := services.NewBackgroundJobsService(pool)
	scheduler := services.NewJobScheduler(pool, service)

	return &BackgroundJobsHandler{
		service:   service,
		scheduler: scheduler,
	}
}

// =====================================================================
// BACKGROUND JOBS ENDPOINTS
// =====================================================================

// EnqueueJob creates a new background job
// POST /api/background-jobs
func (h *BackgroundJobsHandler) EnqueueJob(c *gin.Context) {
	var req struct {
		JobType     string                 `json:"job_type" binding:"required"`
		Payload     map[string]interface{} `json:"payload" binding:"required"`
		Priority    *int                   `json:"priority"`
		MaxAttempts *int                   `json:"max_attempts"`
		ScheduledAt *time.Time             `json:"scheduled_at"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		slog.ErrorContext(c.Request.Context(), "Invalid request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	priority := 0
	if req.Priority != nil {
		priority = *req.Priority
	}

	maxAttempts := 3
	if req.MaxAttempts != nil {
		maxAttempts = *req.MaxAttempts
	}

	job, err := h.service.EnqueueJob(
		c.Request.Context(),
		req.JobType,
		req.Payload,
		priority,
		maxAttempts,
		req.ScheduledAt,
	)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "Failed to enqueue job", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to enqueue job"})
		return
	}

	c.JSON(http.StatusCreated, job)
}

// ListJobs retrieves background jobs with filters
// GET /api/background-jobs?status=pending&job_type=email_send&limit=50&offset=0
func (h *BackgroundJobsHandler) ListJobs(c *gin.Context) {
	filters := services.JobListFilters{
		Limit:     50,
		Offset:    0,
		SortBy:    "created_at",
		SortOrder: "DESC",
	}

	if status := c.Query("status"); status != "" {
		filters.Status = &status
	}

	if jobType := c.Query("job_type"); jobType != "" {
		filters.JobType = &jobType
	}

	if limit, err := strconv.Atoi(c.DefaultQuery("limit", "50")); err == nil && limit > 0 {
		filters.Limit = limit
	}

	if offset, err := strconv.Atoi(c.DefaultQuery("offset", "0")); err == nil && offset >= 0 {
		filters.Offset = offset
	}

	if sortBy := c.Query("sort_by"); sortBy != "" {
		filters.SortBy = sortBy
	}

	if sortOrder := c.Query("sort_order"); sortOrder != "" {
		filters.SortOrder = sortOrder
	}

	jobs, err := h.service.ListJobs(c.Request.Context(), filters)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "Failed to list jobs", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list jobs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"jobs":   jobs,
		"limit":  filters.Limit,
		"offset": filters.Offset,
	})
}

// GetJobStatus retrieves a specific job by ID
// GET /api/background-jobs/:id
func (h *BackgroundJobsHandler) GetJobStatus(c *gin.Context) {
	jobIDStr := c.Param("id")
	jobID, err := uuid.Parse(jobIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}

	job, err := h.service.GetJobStatus(c.Request.Context(), jobID)
	if err != nil {
		if err.Error() == "job not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
			return
		}
		slog.ErrorContext(c.Request.Context(), "Failed to get job status", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get job status"})
		return
	}

	c.JSON(http.StatusOK, job)
}

// RetryJob manually retries a failed job
// POST /api/background-jobs/:id/retry
func (h *BackgroundJobsHandler) RetryJob(c *gin.Context) {
	jobIDStr := c.Param("id")
	jobID, err := uuid.Parse(jobIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}

	if err := h.service.RetryJob(c.Request.Context(), jobID); err != nil {
		slog.ErrorContext(c.Request.Context(), "Failed to retry job", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retry job"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Job retried successfully"})
}

// CancelJob cancels a pending or running job
// POST /api/background-jobs/:id/cancel
func (h *BackgroundJobsHandler) CancelJob(c *gin.Context) {
	jobIDStr := c.Param("id")
	jobID, err := uuid.Parse(jobIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}

	if err := h.service.CancelJob(c.Request.Context(), jobID); err != nil {
		slog.ErrorContext(c.Request.Context(), "Failed to cancel job", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel job"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Job cancelled successfully"})
}

// =====================================================================
// SCHEDULED JOBS ENDPOINTS
// =====================================================================

// CreateScheduledJob creates a new recurring job
// POST /api/scheduled-jobs
func (h *BackgroundJobsHandler) CreateScheduledJob(c *gin.Context) {
	var req services.CreateScheduledJobRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		slog.ErrorContext(c.Request.Context(), "Invalid request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	job, err := h.scheduler.CreateScheduledJob(c.Request.Context(), req)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "Failed to create scheduled job", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, job)
}

// ListScheduledJobs retrieves all scheduled jobs
// GET /api/scheduled-jobs?active_only=true
func (h *BackgroundJobsHandler) ListScheduledJobs(c *gin.Context) {
	activeOnly := c.DefaultQuery("active_only", "false") == "true"

	jobs, err := h.scheduler.ListScheduledJobs(c.Request.Context(), activeOnly)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "Failed to list scheduled jobs", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list scheduled jobs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"scheduled_jobs": jobs})
}

// GetScheduledJob retrieves a specific scheduled job
// GET /api/scheduled-jobs/:id
func (h *BackgroundJobsHandler) GetScheduledJob(c *gin.Context) {
	jobIDStr := c.Param("id")
	jobID, err := uuid.Parse(jobIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}

	job, err := h.scheduler.GetScheduledJob(c.Request.Context(), jobID)
	if err != nil {
		if err.Error() == "scheduled job not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Scheduled job not found"})
			return
		}
		slog.ErrorContext(c.Request.Context(), "Failed to get scheduled job", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get scheduled job"})
		return
	}

	c.JSON(http.StatusOK, job)
}

// UpdateScheduledJob updates a scheduled job
// PUT /api/scheduled-jobs/:id
func (h *BackgroundJobsHandler) UpdateScheduledJob(c *gin.Context) {
	jobIDStr := c.Param("id")
	jobID, err := uuid.Parse(jobIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}

	var req services.UpdateScheduledJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.ErrorContext(c.Request.Context(), "Invalid request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	job, err := h.scheduler.UpdateScheduledJob(c.Request.Context(), jobID, req)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "Failed to update scheduled job", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update scheduled job"})
		return
	}

	c.JSON(http.StatusOK, job)
}

// DeleteScheduledJob deletes a scheduled job
// DELETE /api/scheduled-jobs/:id
func (h *BackgroundJobsHandler) DeleteScheduledJob(c *gin.Context) {
	jobIDStr := c.Param("id")
	jobID, err := uuid.Parse(jobIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}

	if err := h.scheduler.DeleteScheduledJob(c.Request.Context(), jobID); err != nil {
		slog.ErrorContext(c.Request.Context(), "Failed to delete scheduled job", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete scheduled job"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Scheduled job deleted successfully"})
}

// EnableScheduledJob enables a scheduled job
// POST /api/scheduled-jobs/:id/enable
func (h *BackgroundJobsHandler) EnableScheduledJob(c *gin.Context) {
	jobIDStr := c.Param("id")
	jobID, err := uuid.Parse(jobIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}

	if err := h.scheduler.EnableScheduledJob(c.Request.Context(), jobID); err != nil {
		slog.ErrorContext(c.Request.Context(), "Failed to enable scheduled job", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to enable scheduled job"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Scheduled job enabled successfully"})
}

// DisableScheduledJob disables a scheduled job
// POST /api/scheduled-jobs/:id/disable
func (h *BackgroundJobsHandler) DisableScheduledJob(c *gin.Context) {
	jobIDStr := c.Param("id")
	jobID, err := uuid.Parse(jobIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}

	if err := h.scheduler.DisableScheduledJob(c.Request.Context(), jobID); err != nil {
		slog.ErrorContext(c.Request.Context(), "Failed to disable scheduled job", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to disable scheduled job"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Scheduled job disabled successfully"})
}

// =====================================================================
// ROUTE REGISTRATION
// =====================================================================

// RegisterRoutes registers all background jobs routes
func (h *BackgroundJobsHandler) RegisterRoutes(router *gin.RouterGroup) {
	// Background jobs routes
	jobs := router.Group("/background-jobs")
	{
		jobs.POST("", h.EnqueueJob)                   // Create job
		jobs.GET("", h.ListJobs)                      // List jobs
		jobs.GET("/:id", h.GetJobStatus)              // Get job status
		jobs.POST("/:id/retry", h.RetryJob)           // Retry job
		jobs.POST("/:id/cancel", h.CancelJob)         // Cancel job
	}

	// Scheduled jobs routes
	scheduled := router.Group("/scheduled-jobs")
	{
		scheduled.POST("", h.CreateScheduledJob)           // Create scheduled job
		scheduled.GET("", h.ListScheduledJobs)             // List scheduled jobs
		scheduled.GET("/:id", h.GetScheduledJob)           // Get scheduled job
		scheduled.PUT("/:id", h.UpdateScheduledJob)        // Update scheduled job
		scheduled.DELETE("/:id", h.DeleteScheduledJob)     // Delete scheduled job
		scheduled.POST("/:id/enable", h.EnableScheduledJob) // Enable scheduled job
		scheduled.POST("/:id/disable", h.DisableScheduledJob) // Disable scheduled job
	}

	slog.Info("Background jobs routes registered")
}

// GetScheduler returns the scheduler instance (for starting in main.go)
func (h *BackgroundJobsHandler) GetScheduler() *services.JobScheduler {
	return h.scheduler
}

// GetService returns the service instance (for creating workers in main.go)
func (h *BackgroundJobsHandler) GetService() *services.BackgroundJobsService {
	return h.service
}
