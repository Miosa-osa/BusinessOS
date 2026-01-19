package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/rhl/businessos-backend/internal/database/sqlc"
	"github.com/rhl/businessos-backend/internal/middleware"
)

// ============================================================================
// API Request/Response Types for Tables Module
// ============================================================================

// TableResponse represents a table in API responses
type TableResponse struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description *string                `json:"description,omitempty"`
	Icon        *string                `json:"icon,omitempty"`
	Color       *string                `json:"color,omitempty"`
	WorkspaceID *string                `json:"workspace_id,omitempty"`
	Settings    map[string]interface{} `json:"settings,omitempty"`
	RowCount    int64                  `json:"row_count"`
	CreatedAt   string                 `json:"created_at"`
	UpdatedAt   string                 `json:"updated_at"`
	// Nested data for detail view
	Columns []FieldResponse `json:"columns,omitempty"`
	Views   []ViewResponse  `json:"views,omitempty"`
}

// CreateTableRequest represents the request to create a table
type CreateTableRequest struct {
	Name        string                 `json:"name" binding:"required"`
	Description *string                `json:"description,omitempty"`
	Icon        *string                `json:"icon,omitempty"`
	Color       *string                `json:"color,omitempty"`
	WorkspaceID *string                `json:"workspace_id,omitempty"`
	Settings    map[string]interface{} `json:"settings,omitempty"`
}

// UpdateTableRequest represents the request to update a table
type UpdateTableRequest struct {
	Name        *string                `json:"name,omitempty"`
	Description *string                `json:"description,omitempty"`
	Icon        *string                `json:"icon,omitempty"`
	Color       *string                `json:"color,omitempty"`
	Settings    map[string]interface{} `json:"settings,omitempty"`
}

// FieldResponse represents a column/field in API responses
type FieldResponse struct {
	ID          string                 `json:"id"`
	TableID     string                 `json:"table_id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Description *string                `json:"description,omitempty"`
	Position    int                    `json:"position"`
	Config      map[string]interface{} `json:"config,omitempty"`
	Required    bool                   `json:"required"`
	Unique      bool                   `json:"unique"`
	Hidden      bool                   `json:"hidden"`
	CreatedAt   string                 `json:"created_at"`
	UpdatedAt   string                 `json:"updated_at"`
	// Options for select fields
	Options []FieldOptionResponse `json:"options,omitempty"`
}

// CreateFieldRequest represents the request to create a field
type CreateFieldRequest struct {
	Name        string                 `json:"name" binding:"required"`
	Type        string                 `json:"type" binding:"required"`
	Description *string                `json:"description,omitempty"`
	Position    *int                   `json:"position,omitempty"`
	Config      map[string]interface{} `json:"config,omitempty"`
	Required    *bool                  `json:"required,omitempty"`
	Unique      *bool                  `json:"unique,omitempty"`
	// Options for select fields (single_select, multi_select)
	Options []CreateFieldOptionRequest `json:"options,omitempty"`
}

// UpdateFieldRequest represents the request to update a field
type UpdateFieldRequest struct {
	Name        *string                `json:"name,omitempty"`
	Description *string                `json:"description,omitempty"`
	Config      map[string]interface{} `json:"config,omitempty"`
	Required    *bool                  `json:"required,omitempty"`
	Hidden      *bool                  `json:"hidden,omitempty"`
}

// FieldOptionResponse represents a select option in API responses
type FieldOptionResponse struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Color    *string `json:"color,omitempty"`
	Position int     `json:"position"`
}

// CreateFieldOptionRequest represents the request to create a field option
type CreateFieldOptionRequest struct {
	Name     string  `json:"name" binding:"required"`
	Color    *string `json:"color,omitempty"`
	Position *int    `json:"position,omitempty"`
}

// RecordResponse represents a row/record in API responses
type RecordResponse struct {
	ID         string                 `json:"id"`
	TableID    string                 `json:"table_id"`
	Data       map[string]interface{} `json:"data"`
	Position   *int                   `json:"position,omitempty"`
	CreatedBy  *string                `json:"created_by,omitempty"`
	ModifiedBy *string                `json:"modified_by,omitempty"`
	CreatedAt  string                 `json:"created_at"`
	UpdatedAt  string                 `json:"updated_at"`
}

// CreateRecordRequest represents the request to create a record
type CreateRecordRequest struct {
	Data map[string]interface{} `json:"data" binding:"required"`
}

// UpdateRecordRequest represents the request to update a record
type UpdateRecordRequest struct {
	Data map[string]interface{} `json:"data" binding:"required"`
}

// ViewResponse represents a view in API responses
type ViewResponse struct {
	ID           string                 `json:"id"`
	TableID      string                 `json:"table_id"`
	Name         string                 `json:"name"`
	Type         string                 `json:"type"`
	Description  *string                `json:"description,omitempty"`
	IsDefault    bool                   `json:"is_default"`
	Config       map[string]interface{} `json:"config,omitempty"`
	Filters      []interface{}          `json:"filters,omitempty"`
	Sorts        []interface{}          `json:"sorts,omitempty"`
	GroupBy      *string                `json:"group_by,omitempty"`
	ViewSettings map[string]interface{} `json:"view_settings,omitempty"`
	Position     int                    `json:"position"`
	CreatedAt    string                 `json:"created_at"`
	UpdatedAt    string                 `json:"updated_at"`
}

// CreateViewRequest represents the request to create a view
type CreateViewRequest struct {
	Name         string                 `json:"name" binding:"required"`
	Type         string                 `json:"type" binding:"required"`
	Description  *string                `json:"description,omitempty"`
	Config       map[string]interface{} `json:"config,omitempty"`
	Filters      []interface{}          `json:"filters,omitempty"`
	Sorts        []interface{}          `json:"sorts,omitempty"`
	GroupBy      *string                `json:"group_by,omitempty"`
	ViewSettings map[string]interface{} `json:"view_settings,omitempty"`
	Position     *int                   `json:"position,omitempty"`
}

// UpdateViewRequest represents the request to update a view
type UpdateViewRequest struct {
	Name         *string                `json:"name,omitempty"`
	Description  *string                `json:"description,omitempty"`
	Config       map[string]interface{} `json:"config,omitempty"`
	Filters      []interface{}          `json:"filters,omitempty"`
	Sorts        []interface{}          `json:"sorts,omitempty"`
	GroupBy      *string                `json:"group_by,omitempty"`
	ViewSettings map[string]interface{} `json:"view_settings,omitempty"`
	IsDefault    *bool                  `json:"is_default,omitempty"`
}

// ============================================================================
// Helper Functions (table-specific to avoid conflicts)
// ============================================================================

// tableParseUUID parses a string UUID into pgtype.UUID
func tableParseUUID(s string) (pgtype.UUID, error) {
	parsed, err := uuid.Parse(s)
	if err != nil {
		return pgtype.UUID{}, err
	}
	return pgtype.UUID{Bytes: parsed, Valid: true}, nil
}

// tableUUIDToString converts a pgtype.UUID to string
func tableUUIDToString(u pgtype.UUID) string {
	if !u.Valid {
		return ""
	}
	return uuid.UUID(u.Bytes).String()
}

// tableUUIDToPtr converts a pgtype.UUID to a string pointer
func tableUUIDToPtr(u pgtype.UUID) *string {
	if !u.Valid {
		return nil
	}
	s := uuid.UUID(u.Bytes).String()
	return &s
}

// tableJSONBytesToMap converts JSON bytes to a map
func tableJSONBytesToMap(data []byte) map[string]interface{} {
	if data == nil {
		return nil
	}
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil
	}
	return result
}

// tableJSONBytesToSlice converts JSON bytes to a slice
func tableJSONBytesToSlice(data []byte) []interface{} {
	if data == nil {
		return nil
	}
	var result []interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil
	}
	return result
}

// tableMapToJSONBytes converts a map to JSON bytes
func tableMapToJSONBytes(m map[string]interface{}) []byte {
	if m == nil {
		return []byte("{}")
	}
	data, err := json.Marshal(m)
	if err != nil {
		return []byte("{}")
	}
	return data
}

// tableSliceToJSONBytes converts a slice to JSON bytes
func tableSliceToJSONBytes(s []interface{}) []byte {
	if s == nil {
		return []byte("[]")
	}
	data, err := json.Marshal(s)
	if err != nil {
		return []byte("[]")
	}
	return data
}

// tableBoolPtr returns a pointer to a bool
func tableBoolPtr(b bool) *bool {
	return &b
}

// tableGetBool returns the value of a bool pointer or default
func tableGetBool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

// ============================================================================
// Table Handlers
// ============================================================================

// ListTables returns all tables for the current user
// GET /api/tables
func (h *Handlers) ListTables(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	ctx := c.Request.Context()
	queries := sqlc.New(h.pool)

	// Optional workspace filter
	var workspaceID pgtype.UUID
	if wsID := c.Query("workspace_id"); wsID != "" {
		parsed, err := tableParseUUID(wsID)
		if err == nil {
			workspaceID = parsed
		}
	}

	tables, err := queries.ListCustomTables(ctx, sqlc.ListCustomTablesParams{
		UserID:      user.ID,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		slog.Error("Failed to list tables", "error", err, "user_id", user.ID, "workspace_id", workspaceID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list tables"})
		return
	}

	// Get row counts for each table
	response := make([]TableResponse, len(tables))
	for i, t := range tables {
		rowCount, err := queries.CountCustomRecords(ctx, t.ID)
		if err != nil {
			slog.Warn("Failed to count records for table", "error", err, "table_id", t.ID)
			rowCount = 0 // Use 0 as fallback
		}
		response[i] = tableToResponse(t, rowCount)
	}

	c.JSON(http.StatusOK, response)
}

// GetTable returns a single table with its columns and views
// GET /api/tables/:id
func (h *Handlers) GetTable(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	tableID, err := tableParseUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid table ID"})
		return
	}

	ctx := c.Request.Context()
	queries := sqlc.New(h.pool)

	table, err := queries.GetCustomTable(ctx, sqlc.GetCustomTableParams{
		ID:     tableID,
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Table not found"})
		return
	}

	// Get columns, views, and row count
	fields, _ := queries.ListCustomFields(ctx, tableID)
	views, _ := queries.ListCustomViews(ctx, tableID)
	rowCount, _ := queries.CountCustomRecords(ctx, tableID)

	// Get options for select fields
	fieldResponses := make([]FieldResponse, len(fields))
	for i, f := range fields {
		fieldResponses[i] = fieldToResponse(f)
		// Load options for select fields
		if f.FieldType == "single_select" || f.FieldType == "multi_select" {
			options, _ := queries.ListFieldOptions(ctx, f.ID)
			fieldResponses[i].Options = make([]FieldOptionResponse, len(options))
			for j, o := range options {
				fieldResponses[i].Options[j] = optionToResponse(o)
			}
		}
	}

	viewResponses := make([]ViewResponse, len(views))
	for i, v := range views {
		viewResponses[i] = viewToResponse(v)
	}

	response := tableToResponse(table, rowCount)
	response.Columns = fieldResponses
	response.Views = viewResponses

	c.JSON(http.StatusOK, response)
}

// CreateTable creates a new table
// POST /api/tables
func (h *Handlers) CreateTable(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var req CreateTableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	ctx := c.Request.Context()
	queries := sqlc.New(h.pool)

	// Parse workspace ID if provided
	var workspaceID pgtype.UUID
	if req.WorkspaceID != nil {
		parsed, err := tableParseUUID(*req.WorkspaceID)
		if err == nil {
			workspaceID = parsed
		}
	}

	table, err := queries.CreateCustomTable(ctx, sqlc.CreateCustomTableParams{
		UserID:      user.ID,
		Name:        req.Name,
		Description: req.Description,
		Icon:        req.Icon,
		Color:       req.Color,
		WorkspaceID: workspaceID,
		Settings:    tableMapToJSONBytes(req.Settings),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create table"})
		return
	}

	// Create default view (Grid)
	_, err = queries.CreateCustomView(ctx, sqlc.CreateCustomViewParams{
		TableID:      table.ID,
		Name:         "Grid View",
		ViewType:     "grid",
		Config:       []byte("{}"),
		Filters:      []byte("[]"),
		Sorts:        []byte("[]"),
		ViewSettings: []byte("{}"),
		Position:     0,
	})
	if err != nil {
		// Log error but don't fail - table was created successfully
	}

	// Create default primary field (Name)
	_, err = queries.CreateCustomField(ctx, sqlc.CreateCustomFieldParams{
		TableID:      table.ID,
		Name:         "Name",
		FieldType:    "text",
		Position:     0,
		Config:       []byte(`{"is_primary": true}`),
		Required:     tableBoolPtr(true),
		UniqueValues: tableBoolPtr(false),
	})
	if err != nil {
		// Log error but don't fail
	}

	c.JSON(http.StatusCreated, tableToResponse(table, 0))
}

// UpdateTable updates an existing table
// PUT /api/tables/:id
func (h *Handlers) UpdateTable(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	tableID, err := tableParseUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid table ID"})
		return
	}

	var req UpdateTableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	ctx := c.Request.Context()
	queries := sqlc.New(h.pool)

	// Get existing table
	existing, err := queries.GetCustomTable(ctx, sqlc.GetCustomTableParams{
		ID:     tableID,
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Table not found"})
		return
	}

	// Merge updates
	name := existing.Name
	if req.Name != nil {
		name = *req.Name
	}
	description := existing.Description
	if req.Description != nil {
		description = req.Description
	}
	icon := existing.Icon
	if req.Icon != nil {
		icon = req.Icon
	}
	color := existing.Color
	if req.Color != nil {
		color = req.Color
	}
	settings := existing.Settings
	if req.Settings != nil {
		settings = tableMapToJSONBytes(req.Settings)
	}

	table, err := queries.UpdateCustomTable(ctx, sqlc.UpdateCustomTableParams{
		ID:          tableID,
		Name:        name,
		Description: description,
		Icon:        icon,
		Color:       color,
		Settings:    settings,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update table"})
		return
	}

	rowCount, _ := queries.CountCustomRecords(ctx, tableID)
	c.JSON(http.StatusOK, tableToResponse(table, rowCount))
}

// DeleteTable deletes a table and all its data
// DELETE /api/tables/:id
func (h *Handlers) DeleteTable(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	tableID, err := tableParseUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid table ID"})
		return
	}

	ctx := c.Request.Context()
	queries := sqlc.New(h.pool)

	err = queries.DeleteCustomTable(ctx, sqlc.DeleteCustomTableParams{
		ID:     tableID,
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete table"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ============================================================================
// Field (Column) Handlers
// ============================================================================

// ListFields returns all fields for a table
// GET /api/tables/:id/fields
func (h *Handlers) ListFields(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	tableID, err := tableParseUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid table ID"})
		return
	}

	ctx := c.Request.Context()
	queries := sqlc.New(h.pool)

	// Verify user owns table
	_, err = queries.GetCustomTable(ctx, sqlc.GetCustomTableParams{
		ID:     tableID,
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Table not found"})
		return
	}

	fields, err := queries.ListCustomFields(ctx, tableID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list fields"})
		return
	}

	response := make([]FieldResponse, len(fields))
	for i, f := range fields {
		response[i] = fieldToResponse(f)
		// Load options for select fields
		if f.FieldType == "single_select" || f.FieldType == "multi_select" {
			options, _ := queries.ListFieldOptions(ctx, f.ID)
			response[i].Options = make([]FieldOptionResponse, len(options))
			for j, o := range options {
				response[i].Options[j] = optionToResponse(o)
			}
		}
	}

	c.JSON(http.StatusOK, response)
}

// CreateField creates a new field in a table
// POST /api/tables/:id/fields
func (h *Handlers) CreateField(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	tableID, err := tableParseUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid table ID"})
		return
	}

	var req CreateFieldRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	ctx := c.Request.Context()
	queries := sqlc.New(h.pool)

	// Verify user owns table
	_, err = queries.GetCustomTable(ctx, sqlc.GetCustomTableParams{
		ID:     tableID,
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Table not found"})
		return
	}

	// Get next position
	position := int32(0)
	if req.Position != nil {
		position = int32(*req.Position)
	} else {
		existingFields, _ := queries.ListCustomFields(ctx, tableID)
		position = int32(len(existingFields))
	}

	field, err := queries.CreateCustomField(ctx, sqlc.CreateCustomFieldParams{
		TableID:      tableID,
		Name:         req.Name,
		FieldType:    req.Type,
		Description:  req.Description,
		Position:     position,
		Config:       tableMapToJSONBytes(req.Config),
		Required:     req.Required,
		UniqueValues: req.Unique,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create field"})
		return
	}

	// Create options for select fields
	response := fieldToResponse(field)
	if (req.Type == "single_select" || req.Type == "multi_select") && len(req.Options) > 0 {
		response.Options = make([]FieldOptionResponse, len(req.Options))
		for i, opt := range req.Options {
			pos := int32(i)
			if opt.Position != nil {
				pos = int32(*opt.Position)
			}
			option, err := queries.CreateFieldOption(ctx, sqlc.CreateFieldOptionParams{
				FieldID:  field.ID,
				Name:     opt.Name,
				Color:    opt.Color,
				Position: pos,
			})
			if err == nil {
				response.Options[i] = optionToResponse(option)
			}
		}
	}

	c.JSON(http.StatusCreated, response)
}

// UpdateField updates an existing field
// PUT /api/tables/:tableId/fields/:fieldId
func (h *Handlers) UpdateField(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	tableID, err := tableParseUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid table ID"})
		return
	}

	fieldID, err := tableParseUUID(c.Param("columnId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid column ID"})
		return
	}

	var req UpdateFieldRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	ctx := c.Request.Context()
	queries := sqlc.New(h.pool)

	// Verify user owns table
	_, err = queries.GetCustomTable(ctx, sqlc.GetCustomTableParams{
		ID:     tableID,
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Table not found"})
		return
	}

	// Get existing field
	existing, err := queries.GetCustomField(ctx, fieldID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Field not found"})
		return
	}

	// Verify field belongs to table
	if tableUUIDToString(existing.TableID) != tableUUIDToString(tableID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Field does not belong to this table"})
		return
	}

	// Merge updates
	name := existing.Name
	if req.Name != nil {
		name = *req.Name
	}
	description := existing.Description
	if req.Description != nil {
		description = req.Description
	}
	config := existing.Config
	if req.Config != nil {
		config = tableMapToJSONBytes(req.Config)
	}
	required := existing.Required
	if req.Required != nil {
		required = req.Required
	}
	hidden := existing.Hidden
	if req.Hidden != nil {
		hidden = req.Hidden
	}

	field, err := queries.UpdateCustomField(ctx, sqlc.UpdateCustomFieldParams{
		ID:          fieldID,
		Name:        name,
		Description: description,
		Config:      config,
		Required:    required,
		Hidden:      hidden,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update field"})
		return
	}

	c.JSON(http.StatusOK, fieldToResponse(field))
}

// DeleteField deletes a field from a table
// DELETE /api/tables/:tableId/fields/:fieldId
func (h *Handlers) DeleteField(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	tableID, err := tableParseUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid table ID"})
		return
	}

	fieldID, err := tableParseUUID(c.Param("columnId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid column ID"})
		return
	}

	ctx := c.Request.Context()
	queries := sqlc.New(h.pool)

	// Verify user owns table
	_, err = queries.GetCustomTable(ctx, sqlc.GetCustomTableParams{
		ID:     tableID,
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Table not found"})
		return
	}

	// Verify field belongs to table
	existing, err := queries.GetCustomField(ctx, fieldID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Field not found"})
		return
	}
	if tableUUIDToString(existing.TableID) != tableUUIDToString(tableID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Field does not belong to this table"})
		return
	}

	err = queries.DeleteCustomField(ctx, fieldID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete field"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ReorderFields reorders fields in a table
// POST /api/tables/:id/fields/reorder
func (h *Handlers) ReorderFields(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	tableID, err := tableParseUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid table ID"})
		return
	}

	var req struct {
		ColumnIDs []string `json:"column_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	ctx := c.Request.Context()
	queries := sqlc.New(h.pool)

	// Verify user owns table
	_, err = queries.GetCustomTable(ctx, sqlc.GetCustomTableParams{
		ID:     tableID,
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Table not found"})
		return
	}

	// Update positions
	for i, fid := range req.ColumnIDs {
		fieldID, err := tableParseUUID(fid)
		if err != nil {
			continue
		}
		queries.UpdateCustomFieldPosition(ctx, sqlc.UpdateCustomFieldPositionParams{
			ID:       fieldID,
			Position: int32(i),
		})
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ============================================================================
// Record (Row) Handlers
// ============================================================================

// ListRecords returns all records for a table
// GET /api/tables/:id/rows
func (h *Handlers) ListRecords(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	tableID, err := tableParseUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid table ID"})
		return
	}

	ctx := c.Request.Context()
	queries := sqlc.New(h.pool)

	// Verify user owns table
	_, err = queries.GetCustomTable(ctx, sqlc.GetCustomTableParams{
		ID:     tableID,
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Table not found"})
		return
	}

	// Parse pagination params (frontend uses page/page_size, convert to limit/offset)
	page := 1
	pageSize := 100
	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if ps := c.Query("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 1000 {
			pageSize = parsed
		}
	}
	// Also support limit/offset for backwards compatibility
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 1000 {
			pageSize = parsed
		}
	}
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			page = (parsed / pageSize) + 1
		}
	}

	limit := int32(pageSize)
	offset := int32((page - 1) * pageSize)

	records, err := queries.ListCustomRecords(ctx, sqlc.ListCustomRecordsParams{
		TableID:   tableID,
		LimitVal:  limit,
		OffsetVal: offset,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list rows"})
		return
	}

	response := make([]RecordResponse, len(records))
	for i, r := range records {
		response[i] = recordToResponse(r)
	}

	// Get total count for pagination
	total, _ := queries.CountCustomRecords(ctx, tableID)
	hasMore := int64(page*pageSize) < total

	c.JSON(http.StatusOK, gin.H{
		"rows":      response,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
		"has_more":  hasMore,
	})
}

// GetRecord returns a single record
// GET /api/tables/:id/records/:recordId
func (h *Handlers) GetRecord(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	tableID, err := tableParseUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid table ID"})
		return
	}

	recordID, err := tableParseUUID(c.Param("rowId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid row ID"})
		return
	}

	ctx := c.Request.Context()
	queries := sqlc.New(h.pool)

	// Verify user owns table
	_, err = queries.GetCustomTable(ctx, sqlc.GetCustomTableParams{
		ID:     tableID,
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Table not found"})
		return
	}

	record, err := queries.GetCustomRecord(ctx, recordID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Record not found"})
		return
	}

	// Verify record belongs to table
	if tableUUIDToString(record.TableID) != tableUUIDToString(tableID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Record does not belong to this table"})
		return
	}

	c.JSON(http.StatusOK, recordToResponse(record))
}

// CreateRecord creates a new record in a table
// POST /api/tables/:id/records
func (h *Handlers) CreateRecord(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	tableID, err := tableParseUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid table ID"})
		return
	}

	var req CreateRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	ctx := c.Request.Context()
	queries := sqlc.New(h.pool)

	// Verify user owns table
	_, err = queries.GetCustomTable(ctx, sqlc.GetCustomTableParams{
		ID:     tableID,
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Table not found"})
		return
	}

	record, err := queries.CreateCustomRecord(ctx, sqlc.CreateCustomRecordParams{
		TableID:   tableID,
		Data:      tableMapToJSONBytes(req.Data),
		CreatedBy: &user.ID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create record"})
		return
	}

	c.JSON(http.StatusCreated, recordToResponse(record))
}

// UpdateRecord updates an existing record
// PUT /api/tables/:id/records/:recordId
func (h *Handlers) UpdateRecord(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	tableID, err := tableParseUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid table ID"})
		return
	}

	recordID, err := tableParseUUID(c.Param("rowId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid row ID"})
		return
	}

	var req UpdateRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	ctx := c.Request.Context()
	queries := sqlc.New(h.pool)

	// Verify user owns table
	_, err = queries.GetCustomTable(ctx, sqlc.GetCustomTableParams{
		ID:     tableID,
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Table not found"})
		return
	}

	// Verify record belongs to table
	existing, err := queries.GetCustomRecord(ctx, recordID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Record not found"})
		return
	}
	if tableUUIDToString(existing.TableID) != tableUUIDToString(tableID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Record does not belong to this table"})
		return
	}

	record, err := queries.UpdateCustomRecord(ctx, sqlc.UpdateCustomRecordParams{
		ID:         recordID,
		Data:       tableMapToJSONBytes(req.Data),
		ModifiedBy: &user.ID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update record"})
		return
	}

	c.JSON(http.StatusOK, recordToResponse(record))
}

// UpdateRecordField updates a single field in a record
// PATCH /api/tables/:id/records/:recordId/fields/:fieldId
func (h *Handlers) UpdateRecordField(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	tableID, err := tableParseUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid table ID"})
		return
	}

	recordID, err := tableParseUUID(c.Param("rowId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid row ID"})
		return
	}

	fieldID := c.Param("fieldId")

	var req struct {
		Value interface{} `json:"value"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	ctx := c.Request.Context()
	queries := sqlc.New(h.pool)

	// Verify user owns table
	_, err = queries.GetCustomTable(ctx, sqlc.GetCustomTableParams{
		ID:     tableID,
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Table not found"})
		return
	}

	// Verify record belongs to table
	existing, err := queries.GetCustomRecord(ctx, recordID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Record not found"})
		return
	}
	if tableUUIDToString(existing.TableID) != tableUUIDToString(tableID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Record does not belong to this table"})
		return
	}

	// Marshal value to JSON
	valueJSON, err := json.Marshal(req.Value)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid value"})
		return
	}

	record, err := queries.UpdateCustomRecordField(ctx, sqlc.UpdateCustomRecordFieldParams{
		ID:         recordID,
		Column2:    fieldID,
		Column3:    valueJSON,
		ModifiedBy: &user.ID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update record field"})
		return
	}

	c.JSON(http.StatusOK, recordToResponse(record))
}

// DeleteRecord deletes a record
// DELETE /api/tables/:id/records/:recordId
func (h *Handlers) DeleteRecord(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	tableID, err := tableParseUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid table ID"})
		return
	}

	recordID, err := tableParseUUID(c.Param("rowId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid row ID"})
		return
	}

	ctx := c.Request.Context()
	queries := sqlc.New(h.pool)

	// Verify user owns table
	_, err = queries.GetCustomTable(ctx, sqlc.GetCustomTableParams{
		ID:     tableID,
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Table not found"})
		return
	}

	// Verify record belongs to table
	existing, err := queries.GetCustomRecord(ctx, recordID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Record not found"})
		return
	}
	if tableUUIDToString(existing.TableID) != tableUUIDToString(tableID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Record does not belong to this table"})
		return
	}

	err = queries.DeleteCustomRecord(ctx, recordID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete record"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// BulkDeleteRecords deletes multiple records
// POST /api/tables/:id/records/bulk-delete
func (h *Handlers) BulkDeleteRecords(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	tableID, err := tableParseUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid table ID"})
		return
	}

	var req struct {
		RowIDs []string `json:"row_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	ctx := c.Request.Context()
	queries := sqlc.New(h.pool)

	// Verify user owns table
	_, err = queries.GetCustomTable(ctx, sqlc.GetCustomTableParams{
		ID:     tableID,
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Table not found"})
		return
	}

	// Delete each row
	deleted := 0
	for _, rid := range req.RowIDs {
		recordID, err := tableParseUUID(rid)
		if err != nil {
			continue
		}
		// Verify record belongs to table before deleting
		record, err := queries.GetCustomRecord(ctx, recordID)
		if err != nil {
			continue
		}
		if tableUUIDToString(record.TableID) != tableUUIDToString(tableID) {
			continue
		}
		if err := queries.DeleteCustomRecord(ctx, recordID); err == nil {
			deleted++
		}
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "deleted": deleted})
}

// ============================================================================
// View Handlers
// ============================================================================

// ListViews returns all views for a table
// GET /api/tables/:id/views
func (h *Handlers) ListViews(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	tableID, err := tableParseUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid table ID"})
		return
	}

	ctx := c.Request.Context()
	queries := sqlc.New(h.pool)

	// Verify user owns table
	_, err = queries.GetCustomTable(ctx, sqlc.GetCustomTableParams{
		ID:     tableID,
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Table not found"})
		return
	}

	views, err := queries.ListCustomViews(ctx, tableID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list views"})
		return
	}

	response := make([]ViewResponse, len(views))
	for i, v := range views {
		response[i] = viewToResponse(v)
	}

	c.JSON(http.StatusOK, response)
}

// CreateView creates a new view for a table
// POST /api/tables/:id/views
func (h *Handlers) CreateView(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	tableID, err := tableParseUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid table ID"})
		return
	}

	var req CreateViewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	ctx := c.Request.Context()
	queries := sqlc.New(h.pool)

	// Verify user owns table
	_, err = queries.GetCustomTable(ctx, sqlc.GetCustomTableParams{
		ID:     tableID,
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Table not found"})
		return
	}

	// Get next position
	position := int32(0)
	if req.Position != nil {
		position = int32(*req.Position)
	} else {
		existingViews, _ := queries.ListCustomViews(ctx, tableID)
		position = int32(len(existingViews))
	}

	// Parse group_by if provided
	var groupBy pgtype.UUID
	if req.GroupBy != nil {
		parsed, err := tableParseUUID(*req.GroupBy)
		if err == nil {
			groupBy = parsed
		}
	}

	view, err := queries.CreateCustomView(ctx, sqlc.CreateCustomViewParams{
		TableID:      tableID,
		Name:         req.Name,
		ViewType:     req.Type,
		Description:  req.Description,
		Config:       tableMapToJSONBytes(req.Config),
		Filters:      tableSliceToJSONBytes(req.Filters),
		Sorts:        tableSliceToJSONBytes(req.Sorts),
		GroupBy:      groupBy,
		ViewSettings: tableMapToJSONBytes(req.ViewSettings),
		Position:     position,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create view"})
		return
	}

	c.JSON(http.StatusCreated, viewToResponse(view))
}

// UpdateView updates an existing view
// PUT /api/tables/:tableId/views/:viewId
func (h *Handlers) UpdateView(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	tableID, err := tableParseUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid table ID"})
		return
	}

	viewID, err := tableParseUUID(c.Param("viewId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid view ID"})
		return
	}

	var req UpdateViewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	ctx := c.Request.Context()
	queries := sqlc.New(h.pool)

	// Verify user owns table
	_, err = queries.GetCustomTable(ctx, sqlc.GetCustomTableParams{
		ID:     tableID,
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Table not found"})
		return
	}

	// Get existing view
	existing, err := queries.GetCustomView(ctx, viewID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "View not found"})
		return
	}

	// Verify view belongs to table
	if tableUUIDToString(existing.TableID) != tableUUIDToString(tableID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "View does not belong to this table"})
		return
	}

	// Merge updates
	name := existing.Name
	if req.Name != nil {
		name = *req.Name
	}
	description := existing.Description
	if req.Description != nil {
		description = req.Description
	}
	config := existing.Config
	if req.Config != nil {
		config = tableMapToJSONBytes(req.Config)
	}
	filters := existing.Filters
	if req.Filters != nil {
		filters = tableSliceToJSONBytes(req.Filters)
	}
	sorts := existing.Sorts
	if req.Sorts != nil {
		sorts = tableSliceToJSONBytes(req.Sorts)
	}
	groupBy := existing.GroupBy
	if req.GroupBy != nil {
		parsed, err := tableParseUUID(*req.GroupBy)
		if err == nil {
			groupBy = parsed
		}
	}
	viewSettings := existing.ViewSettings
	if req.ViewSettings != nil {
		viewSettings = tableMapToJSONBytes(req.ViewSettings)
	}

	// Handle is_default separately
	if req.IsDefault != nil && *req.IsDefault {
		queries.SetDefaultView(ctx, sqlc.SetDefaultViewParams{
			TableID: tableID,
			ID:      viewID,
		})
	}

	view, err := queries.UpdateCustomView(ctx, sqlc.UpdateCustomViewParams{
		ID:           viewID,
		Name:         name,
		Description:  description,
		Config:       config,
		Filters:      filters,
		Sorts:        sorts,
		GroupBy:      groupBy,
		ViewSettings: viewSettings,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update view"})
		return
	}

	c.JSON(http.StatusOK, viewToResponse(view))
}

// DeleteView deletes a view
// DELETE /api/tables/:tableId/views/:viewId
func (h *Handlers) DeleteView(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	tableID, err := tableParseUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid table ID"})
		return
	}

	viewID, err := tableParseUUID(c.Param("viewId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid view ID"})
		return
	}

	ctx := c.Request.Context()
	queries := sqlc.New(h.pool)

	// Verify user owns table
	_, err = queries.GetCustomTable(ctx, sqlc.GetCustomTableParams{
		ID:     tableID,
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Table not found"})
		return
	}

	// Verify view belongs to table
	existing, err := queries.GetCustomView(ctx, viewID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "View not found"})
		return
	}
	if tableUUIDToString(existing.TableID) != tableUUIDToString(tableID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "View does not belong to this table"})
		return
	}

	// Don't allow deleting the last view
	views, _ := queries.ListCustomViews(ctx, tableID)
	if len(views) <= 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete the last view"})
		return
	}

	err = queries.DeleteCustomView(ctx, viewID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete view"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ============================================================================
// Response Transformers
// ============================================================================

func tableToResponse(t sqlc.CustomTable, rowCount int64) TableResponse {
	return TableResponse{
		ID:          tableUUIDToString(t.ID),
		Name:        t.Name,
		Description: t.Description,
		Icon:        t.Icon,
		Color:       t.Color,
		WorkspaceID: tableUUIDToPtr(t.WorkspaceID),
		Settings:    tableJSONBytesToMap(t.Settings),
		RowCount:    rowCount,
		CreatedAt:   t.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   t.UpdatedAt.Time.Format("2006-01-02T15:04:05Z"),
	}
}

func fieldToResponse(f sqlc.CustomField) FieldResponse {
	fieldType := ""
	if ft, ok := f.FieldType.(string); ok {
		fieldType = ft
	}
	return FieldResponse{
		ID:          tableUUIDToString(f.ID),
		TableID:     tableUUIDToString(f.TableID),
		Name:        f.Name,
		Type:        fieldType,
		Description: f.Description,
		Position:    int(f.Position),
		Config:      tableJSONBytesToMap(f.Config),
		Required:    tableGetBool(f.Required),
		Unique:      tableGetBool(f.UniqueValues),
		Hidden:      tableGetBool(f.Hidden),
		CreatedAt:   f.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   f.UpdatedAt.Time.Format("2006-01-02T15:04:05Z"),
	}
}

func optionToResponse(o sqlc.CustomFieldOption) FieldOptionResponse {
	return FieldOptionResponse{
		ID:       tableUUIDToString(o.ID),
		Name:     o.Name,
		Color:    o.Color,
		Position: int(o.Position),
	}
}

func recordToResponse(r sqlc.CustomRecord) RecordResponse {
	var position *int
	if r.Position != nil {
		p := int(*r.Position)
		position = &p
	}
	return RecordResponse{
		ID:         tableUUIDToString(r.ID),
		TableID:    tableUUIDToString(r.TableID),
		Data:       tableJSONBytesToMap(r.Data),
		Position:   position,
		CreatedBy:  r.CreatedBy,
		ModifiedBy: r.ModifiedBy,
		CreatedAt:  r.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:  r.UpdatedAt.Time.Format("2006-01-02T15:04:05Z"),
	}
}

func viewToResponse(v sqlc.CustomView) ViewResponse {
	viewType := ""
	if vt, ok := v.ViewType.(string); ok {
		viewType = vt
	}
	return ViewResponse{
		ID:           tableUUIDToString(v.ID),
		TableID:      tableUUIDToString(v.TableID),
		Name:         v.Name,
		Type:         viewType,
		Description:  v.Description,
		IsDefault:    tableGetBool(v.IsDefault),
		Config:       tableJSONBytesToMap(v.Config),
		Filters:      tableJSONBytesToSlice(v.Filters),
		Sorts:        tableJSONBytesToSlice(v.Sorts),
		GroupBy:      tableUUIDToPtr(v.GroupBy),
		ViewSettings: tableJSONBytesToMap(v.ViewSettings),
		Position:     int(v.Position),
		CreatedAt:    v.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:    v.UpdatedAt.Time.Format("2006-01-02T15:04:05Z"),
	}
}
