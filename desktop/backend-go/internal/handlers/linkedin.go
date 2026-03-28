package handlers

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/integrations/linkedin"
	"github.com/rhl/businessos-backend/internal/utils"
)

// LinkedInHandler manages LinkedIn integration endpoints.
type LinkedInHandler struct {
	logger        *slog.Logger
	repo          *linkedin.Repository
	scorer        *linkedin.ICPScorer
	importer      *linkedin.CSVImporter
	outreachQueue *linkedin.OutreachQueueManager
}

// NewLinkedInHandler creates a new LinkedIn handler.
func NewLinkedInHandler(
	logger *slog.Logger,
	repo *linkedin.Repository,
	scorer *linkedin.ICPScorer,
	importer *linkedin.CSVImporter,
	outreachQueue *linkedin.OutreachQueueManager,
) *LinkedInHandler {
	return &LinkedInHandler{
		logger:        logger,
		repo:          repo,
		scorer:        scorer,
		importer:      importer,
		outreachQueue: outreachQueue,
	}
}

// ImportCSV handles POST /api/linkedin/import
// Accepts CSV content and imports/updates contacts.
func (h *LinkedInHandler) ImportCSV(c *gin.Context) {
	var req linkedin.ImportCSVRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, h.logger, err)
		return
	}

	if req.CSVContent == "" {
		utils.RespondBadRequest(c, h.logger, "csv_content is required")
		return
	}

	// Parse and import CSV
	imported, updated, failed, errors := h.importer.ImportCSV(req.CSVContent)

	h.logger.Info("LinkedIn CSV imported",
		"imported", imported,
		"updated", updated,
		"failed", failed,
	)

	c.JSON(http.StatusOK, linkedin.ImportCSVResponse{
		ContactsImported: imported,
		ContactsUpdated:  updated,
		ContactsFailed:   failed,
		Errors:           errors,
	})
}

// GetContacts handles GET /api/linkedin/contacts
// Returns paginated list of all contacts.
func (h *LinkedInHandler) GetContacts(c *gin.Context) {
	page := 1
	if p := c.Query("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}

	pageSize := 50
	if ps := c.Query("page_size"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 && v <= 200 {
			pageSize = v
		}
	}

	offset := (page - 1) * pageSize

	// Get qualified contacts with pagination (minScore=0.0 returns all contacts)
	contacts, err := h.repo.GetQualifiedContactsPaginated(0.0, pageSize, offset)
	if err != nil {
		h.logger.Error("Failed to fetch contacts", "error", err)
		utils.RespondInternalError(c, h.logger, "fetch LinkedIn contacts", err)
		return
	}

	if contacts == nil {
		contacts = []*linkedin.Contact{}
	}

	// Get total count for pagination metadata
	total, err := h.repo.CountQualifiedContacts(0.0)
	if err != nil {
		h.logger.Error("Failed to count contacts", "error", err)
		total = int64(len(contacts))
	}

	// Convert pointers to values for response
	contactValues := make([]linkedin.Contact, len(contacts))
	for i, c := range contacts {
		if c != nil {
			contactValues[i] = *c
		}
	}

	c.JSON(http.StatusOK, linkedin.ContactListResponse{
		Contacts: contactValues,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		HasMore:  int64(offset+pageSize) < total,
	})
}

// ICPScoreContacts handles POST /api/linkedin/icp-score
// Scores all contacts and returns qualified count (score >= min_score).
func (h *LinkedInHandler) ICPScoreContacts(c *gin.Context) {
	minScore := 0.7
	if ms := c.Query("min_score"); ms != "" {
		if score, err := parseFloatParam(ms); err == nil {
			minScore = score
		}
	}

	// Get all unscored contacts
	contacts, err := h.repo.GetQualifiedContacts(0.0, 10000)
	if err != nil {
		h.logger.Error("Failed to fetch contacts", "error", err)
		utils.RespondInternalError(c, h.logger, "fetch qualified contacts for ICP scoring", err)
		return
	}

	var qualified int
	for _, contact := range contacts {
		if contact.ICPScore == 0 { // Only score if not already scored
			score := h.scorer.ScoreContact(contact)
			contact.ICPScore = score

			// Update contact
			if err := h.repo.UpdateContact(contact); err != nil {
				h.logger.Error("Failed to update contact score", "id", contact.ID, "error", err)
				continue
			}

			if score >= minScore {
				qualified++
			}
		}
	}

	h.logger.Info("LinkedIn contacts scored",
		"total_contacts", len(contacts),
		"qualified", qualified,
		"min_score", minScore,
	)

	c.JSON(http.StatusOK, gin.H{
		"qualified":      qualified,
		"total_contacts": len(contacts),
		"min_score":      minScore,
	})
}

// EnrollOutreach handles POST /api/linkedin/outreach/enroll
// Enrolls qualified contacts into an outreach sequence.
func (h *LinkedInHandler) EnrollOutreach(c *gin.Context) {
	var req linkedin.EnrollOutreachRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, h.logger, err)
		return
	}

	if req.SequenceID <= 0 {
		utils.RespondBadRequest(c, h.logger, "sequence_id is required")
		return
	}

	if req.MinScore == 0 {
		req.MinScore = 0.7 // Default
	}

	// Get qualified contacts
	contacts, err := h.repo.GetQualifiedContacts(req.MinScore, 1000)
	if err != nil {
		h.logger.Error("Failed to fetch qualified contacts", "error", err)
		utils.RespondInternalError(c, h.logger, "enroll contact in outreach sequence", err)
		return
	}

	// Extract contact IDs
	var contactIDs []int64
	for _, contact := range contacts {
		contactIDs = append(contactIDs, contact.ID)
	}

	// Limit target count if specified
	if req.TargetCount > 0 && len(contactIDs) > req.TargetCount {
		contactIDs = contactIDs[:req.TargetCount]
	}

	// Enqueue messages for outreach
	queued, errors := h.outreachQueue.EnqueueBatch(c.Request.Context(), contactIDs, req.SequenceID)

	h.logger.Info("Outreach enrollment completed",
		"sequence_id", req.SequenceID,
		"queued", queued,
		"failed", len(errors),
	)

	c.JSON(http.StatusOK, linkedin.EnrollOutreachResponse{
		Enrolled: queued,
		Skipped:  len(contactIDs) - queued,
		Errors:   errors,
	})
}

// Helper functions

func parseFloatParam(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}
