package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/database/sqlc"
	"github.com/rhl/businessos-backend/internal/middleware"
	"github.com/rhl/businessos-backend/internal/utils"
)

// ClientHandler handles client management operations
type ClientHandler struct {
	pool *pgxpool.Pool
}

// NewClientHandler creates a new ClientHandler
func NewClientHandler(pool *pgxpool.Pool) *ClientHandler {
	return &ClientHandler{pool: pool}
}

// ListClients returns all clients for the current user
func (h *ClientHandler) ListClients(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.RespondUnauthorized(c, slog.Default())
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

	pg := ParsePagination(c)

	clients, err := queries.ListClients(c.Request.Context(), sqlc.ListClientsParams{
		UserID:     user.ID,
		Status:     sqlc.NullClientstatus{Clientstatus: status, Valid: c.Query("status") != ""},
		ClientType: sqlc.NullClienttype{Clienttype: clientType, Valid: c.Query("type") != ""},
		Search:     utils.StringPtr(search),
	})
	if err != nil {
		utils.RespondInternalError(c, slog.Default(), "list clients", nil)
		return
	}

	// Apply in-memory pagination (SQL query has no LIMIT/OFFSET; all matching rows fetched)
	all := TransformClients(clients)
	total := int64(len(all))
	start := int(pg.Offset)
	end := start + int(pg.Limit)
	if start > len(all) {
		start = len(all)
	}
	if end > len(all) {
		end = len(all)
	}

	c.JSON(http.StatusOK, NewPaginatedResponse(all[start:end], total, pg))
}

// CreateClient creates a new client
func (h *ClientHandler) CreateClient(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.RespondUnauthorized(c, slog.Default())
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
		utils.RespondInvalidRequest(c, slog.Default(), err)
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

	// Handle JSON fields (nil for empty — SimpleProtocol compatibility)
	var tags []byte
	if req.Tags != nil {
		tags = req.Tags
	}
	var customFields []byte
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
		utils.RespondInternalError(c, slog.Default(), "create client", nil)
		return
	}

	c.JSON(http.StatusCreated, TransformClient(client))
}

// GetClient returns a single client with its contacts, interactions, and deals
func (h *ClientHandler) GetClient(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondInvalidID(c, slog.Default(), "client_id")
		return
	}

	pgID := pgtype.UUID{Bytes: id, Valid: true}
	queries := sqlc.New(h.pool)
	client, err := queries.GetClient(c.Request.Context(), sqlc.GetClientParams{
		ID:     pgID,
		UserID: user.ID,
	})
	if err != nil {
		utils.RespondNotFound(c, slog.Default(), "Client")
		return
	}

	// Fetch related data — errors are non-fatal, return empty slices
	contacts, err := queries.ListClientContacts(c.Request.Context(), pgID)
	if err != nil {
		slog.Warn("failed to load client contacts", "client_id", id, "error", err)
		contacts = nil
	}
	interactions, err := queries.ListClientInteractions(c.Request.Context(), pgID)
	if err != nil {
		slog.Warn("failed to load client interactions", "client_id", id, "error", err)
		interactions = nil
	}
	deals, err := queries.ListClientDeals(c.Request.Context(), pgID)
	if err != nil {
		slog.Warn("failed to load client deals", "client_id", id, "error", err)
		deals = nil
	}

	resp := TransformClient(client)
	c.JSON(http.StatusOK, gin.H{
		"id":                resp.ID,
		"user_id":           resp.UserID,
		"name":              resp.Name,
		"type":              resp.Type,
		"email":             resp.Email,
		"phone":             resp.Phone,
		"website":           resp.Website,
		"industry":          resp.Industry,
		"company_size":      resp.CompanySize,
		"address":           resp.Address,
		"city":              resp.City,
		"state":             resp.State,
		"zip_code":          resp.ZipCode,
		"country":           resp.Country,
		"status":            resp.Status,
		"source":            resp.Source,
		"assigned_to":       resp.AssignedTo,
		"lifetime_value":    resp.LifetimeValue,
		"tags":              resp.Tags,
		"custom_fields":     resp.CustomFields,
		"notes":             resp.Notes,
		"created_at":        resp.CreatedAt,
		"updated_at":        resp.UpdatedAt,
		"last_contacted_at": resp.LastContactedAt,
		"contacts":          TransformContacts(contacts),
		"interactions":      TransformInteractions(interactions),
		"deals":             transformClientDealsFromRows(deals),
	})
}

// UpdateClient updates an existing client
func (h *ClientHandler) UpdateClient(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondInvalidID(c, slog.Default(), "client_id")
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
		utils.RespondInvalidRequest(c, slog.Default(), err)
		return
	}

	queries := sqlc.New(h.pool)

	// Get existing client first
	existing, err := queries.GetClient(c.Request.Context(), sqlc.GetClientParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		utils.RespondNotFound(c, slog.Default(), "Client")
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
		utils.RespondInternalError(c, slog.Default(), "update client", nil)
		return
	}

	c.JSON(http.StatusOK, TransformClient(client))
}

// UpdateClientStatus updates only the status of a client
func (h *ClientHandler) UpdateClientStatus(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondInvalidID(c, slog.Default(), "client_id")
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, slog.Default(), err)
		return
	}

	queries := sqlc.New(h.pool)

	// Verify ownership
	_, err = queries.GetClient(c.Request.Context(), sqlc.GetClientParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		utils.RespondNotFound(c, slog.Default(), "Client")
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
		utils.RespondInternalError(c, slog.Default(), "update client status", nil)
		return
	}

	c.JSON(http.StatusOK, TransformClient(client))
}

// DeleteClient deletes a client
func (h *ClientHandler) DeleteClient(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondInvalidID(c, slog.Default(), "client_id")
		return
	}

	queries := sqlc.New(h.pool)
	err = queries.DeleteClient(c.Request.Context(), sqlc.DeleteClientParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		utils.RespondInternalError(c, slog.Default(), "delete client", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Client deleted"})
}

// RegisterClientRoutes registers all client management routes on the given router group.
func RegisterClientRoutes(api *gin.RouterGroup, h *ClientHandler, auth gin.HandlerFunc) {
	clients := api.Group("/clients")
	clients.Use(auth, middleware.RequireAuth())
	{
		clients.GET("", h.ListClients)
		clients.POST("", h.CreateClient)
		clients.GET("/:id", h.GetClient)
		clients.PUT("/:id", h.UpdateClient)
		clients.PATCH("/:id/status", h.UpdateClientStatus)
		clients.DELETE("/:id", h.DeleteClient)
		// Contacts
		clients.GET("/:id/contacts", h.ListClientContacts)
		clients.POST("/:id/contacts", h.CreateClientContact)
		clients.PUT("/:id/contacts/:contactId", h.UpdateClientContact)
		clients.DELETE("/:id/contacts/:contactId", h.DeleteClientContact)
		// Interactions
		clients.GET("/:id/interactions", h.ListClientInteractions)
		clients.POST("/:id/interactions", h.CreateClientInteraction)
		// Deals
		clients.GET("/:id/deals", h.ListClientDeals)
		clients.POST("/:id/deals", h.CreateClientDeal)
		clients.PUT("/:id/deals/:dealId", h.UpdateClientDeal)
	}
}
