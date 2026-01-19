package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rhl/businessos-backend/internal/database/sqlc"
	"github.com/rhl/businessos-backend/internal/middleware"
)

// ListClients returns all clients for the current user
func (h *Handlers) ListClients(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	queries := sqlc.New(h.pool)

	// Parse optional filters
	var status sqlc.Clientstatus
	if s := c.Query("status"); s != "" {
		status = stringToClientStatus(s)
	}

	var clientType sqlc.Clienttype
	if t := c.Query("type"); t != "" {
		clientType = stringToClientType(t)
	}

	search := c.Query("search")

	clients, err := queries.ListClients(c.Request.Context(), sqlc.ListClientsParams{
		UserID:     user.ID,
		Status:     sqlc.NullClientstatus{Clientstatus: status, Valid: c.Query("status") != ""},
		ClientType: sqlc.NullClienttype{Clienttype: clientType, Valid: c.Query("type") != ""},
		Search:     stringPtr(search),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list clients"})
		return
	}

	c.JSON(http.StatusOK, TransformClients(clients))
}

// CreateClient creates a new client
func (h *Handlers) CreateClient(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var req struct {
		Name         string          `json:"name" binding:"required"`
		Type         *string         `json:"type"`
		Email        *string         `json:"email"`
		Phone        *string         `json:"phone"`
		Website      *string         `json:"website"`
		Industry     *string         `json:"industry"`
		CompanySize  *string         `json:"company_size"`
		Address      *string         `json:"address"`
		City         *string         `json:"city"`
		State        *string         `json:"state"`
		ZipCode      *string         `json:"zip_code"`
		Country      *string         `json:"country"`
		Status       *string         `json:"status"`
		Source       *string         `json:"source"`
		AssignedTo   *string         `json:"assigned_to"`
		Tags         json.RawMessage `json:"tags"`
		CustomFields json.RawMessage `json:"custom_fields"`
		Notes        *string         `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	queries := sqlc.New(h.pool)

	// Parse type
	var clientType sqlc.NullClienttype
	if req.Type != nil {
		clientType = sqlc.NullClienttype{
			Clienttype: stringToClientType(*req.Type),
			Valid:      true,
		}
	}

	// Parse status
	var status sqlc.NullClientstatus
	if req.Status != nil {
		status = sqlc.NullClientstatus{
			Clientstatus: stringToClientStatus(*req.Status),
			Valid:        true,
		}
	}

	// Handle JSON fields
	tags := []byte("[]")
	if req.Tags != nil {
		tags = req.Tags
	}
	customFields := []byte("{}")
	if req.CustomFields != nil {
		customFields = req.CustomFields
	}

	client, err := queries.CreateClient(c.Request.Context(), sqlc.CreateClientParams{
		UserID:       user.ID,
		Name:         req.Name,
		Type:         clientType,
		Email:        req.Email,
		Phone:        req.Phone,
		Website:      req.Website,
		Industry:     req.Industry,
		CompanySize:  req.CompanySize,
		Address:      req.Address,
		City:         req.City,
		State:        req.State,
		ZipCode:      req.ZipCode,
		Country:      req.Country,
		Status:       status,
		Source:       req.Source,
		AssignedTo:   req.AssignedTo,
		Tags:         tags,
		CustomFields: customFields,
		Notes:        req.Notes,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create client"})
		return
	}

	c.JSON(http.StatusCreated, TransformClient(client))
}

// GetClient returns a single client
func (h *Handlers) GetClient(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid client ID"})
		return
	}

	queries := sqlc.New(h.pool)
	client, err := queries.GetClient(c.Request.Context(), sqlc.GetClientParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Client not found"})
		return
	}

	c.JSON(http.StatusOK, TransformClient(client))
}

// UpdateClient updates an existing client
func (h *Handlers) UpdateClient(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid client ID"})
		return
	}

	var req struct {
		Name         *string         `json:"name"`
		Type         *string         `json:"type"`
		Email        *string         `json:"email"`
		Phone        *string         `json:"phone"`
		Website      *string         `json:"website"`
		Industry     *string         `json:"industry"`
		CompanySize  *string         `json:"company_size"`
		Address      *string         `json:"address"`
		City         *string         `json:"city"`
		State        *string         `json:"state"`
		ZipCode      *string         `json:"zip_code"`
		Country      *string         `json:"country"`
		Status       *string         `json:"status"`
		Source       *string         `json:"source"`
		AssignedTo   *string         `json:"assigned_to"`
		Tags         json.RawMessage `json:"tags"`
		CustomFields json.RawMessage `json:"custom_fields"`
		Notes        *string         `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	queries := sqlc.New(h.pool)

	// Get existing client first
	existing, err := queries.GetClient(c.Request.Context(), sqlc.GetClientParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Client not found"})
		return
	}

	// Build update params with existing values as defaults
	name := existing.Name
	if req.Name != nil {
		name = *req.Name
	}

	clientType := existing.Type
	if req.Type != nil {
		clientType = sqlc.NullClienttype{
			Clienttype: stringToClientType(*req.Type),
			Valid:      true,
		}
	}

	status := existing.Status
	if req.Status != nil {
		status = sqlc.NullClientstatus{
			Clientstatus: stringToClientStatus(*req.Status),
			Valid:        true,
		}
	}

	email := existing.Email
	if req.Email != nil {
		email = req.Email
	}

	phone := existing.Phone
	if req.Phone != nil {
		phone = req.Phone
	}

	website := existing.Website
	if req.Website != nil {
		website = req.Website
	}

	industry := existing.Industry
	if req.Industry != nil {
		industry = req.Industry
	}

	companySize := existing.CompanySize
	if req.CompanySize != nil {
		companySize = req.CompanySize
	}

	address := existing.Address
	if req.Address != nil {
		address = req.Address
	}

	city := existing.City
	if req.City != nil {
		city = req.City
	}

	state := existing.State
	if req.State != nil {
		state = req.State
	}

	zipCode := existing.ZipCode
	if req.ZipCode != nil {
		zipCode = req.ZipCode
	}

	country := existing.Country
	if req.Country != nil {
		country = req.Country
	}

	source := existing.Source
	if req.Source != nil {
		source = req.Source
	}

	assignedTo := existing.AssignedTo
	if req.AssignedTo != nil {
		assignedTo = req.AssignedTo
	}

	tags := existing.Tags
	if req.Tags != nil {
		tags = req.Tags
	}

	customFields := existing.CustomFields
	if req.CustomFields != nil {
		customFields = req.CustomFields
	}

	notes := existing.Notes
	if req.Notes != nil {
		notes = req.Notes
	}

	client, err := queries.UpdateClient(c.Request.Context(), sqlc.UpdateClientParams{
		ID:           pgtype.UUID{Bytes: id, Valid: true},
		Name:         name,
		Type:         clientType,
		Email:        email,
		Phone:        phone,
		Website:      website,
		Industry:     industry,
		CompanySize:  companySize,
		Address:      address,
		City:         city,
		State:        state,
		ZipCode:      zipCode,
		Country:      country,
		Status:       status,
		Source:       source,
		AssignedTo:   assignedTo,
		Tags:         tags,
		CustomFields: customFields,
		Notes:        notes,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update client"})
		return
	}

	c.JSON(http.StatusOK, TransformClient(client))
}

// UpdateClientStatus updates only the status of a client
func (h *Handlers) UpdateClientStatus(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid client ID"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	queries := sqlc.New(h.pool)

	// Verify ownership
	_, err = queries.GetClient(c.Request.Context(), sqlc.GetClientParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Client not found"})
		return
	}

	client, err := queries.UpdateClientStatus(c.Request.Context(), sqlc.UpdateClientStatusParams{
		ID: pgtype.UUID{Bytes: id, Valid: true},
		Status: sqlc.NullClientstatus{
			Clientstatus: stringToClientStatus(req.Status),
			Valid:        true,
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update client status"})
		return
	}

	c.JSON(http.StatusOK, TransformClient(client))
}

// DeleteClient deletes a client
func (h *Handlers) DeleteClient(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid client ID"})
		return
	}

	queries := sqlc.New(h.pool)
	err = queries.DeleteClient(c.Request.Context(), sqlc.DeleteClientParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete client"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Client deleted"})
}

// ListClientContacts returns all contacts for a client
func (h *Handlers) ListClientContacts(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid client ID"})
		return
	}

	queries := sqlc.New(h.pool)

	// Verify client ownership
	_, err = queries.GetClient(c.Request.Context(), sqlc.GetClientParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Client not found"})
		return
	}

	contacts, err := queries.ListClientContacts(c.Request.Context(), pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list contacts"})
		return
	}

	c.JSON(http.StatusOK, TransformContacts(contacts))
}

// CreateClientContact creates a new contact for a client
func (h *Handlers) CreateClientContact(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid client ID"})
		return
	}

	var req struct {
		Name      string  `json:"name" binding:"required"`
		Email     *string `json:"email"`
		Phone     *string `json:"phone"`
		Role      *string `json:"role"`
		IsPrimary *bool   `json:"is_primary"`
		Notes     *string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	queries := sqlc.New(h.pool)

	// Verify client ownership
	_, err = queries.GetClient(c.Request.Context(), sqlc.GetClientParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Client not found"})
		return
	}

	contact, err := queries.CreateClientContact(c.Request.Context(), sqlc.CreateClientContactParams{
		ClientID:  pgtype.UUID{Bytes: id, Valid: true},
		Name:      req.Name,
		Email:     req.Email,
		Phone:     req.Phone,
		Role:      req.Role,
		IsPrimary: req.IsPrimary,
		Notes:     req.Notes,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create contact"})
		return
	}

	c.JSON(http.StatusCreated, TransformContact(contact))
}

// UpdateClientContact updates a client contact
func (h *Handlers) UpdateClientContact(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	clientID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid client ID"})
		return
	}

	contactID, err := uuid.Parse(c.Param("contact_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid contact ID"})
		return
	}

	var req struct {
		Name      string  `json:"name" binding:"required"`
		Email     *string `json:"email"`
		Phone     *string `json:"phone"`
		Role      *string `json:"role"`
		IsPrimary *bool   `json:"is_primary"`
		Notes     *string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	queries := sqlc.New(h.pool)

	// Verify client ownership
	_, err = queries.GetClient(c.Request.Context(), sqlc.GetClientParams{
		ID:     pgtype.UUID{Bytes: clientID, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Client not found"})
		return
	}

	contact, err := queries.UpdateClientContact(c.Request.Context(), sqlc.UpdateClientContactParams{
		ID:        pgtype.UUID{Bytes: contactID, Valid: true},
		Name:      req.Name,
		Email:     req.Email,
		Phone:     req.Phone,
		Role:      req.Role,
		IsPrimary: req.IsPrimary,
		Notes:     req.Notes,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update contact"})
		return
	}

	c.JSON(http.StatusOK, TransformContact(contact))
}

// DeleteClientContact deletes a client contact
func (h *Handlers) DeleteClientContact(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	clientID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid client ID"})
		return
	}

	contactID, err := uuid.Parse(c.Param("contact_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid contact ID"})
		return
	}

	queries := sqlc.New(h.pool)

	// Verify client ownership
	_, err = queries.GetClient(c.Request.Context(), sqlc.GetClientParams{
		ID:     pgtype.UUID{Bytes: clientID, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Client not found"})
		return
	}

	err = queries.DeleteClientContact(c.Request.Context(), sqlc.DeleteClientContactParams{
		ID:       pgtype.UUID{Bytes: contactID, Valid: true},
		ClientID: pgtype.UUID{Bytes: clientID, Valid: true},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete contact"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Contact deleted"})
}

// ListClientInteractions returns all interactions for a client
func (h *Handlers) ListClientInteractions(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid client ID"})
		return
	}

	queries := sqlc.New(h.pool)

	// Verify client ownership
	_, err = queries.GetClient(c.Request.Context(), sqlc.GetClientParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Client not found"})
		return
	}

	interactions, err := queries.ListClientInteractions(c.Request.Context(), pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list interactions"})
		return
	}

	c.JSON(http.StatusOK, TransformInteractions(interactions))
}

// CreateClientInteraction creates a new interaction for a client
func (h *Handlers) CreateClientInteraction(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid client ID"})
		return
	}

	var req struct {
		ContactID   *string `json:"contact_id"`
		Type        string  `json:"type" binding:"required"`
		Subject     string  `json:"subject" binding:"required"`
		Description *string `json:"description"`
		Outcome     *string `json:"outcome"`
		OccurredAt  *string `json:"occurred_at"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	queries := sqlc.New(h.pool)

	// Verify client ownership
	_, err = queries.GetClient(c.Request.Context(), sqlc.GetClientParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Client not found"})
		return
	}

	// Parse contact ID if provided
	var contactID pgtype.UUID
	if req.ContactID != nil {
		if parsed, err := uuid.Parse(*req.ContactID); err == nil {
			contactID = pgtype.UUID{Bytes: parsed, Valid: true}
		}
	}

	// Parse occurred_at or use now
	occurredAt := pgtype.Timestamptz{Time: time.Now(), Valid: true}
	if req.OccurredAt != nil {
		if t, err := time.Parse(time.RFC3339, *req.OccurredAt); err == nil {
			occurredAt = pgtype.Timestamptz{Time: t, Valid: true}
		}
	}

	interaction, err := queries.CreateClientInteraction(c.Request.Context(), sqlc.CreateClientInteractionParams{
		ClientID:    pgtype.UUID{Bytes: id, Valid: true},
		ContactID:   contactID,
		Type:        stringToInteractionType(req.Type),
		Subject:     req.Subject,
		Description: req.Description,
		Outcome:     req.Outcome,
		OccurredAt:  occurredAt,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create interaction"})
		return
	}

	c.JSON(http.StatusCreated, TransformInteraction(interaction))
}

// ListClientDeals returns all deals for a client
func (h *Handlers) ListClientDeals(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid client ID"})
		return
	}

	queries := sqlc.New(h.pool)

	// Verify client ownership
	_, err = queries.GetClient(c.Request.Context(), sqlc.GetClientParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Client not found"})
		return
	}

	deals, err := queries.ListClientDeals(c.Request.Context(), pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list deals"})
		return
	}

	c.JSON(http.StatusOK, TransformDeals(deals))
}

// CreateClientDeal creates a new deal for a client
func (h *Handlers) CreateClientDeal(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid client ID"})
		return
	}

	var req struct {
		Name              string  `json:"name" binding:"required"`
		Value             float64 `json:"value"`
		Stage             *string `json:"stage"`
		Probability       *int32  `json:"probability"`
		ExpectedCloseDate *string `json:"expected_close_date"`
		Notes             *string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	queries := sqlc.New(h.pool)

	// Verify client ownership
	_, err = queries.GetClient(c.Request.Context(), sqlc.GetClientParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Client not found"})
		return
	}

	// Parse stage
	var stage sqlc.NullDealstage
	if req.Stage != nil {
		stage = sqlc.NullDealstage{
			Dealstage: stringToDealStage(*req.Stage),
			Valid:     true,
		}
	}

	// Parse expected close date
	var expectedCloseDate pgtype.Date
	if req.ExpectedCloseDate != nil {
		if t, err := time.Parse("2006-01-02", *req.ExpectedCloseDate); err == nil {
			expectedCloseDate = pgtype.Date{Time: t, Valid: true}
		}
	}

	// Convert value to pgtype.Numeric
	value := pgtype.Numeric{}
	value.Scan(req.Value)

	deal, err := queries.CreateClientDeal(c.Request.Context(), sqlc.CreateClientDealParams{
		ClientID:          pgtype.UUID{Bytes: id, Valid: true},
		Name:              req.Name,
		Value:             value,
		Stage:             stage,
		Probability:       req.Probability,
		ExpectedCloseDate: expectedCloseDate,
		Notes:             req.Notes,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create deal"})
		return
	}

	c.JSON(http.StatusCreated, TransformDeal(deal))
}

// UpdateClientDeal updates a client deal
func (h *Handlers) UpdateClientDeal(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	clientID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid client ID"})
		return
	}

	dealID, err := uuid.Parse(c.Param("deal_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid deal ID"})
		return
	}

	var req struct {
		Name              string  `json:"name" binding:"required"`
		Value             float64 `json:"value"`
		Stage             *string `json:"stage"`
		Probability       *int32  `json:"probability"`
		ExpectedCloseDate *string `json:"expected_close_date"`
		Notes             *string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	queries := sqlc.New(h.pool)

	// Verify client ownership
	_, err = queries.GetClient(c.Request.Context(), sqlc.GetClientParams{
		ID:     pgtype.UUID{Bytes: clientID, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Client not found"})
		return
	}

	// Parse stage
	var stage sqlc.NullDealstage
	if req.Stage != nil {
		stage = sqlc.NullDealstage{
			Dealstage: stringToDealStage(*req.Stage),
			Valid:     true,
		}
	}

	// Parse expected close date
	var expectedCloseDate pgtype.Date
	if req.ExpectedCloseDate != nil {
		if t, err := time.Parse("2006-01-02", *req.ExpectedCloseDate); err == nil {
			expectedCloseDate = pgtype.Date{Time: t, Valid: true}
		}
	}

	// Convert value to pgtype.Numeric
	value := pgtype.Numeric{}
	value.Scan(req.Value)

	deal, err := queries.UpdateClientDeal(c.Request.Context(), sqlc.UpdateClientDealParams{
		ID:                pgtype.UUID{Bytes: dealID, Valid: true},
		Name:              req.Name,
		Value:             value,
		Stage:             stage,
		Probability:       req.Probability,
		ExpectedCloseDate: expectedCloseDate,
		Notes:             req.Notes,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update deal"})
		return
	}

	c.JSON(http.StatusOK, TransformDeal(deal))
}

// stringToClientType converts a string to sqlc.Clienttype
func stringToClientType(t string) sqlc.Clienttype {
	typeMap := map[string]sqlc.Clienttype{
		"company":    sqlc.ClienttypeCompany,
		"individual": sqlc.ClienttypeIndividual,
	}
	if enum, ok := typeMap[strings.ToLower(t)]; ok {
		return enum
	}
	return sqlc.ClienttypeCompany
}

// stringToClientStatus converts a string to sqlc.Clientstatus
func stringToClientStatus(s string) sqlc.Clientstatus {
	typeMap := map[string]sqlc.Clientstatus{
		"lead":     sqlc.ClientstatusLead,
		"prospect": sqlc.ClientstatusProspect,
		"active":   sqlc.ClientstatusActive,
		"inactive": sqlc.ClientstatusInactive,
		"churned":  sqlc.ClientstatusChurned,
	}
	if enum, ok := typeMap[strings.ToLower(s)]; ok {
		return enum
	}
	return sqlc.ClientstatusActive
}

// stringToInteractionType converts a string to sqlc.Interactiontype
func stringToInteractionType(t string) sqlc.Interactiontype {
	typeMap := map[string]sqlc.Interactiontype{
		"call":    sqlc.InteractiontypeCall,
		"email":   sqlc.InteractiontypeEmail,
		"meeting": sqlc.InteractiontypeMeeting,
		"note":    sqlc.InteractiontypeNote,
	}
	if enum, ok := typeMap[strings.ToLower(t)]; ok {
		return enum
	}
	return sqlc.InteractiontypeNote
}

// stringToDealStage converts a string to sqlc.Dealstage
func stringToDealStage(s string) sqlc.Dealstage {
	typeMap := map[string]sqlc.Dealstage{
		"qualification": sqlc.DealstageQualification,
		"proposal":      sqlc.DealstageProposal,
		"negotiation":   sqlc.DealstageNegotiation,
		"closed_won":    sqlc.DealstageClosedWon,
		"closed_lost":   sqlc.DealstageClosedLost,
	}
	if enum, ok := typeMap[strings.ToLower(s)]; ok {
		return enum
	}
	return sqlc.DealstageQualification
}
