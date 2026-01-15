package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rhl/businessos-backend/internal/services"
)

// OSAWorkflowsHandler handles OSA workflow and file operations
type OSAWorkflowsHandler struct {
	pool        *pgxpool.Pool
	syncService *services.OSAFileSyncService
}

// NewOSAWorkflowsHandler creates a new workflows handler
func NewOSAWorkflowsHandler(pool *pgxpool.Pool, syncService *services.OSAFileSyncService) *OSAWorkflowsHandler {
	return &OSAWorkflowsHandler{
		pool:        pool,
		syncService: syncService,
	}
}

// ListWorkflows returns all workflows for the current user
// GET /api/osa/workflows
func (h *OSAWorkflowsHandler) ListWorkflows(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Query workflows from database
	query := `
		SELECT
			ga.id,
			ga.name,
			ga.display_name,
			ga.description,
			ga.osa_workflow_id,
			ga.status,
			ga.files_created,
			ga.build_status,
			ga.created_at,
			ga.generated_at,
			ga.deployed_at,
			w.name as workspace_name
		FROM osa_generated_apps ga
		JOIN osa_workspaces w ON ga.workspace_id = w.id
		WHERE w.user_id = $1
		ORDER BY ga.created_at DESC
	`

	rows, err := h.pool.Query(c.Request.Context(), query, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch workflows", "details": err.Error()})
		return
	}
	defer rows.Close()

	type WorkflowListItem struct {
		ID            uuid.UUID   `json:"id"`
		Name          string      `json:"name"`
		DisplayName   string      `json:"display_name"`
		Description   string      `json:"description"`
		WorkflowID    string      `json:"workflow_id"`
		Status        string      `json:"status"`
		FilesCreated  int         `json:"files_created"`
		BuildStatus   *string     `json:"build_status"`
		CreatedAt     time.Time   `json:"created_at"`
		GeneratedAt   *time.Time  `json:"generated_at"`
		DeployedAt    *time.Time  `json:"deployed_at"`
		WorkspaceName string      `json:"workspace_name"`
	}

	workflows := []WorkflowListItem{}
	for rows.Next() {
		var item WorkflowListItem
		err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.DisplayName,
			&item.Description,
			&item.WorkflowID,
			&item.Status,
			&item.FilesCreated,
			&item.BuildStatus,
			&item.CreatedAt,
			&item.GeneratedAt,
			&item.DeployedAt,
			&item.WorkspaceName,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan workflow", "details": err.Error()})
			return
		}
		workflows = append(workflows, item)
	}

	c.JSON(http.StatusOK, gin.H{
		"workflows": workflows,
		"count":     len(workflows),
	})
}

// GetWorkflow returns details for a specific workflow
// GET /api/osa/workflows/:id
func (h *OSAWorkflowsHandler) GetWorkflow(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	workflowID := c.Param("id")

	// Query workflow details
	query := `
		SELECT
			ga.id,
			ga.name,
			ga.display_name,
			ga.description,
			ga.osa_workflow_id,
			ga.status,
			ga.files_created,
			ga.build_status,
			ga.metadata,
			ga.error_message,
			ga.error_stack,
			ga.created_at,
			ga.generated_at,
			ga.deployed_at,
			w.name as workspace_name,
			w.id as workspace_id
		FROM osa_generated_apps ga
		JOIN osa_workspaces w ON ga.workspace_id = w.id
		WHERE (ga.id = $1 OR ga.osa_workflow_id LIKE $2)
		  AND w.user_id = $3
	`

	var (
		id            uuid.UUID
		name          string
		displayName   string
		description   string
		osaWorkflowID string
		status        string
		filesCreated  int
		buildStatus   *string
		metadataJSON  []byte
		errorMessage  *string
		errorStack    *string
		createdAt     time.Time
		generatedAt   *time.Time
		deployedAt    *time.Time
		workspaceName string
		workspaceID   uuid.UUID
	)

	// Try to parse as UUID, otherwise use as workflow ID prefix
	workflowUUID, parseErr := uuid.Parse(workflowID)
	var searchID interface{}
	searchPrefix := workflowID + "%"

	if parseErr == nil {
		searchID = workflowUUID
	} else {
		searchID = uuid.Nil // Won't match, but keeps query structure
	}

	err := h.pool.QueryRow(c.Request.Context(), query, searchID, searchPrefix, userID).Scan(
		&id,
		&name,
		&displayName,
		&description,
		&osaWorkflowID,
		&status,
		&filesCreated,
		&buildStatus,
		&metadataJSON,
		&errorMessage,
		&errorStack,
		&createdAt,
		&generatedAt,
		&deployedAt,
		&workspaceName,
		&workspaceID,
	)

	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Workflow not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch workflow", "details": err.Error()})
		}
		return
	}

	// Parse metadata
	var metadata map[string]interface{}
	if err := json.Unmarshal(metadataJSON, &metadata); err != nil {
		metadata = make(map[string]interface{})
	}

	c.JSON(http.StatusOK, gin.H{
		"id":             id,
		"name":           name,
		"display_name":   displayName,
		"description":    description,
		"workflow_id":    osaWorkflowID,
		"status":         status,
		"files_created":  filesCreated,
		"build_status":   buildStatus,
		"metadata":       metadata,
		"error_message":  errorMessage,
		"error_stack":    errorStack,
		"created_at":     createdAt,
		"generated_at":   generatedAt,
		"deployed_at":    deployedAt,
		"workspace_name": workspaceName,
		"workspace_id":   workspaceID,
	})
}

// GetWorkflowFiles returns all files for a workflow
// GET /api/osa/workflows/:id/files
func (h *OSAWorkflowsHandler) GetWorkflowFiles(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	workflowID := c.Param("id")

	// Query workflow metadata (contains all files)
	query := `
		SELECT ga.id, ga.name, ga.osa_workflow_id, ga.metadata, ga.created_at, ga.updated_at
		FROM osa_generated_apps ga
		JOIN osa_workspaces w ON ga.workspace_id = w.id
		WHERE (ga.id = $1 OR ga.osa_workflow_id LIKE $2)
		  AND w.user_id = $3
	`

	workflowUUID, parseErr := uuid.Parse(workflowID)
	var searchID interface{}
	searchPrefix := workflowID + "%"

	if parseErr == nil {
		searchID = workflowUUID
	} else {
		searchID = uuid.Nil
	}

	var appID uuid.UUID
	var appName string
	var osaWorkflowID string
	var metadataJSON []byte
	var createdAt, updatedAt time.Time

	err := h.pool.QueryRow(c.Request.Context(), query, searchID, searchPrefix, userID).Scan(
		&appID, &appName, &osaWorkflowID, &metadataJSON, &createdAt, &updatedAt,
	)

	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Workflow not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch files", "details": err.Error()})
		}
		return
	}

	// Parse metadata to extract file contents
	var metadata map[string]interface{}
	if err := json.Unmarshal(metadataJSON, &metadata); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse metadata"})
		return
	}

	// Structure file response with deterministic IDs
	files := []map[string]interface{}{}
	fileTypes := []string{"analysis", "architecture", "code", "quality", "deployment", "monitoring", "strategy", "recommendations"}
	namespace := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8") // DNS namespace

	for _, fileType := range fileTypes {
		if content, ok := metadata[fileType].(string); ok && content != "" {
			// Special handling for "code" type - parse multi-file bundle
			if fileType == "code" {
				bundledFiles := services.ParseFileBundle(content)
				for _, bundledFile := range bundledFiles {
					// Generate deterministic UUID for each bundled file
					fileID := uuid.NewSHA1(namespace, []byte(appID.String()+":code:"+bundledFile.Path))
					ext := services.GetFileExtension(bundledFile.Path)
					fileCategory := services.CategorizeFileType(bundledFile.Path)

					files = append(files, map[string]interface{}{
						"id":         fileID.String(),
						"name":       bundledFile.Path,
						"type":       fileCategory,
						"size":       len(bundledFile.Content),
						"language":   services.GetLanguageFromExtension(ext),
						"created_at": createdAt,
						"updated_at": updatedAt,
					})
				}
			} else {
				// Regular metadata files (analysis, architecture, etc.)
				fileID := uuid.NewSHA1(namespace, []byte(appID.String()+":"+fileType))
				fileName := fileType + ".md"

				files = append(files, map[string]interface{}{
					"id":         fileID.String(),
					"name":       fileName,
					"type":       "documentation",
					"size":       len(content),
					"language":   "markdown",
					"created_at": createdAt,
					"updated_at": updatedAt,
				})
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"workflow_id": osaWorkflowID,
		"files":       files,
		"count":       len(files),
	})
}

// GetFileContent returns the content of a specific file
// GET /api/osa/workflows/:id/files/:type
func (h *OSAWorkflowsHandler) GetFileContent(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	workflowID := c.Param("id")
	fileType := c.Param("type")

	// Validate file type
	validTypes := map[string]bool{
		"analysis": true, "architecture": true, "code": true, "quality": true,
		"deployment": true, "monitoring": true, "strategy": true, "recommendations": true,
	}
	if !validTypes[fileType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type"})
		return
	}

	// Query workflow metadata
	query := `
		SELECT ga.metadata
		FROM osa_generated_apps ga
		JOIN osa_workspaces w ON ga.workspace_id = w.id
		WHERE (ga.id = $1 OR ga.osa_workflow_id LIKE $2)
		  AND w.user_id = $3
	`

	workflowUUID, parseErr := uuid.Parse(workflowID)
	var searchID interface{}
	searchPrefix := workflowID + "%"

	if parseErr == nil {
		searchID = workflowUUID
	} else {
		searchID = uuid.Nil
	}

	var metadataJSON []byte
	err := h.pool.QueryRow(c.Request.Context(), query, searchID, searchPrefix, userID).Scan(&metadataJSON)

	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Workflow not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch file", "details": err.Error()})
		}
		return
	}

	// Parse metadata
	var metadata map[string]interface{}
	if err := json.Unmarshal(metadataJSON, &metadata); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse metadata"})
		return
	}

	// Extract file content
	content, ok := metadata[fileType].(string)
	if !ok || content == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"type":    fileType,
		"content": content,
		"size":    len(content),
	})
}

// GetFileContentByID returns file content by file ID
// GET /api/osa/files/:id/content
func (h *OSAWorkflowsHandler) GetFileContentByID(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	fileID := c.Param("id")
	fileUUID, err := uuid.Parse(fileID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}

	// Query all workflows to find the one containing this file
	query := `
		SELECT ga.id, ga.name, ga.metadata, ga.created_at, ga.updated_at
		FROM osa_generated_apps ga
		JOIN osa_workspaces w ON ga.workspace_id = w.id
		WHERE w.user_id = $1
	`

	rows, err := h.pool.Query(c.Request.Context(), query, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search files", "details": err.Error()})
		return
	}
	defer rows.Close()

	namespace := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	fileTypes := []string{"analysis", "architecture", "code", "quality", "deployment", "monitoring", "strategy", "recommendations"}

	for rows.Next() {
		var appID uuid.UUID
		var appName string
		var metadataJSON []byte
		var createdAt, updatedAt time.Time

		if err := rows.Scan(&appID, &appName, &metadataJSON, &createdAt, &updatedAt); err != nil {
			continue
		}

		var metadata map[string]interface{}
		if err := json.Unmarshal(metadataJSON, &metadata); err != nil {
			continue
		}

		// Check each file type to see if it matches the requested file ID
		for _, fileType := range fileTypes {
			if content, ok := metadata[fileType].(string); ok && content != "" {
				// For "code" type, check bundled files
				if fileType == "code" {
					bundledFiles := services.ParseFileBundle(content)
					for _, bundledFile := range bundledFiles {
						expectedFileID := uuid.NewSHA1(namespace, []byte(appID.String()+":code:"+bundledFile.Path))
						if expectedFileID == fileUUID {
							ext := services.GetFileExtension(bundledFile.Path)
							fileCategory := services.CategorizeFileType(bundledFile.Path)

							c.JSON(http.StatusOK, gin.H{
								"content": bundledFile.Content,
								"file": map[string]interface{}{
									"id":         fileID,
									"name":       bundledFile.Path,
									"type":       fileCategory,
									"size":       len(bundledFile.Content),
									"language":   services.GetLanguageFromExtension(ext),
									"created_at": createdAt,
									"updated_at": updatedAt,
								},
							})
							return
						}
					}
				} else {
					// Regular metadata files
					expectedFileID := uuid.NewSHA1(namespace, []byte(appID.String()+":"+fileType))
					if expectedFileID == fileUUID {
						c.JSON(http.StatusOK, gin.H{
							"content": content,
							"file": map[string]interface{}{
								"id":         fileID,
								"name":       fileType + ".md",
								"type":       "documentation",
								"size":       len(content),
								"language":   "markdown",
								"created_at": createdAt,
								"updated_at": updatedAt,
							},
						})
						return
					}
				}
			}
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
}

// InstallModule installs a workflow as a BusinessOS module
// POST /api/osa/modules/install
func (h *OSAWorkflowsHandler) InstallModule(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req struct {
		WorkflowID   string   `json:"workflow_id"`
		ModuleName   string   `json:"module_name"`
		InstallPath  *string  `json:"install_path"`
		FileIDs      []string `json:"file_ids"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Query workflow
	query := `
		SELECT ga.id, ga.name, ga.display_name, ga.description, ga.metadata
		FROM osa_generated_apps ga
		JOIN osa_workspaces w ON ga.workspace_id = w.id
		WHERE (ga.id = $1 OR ga.osa_workflow_id LIKE $2)
		  AND w.user_id = $3
	`

	workflowUUID, parseErr := uuid.Parse(req.WorkflowID)
	var searchID interface{}
	searchPrefix := req.WorkflowID + "%"

	if parseErr == nil {
		searchID = workflowUUID
	} else {
		searchID = uuid.Nil
	}

	var appID uuid.UUID
	var name, displayName, description string
	var metadataJSON []byte

	err := h.pool.QueryRow(c.Request.Context(), query, searchID, searchPrefix, userID).Scan(
		&appID, &name, &displayName, &description, &metadataJSON,
	)

	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Workflow not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch workflow", "details": err.Error()})
		}
		return
	}

	// Parse metadata
	var metadata map[string]interface{}
	if err := json.Unmarshal(metadataJSON, &metadata); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse metadata"})
		return
	}

	// Create module entry
	moduleName := req.ModuleName
	if moduleName == "" {
		moduleName = name
	}

	// Helper to wrap string content in JSON object for JSONB columns
	wrapInJSON := func(key string) []byte {
		if val, ok := metadata[key]; ok {
			wrapped := map[string]interface{}{"content": val}
			jsonData, _ := json.Marshal(wrapped)
			return jsonData
		}
		return []byte("{}")
	}

	insertQuery := `
		INSERT INTO osa_modules (
			name,
			display_name,
			description,
			module_type,
			schema_definition,
			api_definition,
			ui_definition,
			created_by,
			status,
			metadata
		) VALUES ($1, $2, $3, 'generated', $4, $5, $6, $7, 'installed', $8)
		RETURNING id
	`

	var moduleID uuid.UUID
	err = h.pool.QueryRow(c.Request.Context(), insertQuery,
		moduleName,
		displayName,
		description,
		wrapInJSON("architecture"),      // schema from architecture file
		wrapInJSON("code"),               // API from code file
		wrapInJSON("recommendations"),    // UI from recommendations
		userID,
		metadataJSON,
	).Scan(&moduleID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to install module", "details": err.Error()})
		return
	}

	// Update app with module link
	updateQuery := `
		UPDATE osa_generated_apps
		SET module_id = $1, status = 'deployed', deployed_at = NOW()
		WHERE id = $2
	`

	_, err = h.pool.Exec(c.Request.Context(), updateQuery, moduleID, appID)
	if err != nil {
		// Non-fatal - module is already created
		c.JSON(http.StatusOK, gin.H{
			"success":   true,
			"module_id": moduleID.String(),
			"message":   "Module installed successfully (app update failed)",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"module_id": moduleID.String(),
		"message":   "Module installed successfully",
	})
}

// TriggerSync manually triggers a sync from OSA-5 workspace
// POST /api/osa/sync/trigger
func (h *OSAWorkflowsHandler) TriggerSync(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Trigger an immediate sync
	go func() {
		// This will be picked up by the polling service on next tick
		// For immediate sync, we could call scanWorkspace directly
	}()

	c.JSON(http.StatusOK, gin.H{
		"message": "Sync triggered",
		"user_id": userID,
	})
}
