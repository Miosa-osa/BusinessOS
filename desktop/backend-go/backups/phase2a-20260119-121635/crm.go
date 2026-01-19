package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rhl/businessos-backend/internal/database/sqlc"
	"github.com/rhl/businessos-backend/internal/middleware"
)

// ============================================================================
// COMPANIES HANDLERS
// ============================================================================

// ListCompanies returns all companies for the current user
func (h *Handlers) ListCompanies(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	queries := sqlc.New(h.pool)

	// Parse query params
	industry := c.Query("industry")
	lifecycleStage := c.Query("lifecycle_stage")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	companies, err := queries.ListCompanies(c.Request.Context(), sqlc.ListCompaniesParams{
		UserID:         user.ID,
		Industry:       crmToNullString(industry),
		LifecycleStage: crmToNullString(lifecycleStage),
		LimitVal:       int32(limit),
		OffsetVal:      int32(offset),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list companies"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"companies": transformCompanies(companies),
		"count":     len(companies),
	})
}

// GetCompany returns a single company by ID
func (h *Handlers) GetCompany(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}

	queries := sqlc.New(h.pool)
	company, err := queries.GetCompany(c.Request.Context(), sqlc.GetCompanyParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Company not found"})
		return
	}

	c.JSON(http.StatusOK, transformCompany(company))
}

// CreateCompanyRequest represents the request to create a company
type CreateCompanyRequest struct {
	Name           string                 `json:"name" binding:"required"`
	LegalName      *string                `json:"legal_name"`
	Industry       *string                `json:"industry"`
	CompanySize    *string                `json:"company_size"`
	Website        *string                `json:"website"`
	Email          *string                `json:"email"`
	Phone          *string                `json:"phone"`
	AddressLine1   *string                `json:"address_line1"`
	AddressLine2   *string                `json:"address_line2"`
	City           *string                `json:"city"`
	State          *string                `json:"state"`
	PostalCode     *string                `json:"postal_code"`
	Country        *string                `json:"country"`
	AnnualRevenue  *float64               `json:"annual_revenue"`
	Currency       *string                `json:"currency"`
	TaxID          *string                `json:"tax_id"`
	LinkedinURL    *string                `json:"linkedin_url"`
	TwitterHandle  *string                `json:"twitter_handle"`
	OwnerID        *string                `json:"owner_id"`
	LifecycleStage *string                `json:"lifecycle_stage"`
	LeadSource     *string                `json:"lead_source"`
	LogoURL        *string                `json:"logo_url"`
	CustomFields   map[string]interface{} `json:"custom_fields"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// CreateCompany creates a new company
func (h *Handlers) CreateCompany(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var req CreateCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	queries := sqlc.New(h.pool)

	// Convert custom fields and metadata to JSON
	customFields, _ := json.Marshal(req.CustomFields)
	metadata, _ := json.Marshal(req.Metadata)

	company, err := queries.CreateCompany(c.Request.Context(), sqlc.CreateCompanyParams{
		UserID:         user.ID,
		Name:           req.Name,
		LegalName:      req.LegalName,
		Industry:       req.Industry,
		CompanySize:    req.CompanySize,
		Website:        req.Website,
		Email:          req.Email,
		Phone:          req.Phone,
		AddressLine1:   req.AddressLine1,
		AddressLine2:   req.AddressLine2,
		City:           req.City,
		State:          req.State,
		PostalCode:     req.PostalCode,
		Country:        req.Country,
		AnnualRevenue:  crmToNumeric(req.AnnualRevenue),
		Currency:       req.Currency,
		TaxID:          req.TaxID,
		LinkedinUrl:    req.LinkedinURL,
		TwitterHandle:  req.TwitterHandle,
		OwnerID:        req.OwnerID,
		LifecycleStage: req.LifecycleStage,
		LeadSource:     req.LeadSource,
		LogoUrl:        req.LogoURL,
		CustomFields:   customFields,
		Metadata:       metadata,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create company: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, transformCompany(company))
}

// UpdateCompanyRequest represents the request to update a company
type UpdateCompanyRequest struct {
	Name           string                 `json:"name" binding:"required"`
	LegalName      *string                `json:"legal_name"`
	Industry       *string                `json:"industry"`
	CompanySize    *string                `json:"company_size"`
	Website        *string                `json:"website"`
	Email          *string                `json:"email"`
	Phone          *string                `json:"phone"`
	AddressLine1   *string                `json:"address_line1"`
	AddressLine2   *string                `json:"address_line2"`
	City           *string                `json:"city"`
	State          *string                `json:"state"`
	PostalCode     *string                `json:"postal_code"`
	Country        *string                `json:"country"`
	AnnualRevenue  *float64               `json:"annual_revenue"`
	LifecycleStage *string                `json:"lifecycle_stage"`
	LinkedinURL    *string                `json:"linkedin_url"`
	TwitterHandle  *string                `json:"twitter_handle"`
	LogoURL        *string                `json:"logo_url"`
	CustomFields   map[string]interface{} `json:"custom_fields"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// UpdateCompany updates an existing company
func (h *Handlers) UpdateCompany(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}

	var req UpdateCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	queries := sqlc.New(h.pool)

	// Convert custom fields and metadata to JSON
	customFields, _ := json.Marshal(req.CustomFields)
	metadata, _ := json.Marshal(req.Metadata)

	company, err := queries.UpdateCompany(c.Request.Context(), sqlc.UpdateCompanyParams{
		ID:             pgtype.UUID{Bytes: id, Valid: true},
		Name:           req.Name,
		LegalName:      req.LegalName,
		Industry:       req.Industry,
		CompanySize:    req.CompanySize,
		Website:        req.Website,
		Email:          req.Email,
		Phone:          req.Phone,
		AddressLine1:   req.AddressLine1,
		AddressLine2:   req.AddressLine2,
		City:           req.City,
		State:          req.State,
		PostalCode:     req.PostalCode,
		Country:        req.Country,
		AnnualRevenue:  crmToNumeric(req.AnnualRevenue),
		LifecycleStage: req.LifecycleStage,
		LinkedinUrl:    req.LinkedinURL,
		TwitterHandle:  req.TwitterHandle,
		LogoUrl:        req.LogoURL,
		CustomFields:   customFields,
		Metadata:       metadata,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update company"})
		return
	}

	c.JSON(http.StatusOK, transformCompany(company))
}

// DeleteCompany deletes a company
func (h *Handlers) DeleteCompany(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}

	queries := sqlc.New(h.pool)
	err = queries.DeleteCompany(c.Request.Context(), sqlc.DeleteCompanyParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete company"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Company deleted"})
}

// SearchCompanies searches companies by name or website
func (h *Handlers) SearchCompanies(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search query is required"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	queries := sqlc.New(h.pool)
	companies, err := queries.SearchCompanies(c.Request.Context(), sqlc.SearchCompaniesParams{
		UserID:   user.ID,
		Column2:  &query,
		LimitVal: int32(limit),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search companies"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"companies": transformCompanies(companies),
		"count":     len(companies),
	})
}

// ============================================================================
// CONTACT-COMPANY RELATIONS HANDLERS
// ============================================================================

// ListCompanyContacts returns contacts associated with a company
func (h *Handlers) ListCompanyContacts(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	companyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}

	queries := sqlc.New(h.pool)
	contacts, err := queries.ListCompanyContacts(c.Request.Context(), pgtype.UUID{Bytes: companyID, Valid: true})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list contacts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"contacts": contacts,
		"count":    len(contacts),
	})
}

// LinkContactToCompanyRequest represents the request to link a contact to a company
type LinkContactToCompanyRequest struct {
	ContactID  string  `json:"contact_id" binding:"required"`
	JobTitle   *string `json:"job_title"`
	Department *string `json:"department"`
	RoleType   *string `json:"role_type"`
	IsPrimary  bool    `json:"is_primary"`
}

// LinkContactToCompany links a contact to a company
func (h *Handlers) LinkContactToCompany(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	companyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}

	var req LinkContactToCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	contactID, err := uuid.Parse(req.ContactID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid contact ID"})
		return
	}

	queries := sqlc.New(h.pool)
	relation, err := queries.CreateContactCompanyRelation(c.Request.Context(), sqlc.CreateContactCompanyRelationParams{
		ContactID:  pgtype.UUID{Bytes: contactID, Valid: true},
		CompanyID:  pgtype.UUID{Bytes: companyID, Valid: true},
		JobTitle:   req.JobTitle,
		Department: req.Department,
		RoleType:   req.RoleType,
		IsPrimary:  &req.IsPrimary,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to link contact to company"})
		return
	}

	c.JSON(http.StatusCreated, relation)
}

// UnlinkContactFromCompany removes a contact-company relation
func (h *Handlers) UnlinkContactFromCompany(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	relationID, err := uuid.Parse(c.Param("relationId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid relation ID"})
		return
	}

	queries := sqlc.New(h.pool)
	err = queries.DeleteContactCompanyRelation(c.Request.Context(), pgtype.UUID{Bytes: relationID, Valid: true})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unlink contact"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Contact unlinked from company"})
}

// ============================================================================
// PIPELINES HANDLERS
// ============================================================================

// ListPipelines returns all pipelines for the current user
func (h *Handlers) ListPipelines(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	queries := sqlc.New(h.pool)
	pipelines, err := queries.ListPipelines(c.Request.Context(), user.ID)
	if err != nil {
		slog.Error("Failed to list pipelines", "error", err, "user_id", user.ID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list pipelines"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pipelines": transformPipelines(pipelines),
		"count":     len(pipelines),
	})
}

// GetPipeline returns a single pipeline by ID
func (h *Handlers) GetPipeline(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pipeline ID"})
		return
	}

	queries := sqlc.New(h.pool)
	pipeline, err := queries.GetPipeline(c.Request.Context(), pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pipeline not found"})
		return
	}

	c.JSON(http.StatusOK, transformPipeline(pipeline))
}

// CreatePipelineRequest represents the request to create a pipeline
type CreatePipelineRequest struct {
	Name         string  `json:"name" binding:"required"`
	Description  *string `json:"description"`
	PipelineType *string `json:"pipeline_type"` // sales, hiring, projects, custom
	Currency     *string `json:"currency"`
	IsDefault    bool    `json:"is_default"`
	Color        *string `json:"color"`
	Icon         *string `json:"icon"`
}

// CreatePipeline creates a new pipeline
func (h *Handlers) CreatePipeline(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var req CreatePipelineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	queries := sqlc.New(h.pool)
	pipeline, err := queries.CreatePipeline(c.Request.Context(), sqlc.CreatePipelineParams{
		UserID:       user.ID,
		Name:         req.Name,
		Description:  req.Description,
		PipelineType: req.PipelineType,
		Currency:     req.Currency,
		IsDefault:    &req.IsDefault,
		Color:        req.Color,
		Icon:         req.Icon,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create pipeline"})
		return
	}

	c.JSON(http.StatusCreated, transformPipeline(pipeline))
}

// UpdatePipelineRequest represents the request to update a pipeline
type UpdatePipelineRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description"`
	Currency    *string `json:"currency"`
	Color       *string `json:"color"`
	Icon        *string `json:"icon"`
	IsActive    bool    `json:"is_active"`
}

// UpdatePipeline updates an existing pipeline
func (h *Handlers) UpdatePipeline(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pipeline ID"})
		return
	}

	var req UpdatePipelineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	queries := sqlc.New(h.pool)
	pipeline, err := queries.UpdatePipeline(c.Request.Context(), sqlc.UpdatePipelineParams{
		ID:          pgtype.UUID{Bytes: id, Valid: true},
		Name:        req.Name,
		Description: req.Description,
		Currency:    req.Currency,
		Color:       req.Color,
		Icon:        req.Icon,
		IsActive:    &req.IsActive,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update pipeline"})
		return
	}

	c.JSON(http.StatusOK, transformPipeline(pipeline))
}

// DeletePipeline deletes a pipeline
func (h *Handlers) DeletePipeline(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pipeline ID"})
		return
	}

	queries := sqlc.New(h.pool)
	err = queries.DeletePipeline(c.Request.Context(), sqlc.DeletePipelineParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete pipeline"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pipeline deleted"})
}

// ============================================================================
// PIPELINE STAGES HANDLERS
// ============================================================================

// ListPipelineStages returns all stages for a pipeline
func (h *Handlers) ListPipelineStages(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	pipelineID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pipeline ID"})
		return
	}

	queries := sqlc.New(h.pool)
	stages, err := queries.ListPipelineStages(c.Request.Context(), pgtype.UUID{Bytes: pipelineID, Valid: true})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list stages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stages": transformPipelineStages(stages),
		"count":  len(stages),
	})
}

// CreatePipelineStageRequest represents the request to create a stage
type CreatePipelineStageRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description"`
	Position    int32   `json:"position"`
	Probability *int32  `json:"probability"`
	StageType   *string `json:"stage_type"` // open, won, lost
	RottingDays *int32  `json:"rotting_days"`
	Color       *string `json:"color"`
}

// CreatePipelineStage creates a new pipeline stage
func (h *Handlers) CreatePipelineStage(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	pipelineID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pipeline ID"})
		return
	}

	var req CreatePipelineStageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	queries := sqlc.New(h.pool)
	stage, err := queries.CreatePipelineStage(c.Request.Context(), sqlc.CreatePipelineStageParams{
		PipelineID:  pgtype.UUID{Bytes: pipelineID, Valid: true},
		Name:        req.Name,
		Description: req.Description,
		Position:    req.Position,
		Probability: req.Probability,
		StageType:   req.StageType,
		RottingDays: req.RottingDays,
		Color:       req.Color,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create stage"})
		return
	}

	c.JSON(http.StatusCreated, transformPipelineStage(stage))
}

// UpdatePipelineStageRequest represents the request to update a stage
type UpdatePipelineStageRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description"`
	Probability *int32  `json:"probability"`
	RottingDays *int32  `json:"rotting_days"`
	Color       *string `json:"color"`
}

// UpdatePipelineStage updates an existing stage
func (h *Handlers) UpdatePipelineStage(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	stageID, err := uuid.Parse(c.Param("stageId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid stage ID"})
		return
	}

	var req UpdatePipelineStageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	queries := sqlc.New(h.pool)
	stage, err := queries.UpdatePipelineStage(c.Request.Context(), sqlc.UpdatePipelineStageParams{
		ID:          pgtype.UUID{Bytes: stageID, Valid: true},
		Name:        req.Name,
		Description: req.Description,
		Probability: req.Probability,
		RottingDays: req.RottingDays,
		Color:       req.Color,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update stage"})
		return
	}

	c.JSON(http.StatusOK, transformPipelineStage(stage))
}

// ReorderPipelineStages reorders stages in a pipeline
func (h *Handlers) ReorderPipelineStages(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var req struct {
		StageOrders []struct {
			ID       string `json:"id"`
			Position int32  `json:"position"`
		} `json:"stage_orders"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	queries := sqlc.New(h.pool)
	for _, order := range req.StageOrders {
		stageID, err := uuid.Parse(order.ID)
		if err != nil {
			continue
		}
		queries.UpdateStagePosition(c.Request.Context(), sqlc.UpdateStagePositionParams{
			ID:       pgtype.UUID{Bytes: stageID, Valid: true},
			Position: order.Position,
		})
	}

	c.JSON(http.StatusOK, gin.H{"message": "Stages reordered"})
}

// DeletePipelineStage deletes a stage
func (h *Handlers) DeletePipelineStage(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	stageID, err := uuid.Parse(c.Param("stageId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid stage ID"})
		return
	}

	queries := sqlc.New(h.pool)
	err = queries.DeletePipelineStage(c.Request.Context(), pgtype.UUID{Bytes: stageID, Valid: true})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete stage"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Stage deleted"})
}

// ============================================================================
// CRM DEALS HANDLERS
// ============================================================================

// ListCRMDeals returns all CRM deals for the current user
func (h *Handlers) ListCRMDeals(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	queries := sqlc.New(h.pool)

	// Parse query params
	pipelineID := c.Query("pipeline_id")
	stageID := c.Query("stage_id")
	status := c.Query("status")
	ownerID := c.Query("owner_id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	deals, err := queries.ListCRMDeals(c.Request.Context(), sqlc.ListCRMDealsParams{
		UserID:     user.ID,
		PipelineID: crmToNullUUID(pipelineID),
		StageID:    crmToNullUUID(stageID),
		Status:     crmToNullString(status),
		OwnerID:    crmToNullString(ownerID),
		LimitVal:   int32(limit),
		OffsetVal:  int32(offset),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list deals"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"deals": transformCRMDeals(deals),
		"count": len(deals),
	})
}

// GetCRMDeal returns a single CRM deal by ID
func (h *Handlers) GetCRMDeal(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid deal ID"})
		return
	}

	queries := sqlc.New(h.pool)
	deal, err := queries.GetCRMDeal(c.Request.Context(), pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Deal not found"})
		return
	}

	c.JSON(http.StatusOK, transformCRMDealRow(deal))
}

// CreateCRMDealRequest represents the request to create a CRM deal
type CreateCRMDealRequest struct {
	PipelineID        string                 `json:"pipeline_id" binding:"required"`
	StageID           string                 `json:"stage_id" binding:"required"`
	Name              string                 `json:"name" binding:"required"`
	Description       *string                `json:"description"`
	Amount            *float64               `json:"amount"`
	Currency          *string                `json:"currency"`
	Probability       *int32                 `json:"probability"`
	ExpectedCloseDate *string                `json:"expected_close_date"`
	OwnerID           *string                `json:"owner_id"`
	CompanyID         *string                `json:"company_id"`
	PrimaryContactID  *string                `json:"primary_contact_id"`
	Status            *string                `json:"status"`
	Priority          *string                `json:"priority"`
	LeadSource        *string                `json:"lead_source"`
	CustomFields      map[string]interface{} `json:"custom_fields"`
}

// CreateCRMDeal creates a new CRM deal
func (h *Handlers) CreateCRMDeal(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var req CreateCRMDealRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pipelineID, err := uuid.Parse(req.PipelineID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pipeline ID"})
		return
	}

	stageID, err := uuid.Parse(req.StageID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid stage ID"})
		return
	}

	queries := sqlc.New(h.pool)

	// Convert custom fields to JSON
	customFields, _ := json.Marshal(req.CustomFields)

	// Parse expected close date
	var expectedCloseDate pgtype.Date
	if req.ExpectedCloseDate != nil {
		if t, err := time.Parse("2006-01-02", *req.ExpectedCloseDate); err == nil {
			expectedCloseDate = pgtype.Date{Time: t, Valid: true}
		}
	}

	deal, err := queries.CreateCRMDeal(c.Request.Context(), sqlc.CreateCRMDealParams{
		UserID:            user.ID,
		PipelineID:        pgtype.UUID{Bytes: pipelineID, Valid: true},
		StageID:           pgtype.UUID{Bytes: stageID, Valid: true},
		Name:              req.Name,
		Description:       req.Description,
		Amount:            crmToNumeric(req.Amount),
		Currency:          req.Currency,
		Probability:       req.Probability,
		ExpectedCloseDate: expectedCloseDate,
		OwnerID:           req.OwnerID,
		CompanyID:         crmToNullUUID(crmPtrToString(req.CompanyID)),
		PrimaryContactID:  crmToNullUUID(crmPtrToString(req.PrimaryContactID)),
		Status:            req.Status,
		Priority:          req.Priority,
		LeadSource:        req.LeadSource,
		CustomFields:      customFields,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create deal: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, transformCRMDealBasic(deal))
}

// UpdateCRMDealRequest represents the request to update a CRM deal
type UpdateCRMDealRequest struct {
	Name              string                 `json:"name" binding:"required"`
	Description       *string                `json:"description"`
	Amount            *float64               `json:"amount"`
	Probability       *int32                 `json:"probability"`
	ExpectedCloseDate *string                `json:"expected_close_date"`
	OwnerID           *string                `json:"owner_id"`
	CompanyID         *string                `json:"company_id"`
	PrimaryContactID  *string                `json:"primary_contact_id"`
	Priority          *string                `json:"priority"`
	CustomFields      map[string]interface{} `json:"custom_fields"`
}

// UpdateCRMDeal updates an existing CRM deal
func (h *Handlers) UpdateCRMDeal(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid deal ID"})
		return
	}

	var req UpdateCRMDealRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	queries := sqlc.New(h.pool)

	// Convert custom fields to JSON
	customFields, _ := json.Marshal(req.CustomFields)

	// Parse expected close date
	var expectedCloseDate pgtype.Date
	if req.ExpectedCloseDate != nil {
		if t, err := time.Parse("2006-01-02", *req.ExpectedCloseDate); err == nil {
			expectedCloseDate = pgtype.Date{Time: t, Valid: true}
		}
	}

	deal, err := queries.UpdateCRMDeal(c.Request.Context(), sqlc.UpdateCRMDealParams{
		ID:                pgtype.UUID{Bytes: id, Valid: true},
		Name:              req.Name,
		Description:       req.Description,
		Amount:            crmToNumeric(req.Amount),
		Probability:       req.Probability,
		ExpectedCloseDate: expectedCloseDate,
		OwnerID:           req.OwnerID,
		CompanyID:         crmToNullUUID(crmPtrToString(req.CompanyID)),
		PrimaryContactID:  crmToNullUUID(crmPtrToString(req.PrimaryContactID)),
		Priority:          req.Priority,
		CustomFields:      customFields,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update deal"})
		return
	}

	c.JSON(http.StatusOK, transformCRMDealBasic(deal))
}

// MoveCRMDealStage moves a deal to a different stage
func (h *Handlers) MoveCRMDealStage(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid deal ID"})
		return
	}

	var req struct {
		StageID string `json:"stage_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	stageID, err := uuid.Parse(req.StageID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid stage ID"})
		return
	}

	queries := sqlc.New(h.pool)
	deal, err := queries.UpdateCRMDealStage(c.Request.Context(), sqlc.UpdateCRMDealStageParams{
		ID:      pgtype.UUID{Bytes: id, Valid: true},
		StageID: pgtype.UUID{Bytes: stageID, Valid: true},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to move deal"})
		return
	}

	c.JSON(http.StatusOK, transformCRMDealBasic(deal))
}

// UpdateCRMDealStatus updates the status of a deal (open, won, lost)
func (h *Handlers) UpdateCRMDealStatus(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid deal ID"})
		return
	}

	var req struct {
		Status     string  `json:"status" binding:"required"`
		LostReason *string `json:"lost_reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	queries := sqlc.New(h.pool)
	deal, err := queries.UpdateCRMDealStatus(c.Request.Context(), sqlc.UpdateCRMDealStatusParams{
		ID:         pgtype.UUID{Bytes: id, Valid: true},
		Status:     &req.Status,
		LostReason: req.LostReason,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update deal status"})
		return
	}

	c.JSON(http.StatusOK, transformCRMDealBasic(deal))
}

// DeleteCRMDeal deletes a CRM deal
func (h *Handlers) DeleteCRMDeal(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid deal ID"})
		return
	}

	queries := sqlc.New(h.pool)
	err = queries.DeleteCRMDeal(c.Request.Context(), sqlc.DeleteCRMDealParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete deal"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Deal deleted"})
}

// GetCRMDealStats returns deal statistics for the current user
func (h *Handlers) GetCRMDealStats(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	pipelineID := c.Query("pipeline_id")

	queries := sqlc.New(h.pool)
	stats, err := queries.GetCRMDealStats(c.Request.Context(), sqlc.GetCRMDealStatsParams{
		UserID:     user.ID,
		PipelineID: crmToNullUUID(pipelineID),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get deal stats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total_deals": stats.TotalDeals,
		"open_deals":  stats.OpenDeals,
		"won_deals":   stats.WonDeals,
		"lost_deals":  stats.LostDeals,
		"open_value":  stats.OpenValue,
		"won_value":   stats.WonValue,
		"lost_value":  stats.LostValue,
	})
}

// ============================================================================
// CRM ACTIVITIES HANDLERS
// ============================================================================

// ListCRMActivities returns CRM activities for the current user
func (h *Handlers) ListCRMActivities(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	queries := sqlc.New(h.pool)

	// Parse query params
	activityType := c.Query("activity_type")
	isCompleted := c.Query("is_completed")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	var isCompletedBool *bool
	if isCompleted != "" {
		b := isCompleted == "true"
		isCompletedBool = &b
	}

	activities, err := queries.ListCRMActivities(c.Request.Context(), sqlc.ListCRMActivitiesParams{
		UserID:       user.ID,
		ActivityType: crmToNullActivityType(activityType),
		IsCompleted:  isCompletedBool,
		LimitVal:     int32(limit),
		OffsetVal:    int32(offset),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list activities"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"activities": transformCRMActivities(activities),
		"count":      len(activities),
	})
}

// ListDealActivities returns activities for a specific deal
func (h *Handlers) ListDealActivities(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	dealID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid deal ID"})
		return
	}

	queries := sqlc.New(h.pool)
	activities, err := queries.ListDealActivities(c.Request.Context(), pgtype.UUID{Bytes: dealID, Valid: true})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list activities"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"activities": transformCRMActivities(activities),
		"count":      len(activities),
	})
}

// CreateCRMActivityRequest represents the request to create a CRM activity
type CreateCRMActivityRequest struct {
	ActivityType    string   `json:"activity_type" binding:"required"`
	Subject         string   `json:"subject" binding:"required"`
	Description     *string  `json:"description"`
	Outcome         *string  `json:"outcome"`
	DealID          *string  `json:"deal_id"`
	CompanyID       *string  `json:"company_id"`
	ContactID       *string  `json:"contact_id"`
	Participants    []string `json:"participants"`
	ActivityDate    string   `json:"activity_date" binding:"required"`
	DurationMinutes *int32   `json:"duration_minutes"`
	// Call-specific
	CallDirection   *string `json:"call_direction"`
	CallDisposition *string `json:"call_disposition"`
	CallRecordingURL *string `json:"call_recording_url"`
	// Email-specific
	EmailDirection *string `json:"email_direction"`
	EmailMessageID *string `json:"email_message_id"`
	// Meeting-specific
	MeetingLocation *string `json:"meeting_location"`
	MeetingURL      *string `json:"meeting_url"`
	// Completion
	OwnerID     *string `json:"owner_id"`
	IsCompleted bool    `json:"is_completed"`
}

// CreateCRMActivity creates a new CRM activity
func (h *Handlers) CreateCRMActivity(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var req CreateCRMActivityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	queries := sqlc.New(h.pool)

	// Parse activity date
	activityDate, err := time.Parse(time.RFC3339, req.ActivityDate)
	if err != nil {
		activityDate, err = time.Parse("2006-01-02", req.ActivityDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid activity date format"})
			return
		}
	}

	// Convert participants to JSON
	participants, _ := json.Marshal(req.Participants)

	activity, err := queries.CreateCRMActivity(c.Request.Context(), sqlc.CreateCRMActivityParams{
		UserID:           user.ID,
		ActivityType:     req.ActivityType,
		Subject:          req.Subject,
		Description:      req.Description,
		Outcome:          req.Outcome,
		DealID:           crmToNullUUID(crmPtrToString(req.DealID)),
		CompanyID:        crmToNullUUID(crmPtrToString(req.CompanyID)),
		ContactID:        crmToNullUUID(crmPtrToString(req.ContactID)),
		Participants:     participants,
		ActivityDate:     pgtype.Timestamptz{Time: activityDate, Valid: true},
		DurationMinutes:  req.DurationMinutes,
		CallDirection:    req.CallDirection,
		CallDisposition:  req.CallDisposition,
		CallRecordingUrl: req.CallRecordingURL,
		EmailDirection:   req.EmailDirection,
		EmailMessageID:   req.EmailMessageID,
		MeetingLocation:  req.MeetingLocation,
		MeetingUrl:       req.MeetingURL,
		OwnerID:          req.OwnerID,
		IsCompleted:      &req.IsCompleted,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create activity: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, transformCRMActivity(activity))
}

// CompleteCRMActivity marks an activity as completed
func (h *Handlers) CompleteCRMActivity(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	activityID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid activity ID"})
		return
	}

	var req struct {
		Outcome *string `json:"outcome"`
	}
	c.ShouldBindJSON(&req)

	queries := sqlc.New(h.pool)
	activity, err := queries.CompleteCRMActivity(c.Request.Context(), sqlc.CompleteCRMActivityParams{
		ID:          pgtype.UUID{Bytes: activityID, Valid: true},
		CompletedBy: &user.ID,
		Outcome:     req.Outcome,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete activity"})
		return
	}

	c.JSON(http.StatusOK, transformCRMActivity(activity))
}

// DeleteCRMActivity deletes a CRM activity
func (h *Handlers) DeleteCRMActivity(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	activityID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid activity ID"})
		return
	}

	queries := sqlc.New(h.pool)
	err = queries.DeleteCRMActivity(c.Request.Context(), pgtype.UUID{Bytes: activityID, Valid: true})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete activity"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Activity deleted"})
}

// ============================================================================
// HELPER FUNCTIONS (CRM-specific to avoid redeclaration)
// ============================================================================

func crmToNullString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func crmToNullUUID(s string) pgtype.UUID {
	if s == "" {
		return pgtype.UUID{}
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: id, Valid: true}
}

func crmToNumeric(f *float64) pgtype.Numeric {
	if f == nil {
		return pgtype.Numeric{}
	}
	var n pgtype.Numeric
	n.Scan(*f)
	return n
}

func crmNumericToFloat(n pgtype.Numeric) float64 {
	if !n.Valid {
		return 0
	}
	f, _ := n.Float64Value()
	return f.Float64
}

func crmPtrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func crmToNullActivityType(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// ============================================================================
// TRANSFORM FUNCTIONS
// ============================================================================

func transformCompany(c sqlc.Company) gin.H {
	return gin.H{
		"id":               crmUuidToString(c.ID),
		"user_id":          c.UserID,
		"name":             c.Name,
		"legal_name":       c.LegalName,
		"industry":         c.Industry,
		"company_size":     c.CompanySize,
		"website":          c.Website,
		"email":            c.Email,
		"phone":            c.Phone,
		"address_line1":    c.AddressLine1,
		"address_line2":    c.AddressLine2,
		"city":             c.City,
		"state":            c.State,
		"postal_code":      c.PostalCode,
		"country":          c.Country,
		"annual_revenue":   crmNumericToFloat(c.AnnualRevenue),
		"currency":         c.Currency,
		"linkedin_url":     c.LinkedinUrl,
		"twitter_handle":   c.TwitterHandle,
		"owner_id":         c.OwnerID,
		"lifecycle_stage":  c.LifecycleStage,
		"lead_source":      c.LeadSource,
		"health_score":     c.HealthScore,
		"engagement_score": c.EngagementScore,
		"logo_url":         c.LogoUrl,
		"custom_fields":    crmJsonToMap(c.CustomFields),
		"metadata":         crmJsonToMap(c.Metadata),
		"created_at":       c.CreatedAt.Time,
		"updated_at":       c.UpdatedAt.Time,
	}
}

func transformCompanies(companies []sqlc.Company) []gin.H {
	result := make([]gin.H, len(companies))
	for i, c := range companies {
		result[i] = transformCompany(c)
	}
	return result
}

func transformPipeline(p sqlc.Pipeline) gin.H {
	return gin.H{
		"id":            crmUuidToString(p.ID),
		"user_id":       p.UserID,
		"name":          p.Name,
		"description":   p.Description,
		"pipeline_type": p.PipelineType,
		"currency":      p.Currency,
		"is_default":    p.IsDefault,
		"is_active":     p.IsActive,
		"color":         p.Color,
		"icon":          p.Icon,
		"created_at":    p.CreatedAt.Time,
		"updated_at":    p.UpdatedAt.Time,
	}
}

func transformPipelines(pipelines []sqlc.Pipeline) []gin.H {
	result := make([]gin.H, len(pipelines))
	for i, p := range pipelines {
		result[i] = transformPipeline(p)
	}
	return result
}

func transformPipelineStage(s sqlc.PipelineStage) gin.H {
	return gin.H{
		"id":           crmUuidToString(s.ID),
		"pipeline_id":  crmUuidToString(s.PipelineID),
		"name":         s.Name,
		"description":  s.Description,
		"position":     s.Position,
		"probability":  s.Probability,
		"stage_type":   s.StageType,
		"rotting_days": s.RottingDays,
		"color":        s.Color,
		"created_at":   s.CreatedAt.Time,
		"updated_at":   s.UpdatedAt.Time,
	}
}

func transformPipelineStages(stages []sqlc.PipelineStage) []gin.H {
	result := make([]gin.H, len(stages))
	for i, s := range stages {
		result[i] = transformPipelineStage(s)
	}
	return result
}

func transformCRMDeal(d sqlc.ListCRMDealsRow) gin.H {
	return gin.H{
		"id":                  crmUuidToString(d.ID),
		"user_id":             d.UserID,
		"pipeline_id":         crmUuidToString(d.PipelineID),
		"pipeline_name":       d.PipelineName,
		"stage_id":            crmUuidToString(d.StageID),
		"stage_name":          d.StageName,
		"name":                d.Name,
		"description":         d.Description,
		"amount":              crmNumericToFloat(d.Amount),
		"currency":            d.Currency,
		"probability":         d.Probability,
		"expected_close_date": crmDateToString(d.ExpectedCloseDate),
		"actual_close_date":   crmDateToString(d.ActualCloseDate),
		"owner_id":            d.OwnerID,
		"company_id":          crmUuidToString(d.CompanyID),
		"company_name":        d.CompanyName,
		"primary_contact_id":  crmUuidToString(d.PrimaryContactID),
		"status":              d.Status,
		"lost_reason":         d.LostReason,
		"priority":            d.Priority,
		"lead_source":         d.LeadSource,
		"deal_score":          d.DealScore,
		"custom_fields":       crmJsonToMap(d.CustomFields),
		"created_at":          d.CreatedAt.Time,
		"updated_at":          d.UpdatedAt.Time,
	}
}

func transformCRMDealRow(d sqlc.GetCRMDealRow) gin.H {
	return gin.H{
		"id":                  crmUuidToString(d.ID),
		"user_id":             d.UserID,
		"pipeline_id":         crmUuidToString(d.PipelineID),
		"pipeline_name":       d.PipelineName,
		"stage_id":            crmUuidToString(d.StageID),
		"stage_name":          d.StageName,
		"name":                d.Name,
		"description":         d.Description,
		"amount":              crmNumericToFloat(d.Amount),
		"currency":            d.Currency,
		"probability":         d.Probability,
		"expected_close_date": crmDateToString(d.ExpectedCloseDate),
		"actual_close_date":   crmDateToString(d.ActualCloseDate),
		"owner_id":            d.OwnerID,
		"company_id":          crmUuidToString(d.CompanyID),
		"company_name":        d.CompanyName,
		"primary_contact_id":  crmUuidToString(d.PrimaryContactID),
		"status":              d.Status,
		"lost_reason":         d.LostReason,
		"priority":            d.Priority,
		"lead_source":         d.LeadSource,
		"deal_score":          d.DealScore,
		"custom_fields":       crmJsonToMap(d.CustomFields),
		"created_at":          d.CreatedAt.Time,
		"updated_at":          d.UpdatedAt.Time,
	}
}

func transformCRMDeals(deals []sqlc.ListCRMDealsRow) []gin.H {
	result := make([]gin.H, len(deals))
	for i, d := range deals {
		result[i] = transformCRMDeal(d)
	}
	return result
}

func transformCRMDealBasic(d sqlc.Deal) gin.H {
	return gin.H{
		"id":                  crmUuidToString(d.ID),
		"user_id":             d.UserID,
		"pipeline_id":         crmUuidToString(d.PipelineID),
		"stage_id":            crmUuidToString(d.StageID),
		"name":                d.Name,
		"description":         d.Description,
		"amount":              crmNumericToFloat(d.Amount),
		"currency":            d.Currency,
		"probability":         d.Probability,
		"expected_close_date": crmDateToString(d.ExpectedCloseDate),
		"actual_close_date":   crmDateToString(d.ActualCloseDate),
		"owner_id":            d.OwnerID,
		"company_id":          crmUuidToString(d.CompanyID),
		"primary_contact_id":  crmUuidToString(d.PrimaryContactID),
		"status":              d.Status,
		"lost_reason":         d.LostReason,
		"priority":            d.Priority,
		"lead_source":         d.LeadSource,
		"deal_score":          d.DealScore,
		"custom_fields":       crmJsonToMap(d.CustomFields),
		"created_at":          d.CreatedAt.Time,
		"updated_at":          d.UpdatedAt.Time,
	}
}

func transformCRMActivity(a sqlc.CrmActivity) gin.H {
	return gin.H{
		"id":                 crmUuidToString(a.ID),
		"user_id":            a.UserID,
		"activity_type":      a.ActivityType,
		"subject":            a.Subject,
		"description":        a.Description,
		"outcome":            a.Outcome,
		"deal_id":            crmUuidToString(a.DealID),
		"company_id":         crmUuidToString(a.CompanyID),
		"contact_id":         crmUuidToString(a.ContactID),
		"participants":       crmJsonToSlice(a.Participants),
		"activity_date":      crmTimestampToString(a.ActivityDate),
		"duration_minutes":   a.DurationMinutes,
		"call_direction":     a.CallDirection,
		"call_disposition":   a.CallDisposition,
		"call_recording_url": a.CallRecordingUrl,
		"email_direction":    a.EmailDirection,
		"email_message_id":   a.EmailMessageID,
		"meeting_location":   a.MeetingLocation,
		"meeting_url":        a.MeetingUrl,
		"owner_id":           a.OwnerID,
		"is_completed":       a.IsCompleted,
		"completed_by":       a.CompletedBy,
		"completed_at":       crmTimestampToString(a.CompletedAt),
		"created_at":         a.CreatedAt.Time,
		"updated_at":         a.UpdatedAt.Time,
	}
}

func transformCRMActivities(activities []sqlc.CrmActivity) []gin.H {
	result := make([]gin.H, len(activities))
	for i, a := range activities {
		result[i] = transformCRMActivity(a)
	}
	return result
}

func crmUuidToString(u pgtype.UUID) *string {
	if !u.Valid {
		return nil
	}
	s := uuid.UUID(u.Bytes).String()
	return &s
}

func crmDateToString(d pgtype.Date) *string {
	if !d.Valid {
		return nil
	}
	s := d.Time.Format("2006-01-02")
	return &s
}

func crmTimestampToString(t pgtype.Timestamptz) *string {
	if !t.Valid {
		return nil
	}
	s := t.Time.Format(time.RFC3339)
	return &s
}

func crmJsonToMap(b []byte) map[string]interface{} {
	if len(b) == 0 {
		return nil
	}
	var m map[string]interface{}
	json.Unmarshal(b, &m)
	return m
}

func crmJsonToSlice(b []byte) []interface{} {
	if len(b) == 0 {
		return nil
	}
	var s []interface{}
	json.Unmarshal(b, &s)
	return s
}
