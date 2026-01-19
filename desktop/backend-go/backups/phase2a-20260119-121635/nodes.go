package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rhl/businessos-backend/internal/database/sqlc"
	"github.com/rhl/businessos-backend/internal/middleware"
)

// ListNodes returns all nodes for the current user
func (h *Handlers) ListNodes(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	queries := sqlc.New(h.pool)
	nodes, err := queries.ListNodes(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list nodes"})
		return
	}

	c.JSON(http.StatusOK, TransformNodes(nodes))
}

// NodeTreeItem represents a node with its children for tree display
type NodeTreeItem struct {
	ID            string          `json:"id"`
	ParentID      *string         `json:"parent_id"`
	Name          string          `json:"name"`
	Type          string          `json:"type"`
	Health        *string         `json:"health"`
	Purpose       *string         `json:"purpose"`
	ThisWeekFocus json.RawMessage `json:"this_week_focus"`
	IsActive      bool            `json:"is_active"`
	IsArchived    bool            `json:"is_archived"`
	SortOrder     *int32          `json:"sort_order"`
	UpdatedAt     string          `json:"updated_at"`
	Children      []NodeTreeItem  `json:"children"`
	ChildrenCount int             `json:"children_count"`
}

// buildNodeTree converts flat list of nodes into hierarchical tree
func buildNodeTree(nodes []sqlc.Node, parentID *string) []NodeTreeItem {
	var result []NodeTreeItem

	for _, node := range nodes {
		nodeParentID := getNodeParentIDString(node.ParentID)

		// Check if this node's parent matches the requested parentID
		if (parentID == nil && nodeParentID == nil) || (parentID != nil && nodeParentID != nil && *parentID == *nodeParentID) {
			nodeIDStr := nodeUUIDToString(node.ID)
			children := buildNodeTree(nodes, &nodeIDStr)

			var health *string
			if node.Health.Valid {
				h := string(node.Health.Nodehealth)
				health = &h
			}

			isActive := false
			if node.IsActive != nil {
				isActive = *node.IsActive
			}
			isArchived := false
			if node.IsArchived != nil {
				isArchived = *node.IsArchived
			}

			item := NodeTreeItem{
				ID:            nodeIDStr,
				ParentID:      nodeParentID,
				Name:          node.Name,
				Type:          string(node.Type),
				Health:        health,
				Purpose:       node.Purpose,
				ThisWeekFocus: node.ThisWeekFocus,
				IsActive:      isActive,
				IsArchived:    isArchived,
				SortOrder:     node.SortOrder,
				UpdatedAt:     node.UpdatedAt.Time.Format("2006-01-02T15:04:05Z"),
				Children:      children,
				ChildrenCount: len(children),
			}
			result = append(result, item)
		}
	}

	// Sort by sort_order
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			orderI := int32(0)
			orderJ := int32(0)
			if result[i].SortOrder != nil {
				orderI = *result[i].SortOrder
			}
			if result[j].SortOrder != nil {
				orderJ = *result[j].SortOrder
			}
			if orderI > orderJ {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result
}

func getNodeParentIDString(parentID pgtype.UUID) *string {
	if !parentID.Valid {
		return nil
	}
	s := uuid.UUID(parentID.Bytes).String()
	return &s
}

func nodeUUIDToString(id pgtype.UUID) string {
	return uuid.UUID(id.Bytes).String()
}

// GetNodeTree returns all nodes organized as a tree
func (h *Handlers) GetNodeTree(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	queries := sqlc.New(h.pool)
	nodes, err := queries.GetNodeTree(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get node tree"})
		return
	}

	// Build tree structure starting from root nodes (no parent)
	tree := buildNodeTree(nodes, nil)

	c.JSON(http.StatusOK, tree)
}

// CreateNode creates a new node
func (h *Handlers) CreateNode(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var req struct {
		Name            string   `json:"name" binding:"required"`
		Type            string   `json:"type" binding:"required"`
		ParentID        *string  `json:"parent_id"`
		ContextID       *string  `json:"context_id"`
		Health          *string  `json:"health"`
		Purpose         *string  `json:"purpose"`
		CurrentStatus   *string  `json:"current_status"`
		ThisWeekFocus   []string `json:"this_week_focus"`
		DecisionQueue   []string `json:"decision_queue"`
		DelegationReady []string `json:"delegation_ready"`
		SortOrder       *int32   `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	queries := sqlc.New(h.pool)

	// Parse optional fields
	var parentID, contextID pgtype.UUID
	if req.ParentID != nil {
		if parsed, err := uuid.Parse(*req.ParentID); err == nil {
			parentID = pgtype.UUID{Bytes: parsed, Valid: true}
		}
	}
	if req.ContextID != nil {
		if parsed, err := uuid.Parse(*req.ContextID); err == nil {
			contextID = pgtype.UUID{Bytes: parsed, Valid: true}
		}
	}

	var health sqlc.NullNodehealth
	if req.Health != nil {
		health = sqlc.NullNodehealth{
			Nodehealth: stringToNodeHealth(*req.Health),
			Valid:      true,
		}
	}

	// Handle JSONB arrays
	thisWeekFocus := []byte("[]")
	if req.ThisWeekFocus != nil && len(req.ThisWeekFocus) > 0 {
		if focusJSON, err := json.Marshal(req.ThisWeekFocus); err == nil {
			thisWeekFocus = focusJSON
		}
	}

	decisionQueue := []byte("[]")
	if req.DecisionQueue != nil && len(req.DecisionQueue) > 0 {
		if queueJSON, err := json.Marshal(req.DecisionQueue); err == nil {
			decisionQueue = queueJSON
		}
	}

	delegationReady := []byte("[]")
	if req.DelegationReady != nil && len(req.DelegationReady) > 0 {
		if delegationJSON, err := json.Marshal(req.DelegationReady); err == nil {
			delegationReady = delegationJSON
		}
	}

	node, err := queries.CreateNode(c.Request.Context(), sqlc.CreateNodeParams{
		UserID:          user.ID,
		ParentID:        parentID,
		ContextID:       contextID,
		Name:            req.Name,
		Type:            stringToNodeType(req.Type),
		Health:          health,
		Purpose:         req.Purpose,
		CurrentStatus:   req.CurrentStatus,
		ThisWeekFocus:   thisWeekFocus,
		DecisionQueue:   decisionQueue,
		DelegationReady: delegationReady,
		SortOrder:       req.SortOrder,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create node"})
		return
	}

	// Auto-activate if user has no active nodes
	activeNode, _ := queries.GetActiveNode(c.Request.Context(), user.ID)
	if activeNode.ID.Bytes == [16]byte{} {
		// No active node exists, activate the newly created one
		queries.ActivateNode(c.Request.Context(), user.ID) // Deactivate all nodes
		queries.SetNodeActive(c.Request.Context(), node.ID) // Activate this one
	}

	c.JSON(http.StatusCreated, TransformNode(node))
}

// GetNode returns a single node
func (h *Handlers) GetNode(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid node ID"})
		return
	}

	queries := sqlc.New(h.pool)
	node, err := queries.GetNode(c.Request.Context(), sqlc.GetNodeParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Node not found"})
		return
	}

	// Check if children are requested
	if c.Query("include_children") == "true" {
		children, err := queries.GetNodeChildren(c.Request.Context(), sqlc.GetNodeChildrenParams{
			ParentID: pgtype.UUID{Bytes: id, Valid: true},
			UserID:   user.ID,
		})
		if err == nil {
			c.JSON(http.StatusOK, gin.H{
				"node":     TransformNode(node),
				"children": TransformNodes(children),
			})
			return
		}
	}

	c.JSON(http.StatusOK, TransformNode(node))
}

// GetNodeChildren returns children of a node
func (h *Handlers) GetNodeChildren(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid node ID"})
		return
	}

	queries := sqlc.New(h.pool)
	children, err := queries.GetNodeChildren(c.Request.Context(), sqlc.GetNodeChildrenParams{
		ParentID: pgtype.UUID{Bytes: id, Valid: true},
		UserID:   user.ID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get children"})
		return
	}

	c.JSON(http.StatusOK, TransformNodes(children))
}

// GetActiveNode returns the currently active node
func (h *Handlers) GetActiveNode(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	queries := sqlc.New(h.pool)
	node, err := queries.GetActiveNode(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No active node found"})
		return
	}

	c.JSON(http.StatusOK, TransformNode(node))
}

// ActivateNode sets a node as the active node
func (h *Handlers) ActivateNode(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid node ID"})
		return
	}

	queries := sqlc.New(h.pool)

	// Verify ownership
	_, err = queries.GetNode(c.Request.Context(), sqlc.GetNodeParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Node not found"})
		return
	}

	// Deactivate all other nodes first
	if err := queries.ActivateNode(c.Request.Context(), user.ID); err != nil {
		log.Printf("Warning: failed to deactivate other nodes: %v", err)
	}

	// Set this node as active
	node, err := queries.SetNodeActive(c.Request.Context(), pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to activate node"})
		return
	}

	c.JSON(http.StatusOK, TransformNode(node))
}

// DeactivateNode deactivates a node
func (h *Handlers) DeactivateNode(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid node ID"})
		return
	}

	queries := sqlc.New(h.pool)

	// Verify ownership
	_, err = queries.GetNode(c.Request.Context(), sqlc.GetNodeParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Node not found"})
		return
	}

	node, err := queries.DeactivateNode(c.Request.Context(), pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to deactivate node"})
		return
	}

	c.JSON(http.StatusOK, TransformNode(node))
}

// UpdateNode updates a node
func (h *Handlers) UpdateNode(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid node ID"})
		return
	}

	var req struct {
		Name            *string  `json:"name"`
		Type            *string  `json:"type"`
		ContextID       *string  `json:"context_id"`
		Health          *string  `json:"health"`
		Purpose         *string  `json:"purpose"`
		CurrentStatus   *string  `json:"current_status"`
		ThisWeekFocus   []string `json:"this_week_focus"`
		DecisionQueue   []string `json:"decision_queue"`
		DelegationReady []string `json:"delegation_ready"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	queries := sqlc.New(h.pool)

	// Get existing node
	existing, err := queries.GetNode(c.Request.Context(), sqlc.GetNodeParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Node not found"})
		return
	}

	// Build update params with existing values as defaults
	name := existing.Name
	if req.Name != nil {
		name = *req.Name
	}

	nodeType := existing.Type
	if req.Type != nil {
		nodeType = stringToNodeType(*req.Type)
	}

	contextID := existing.ContextID
	if req.ContextID != nil {
		if parsed, err := uuid.Parse(*req.ContextID); err == nil {
			contextID = pgtype.UUID{Bytes: parsed, Valid: true}
		}
	}

	health := existing.Health
	if req.Health != nil {
		health = sqlc.NullNodehealth{
			Nodehealth: stringToNodeHealth(*req.Health),
			Valid:      true,
		}
	}

	purpose := existing.Purpose
	if req.Purpose != nil {
		purpose = req.Purpose
	}

	currentStatus := existing.CurrentStatus
	if req.CurrentStatus != nil {
		currentStatus = req.CurrentStatus
	}

	thisWeekFocus := existing.ThisWeekFocus
	if req.ThisWeekFocus != nil {
		if focusJSON, err := json.Marshal(req.ThisWeekFocus); err == nil {
			thisWeekFocus = focusJSON
		}
	}

	decisionQueue := existing.DecisionQueue
	if req.DecisionQueue != nil {
		if queueJSON, err := json.Marshal(req.DecisionQueue); err == nil {
			decisionQueue = queueJSON
		}
	}

	delegationReady := existing.DelegationReady
	if req.DelegationReady != nil {
		if delegationJSON, err := json.Marshal(req.DelegationReady); err == nil {
			delegationReady = delegationJSON
		}
	}

	node, err := queries.UpdateNode(c.Request.Context(), sqlc.UpdateNodeParams{
		ID:              pgtype.UUID{Bytes: id, Valid: true},
		Name:            name,
		Type:            nodeType,
		ContextID:       contextID,
		Health:          health,
		Purpose:         purpose,
		CurrentStatus:   currentStatus,
		ThisWeekFocus:   thisWeekFocus,
		DecisionQueue:   decisionQueue,
		DelegationReady: delegationReady,
	})
	if err != nil {
		log.Printf("Failed to update node %s: %v", id.String(), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update node: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, TransformNode(node))
}

// ReorderNodes updates sort order for multiple nodes
func (h *Handlers) ReorderNodes(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var req struct {
		Orders []struct {
			ID        string `json:"id" binding:"required"`
			SortOrder int32  `json:"sort_order" binding:"required"`
		} `json:"orders" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	queries := sqlc.New(h.pool)

	for _, order := range req.Orders {
		id, err := uuid.Parse(order.ID)
		if err != nil {
			continue
		}

		// Verify ownership
		_, err = queries.GetNode(c.Request.Context(), sqlc.GetNodeParams{
			ID:     pgtype.UUID{Bytes: id, Valid: true},
			UserID: user.ID,
		})
		if err != nil {
			continue
		}

		if err := queries.UpdateNodeSortOrder(c.Request.Context(), sqlc.UpdateNodeSortOrderParams{
			ID:        pgtype.UUID{Bytes: id, Valid: true},
			SortOrder: &order.SortOrder,
		}); err != nil {
			log.Printf("Warning: failed to update node sort order: %v", err)
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Nodes reordered"})
}

// ArchiveNode archives a node
func (h *Handlers) ArchiveNode(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid node ID"})
		return
	}

	queries := sqlc.New(h.pool)

	// Verify ownership
	_, err = queries.GetNode(c.Request.Context(), sqlc.GetNodeParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Node not found"})
		return
	}

	node, err := queries.ArchiveNode(c.Request.Context(), pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to archive node"})
		return
	}

	c.JSON(http.StatusOK, TransformNode(node))
}

// UnarchiveNode restores an archived node
func (h *Handlers) UnarchiveNode(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid node ID"})
		return
	}

	queries := sqlc.New(h.pool)

	// Verify ownership
	_, err = queries.GetNode(c.Request.Context(), sqlc.GetNodeParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Node not found"})
		return
	}

	node, err := queries.UnarchiveNode(c.Request.Context(), pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unarchive node"})
		return
	}

	c.JSON(http.StatusOK, TransformNode(node))
}

// DeleteNode deletes a node
func (h *Handlers) DeleteNode(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid node ID"})
		return
	}

	queries := sqlc.New(h.pool)
	err = queries.DeleteNode(c.Request.Context(), sqlc.DeleteNodeParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete node"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Node deleted"})
}

// stringToNodeType converts a string to sqlc.NodeType
func stringToNodeType(t string) sqlc.Nodetype {
	typeMap := map[string]sqlc.Nodetype{
		"business":    sqlc.NodetypeBUSINESS,
		"project":     sqlc.NodetypePROJECT,
		"learning":    sqlc.NodetypeLEARNING,
		"operational": sqlc.NodetypeOPERATIONAL,
	}
	if enum, ok := typeMap[strings.ToLower(t)]; ok {
		return enum
	}
	return sqlc.NodetypeBUSINESS
}

// stringToNodeHealth converts a string to sqlc.Nodehealth
func stringToNodeHealth(h string) sqlc.Nodehealth {
	typeMap := map[string]sqlc.Nodehealth{
		"healthy":         sqlc.NodehealthHEALTHY,
		"needs_attention": sqlc.NodehealthNEEDSATTENTION,
		"critical":        sqlc.NodehealthCRITICAL,
		"not_started":     sqlc.NodehealthNOTSTARTED,
	}
	if enum, ok := typeMap[strings.ToLower(h)]; ok {
		return enum
	}
	return sqlc.NodehealthNOTSTARTED
}

// ===== NODE LINKING HANDLERS =====

// GetNodeLinks returns all linked items for a node
func (h *Handlers) GetNodeLinks(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid node ID"})
		return
	}

	queries := sqlc.New(h.pool)
	nodeID := pgtype.UUID{Bytes: id, Valid: true}

	// Verify ownership
	_, err = queries.GetNode(c.Request.Context(), sqlc.GetNodeParams{
		ID:     nodeID,
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Node not found"})
		return
	}

	// Get all linked items
	projects, _ := queries.GetNodeLinkedProjects(c.Request.Context(), nodeID)
	contexts, _ := queries.GetNodeLinkedContexts(c.Request.Context(), nodeID)
	conversations, _ := queries.GetNodeLinkedConversations(c.Request.Context(), nodeID)

	c.JSON(http.StatusOK, gin.H{
		"projects":      transformNodeLinkedProjects(projects),
		"contexts":      transformNodeLinkedContexts(contexts),
		"conversations": transformNodeLinkedConversations(conversations),
	})
}

// GetNodeLinkCounts returns counts of linked items for a node
func (h *Handlers) GetNodeLinkCounts(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid node ID"})
		return
	}

	queries := sqlc.New(h.pool)
	nodeID := pgtype.UUID{Bytes: id, Valid: true}

	// Verify ownership
	_, err = queries.GetNode(c.Request.Context(), sqlc.GetNodeParams{
		ID:     nodeID,
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Node not found"})
		return
	}

	counts, err := queries.GetNodeLinkCounts(c.Request.Context(), nodeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get link counts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"linked_projects_count":      counts.LinkedProjectsCount,
		"linked_contexts_count":      counts.LinkedContextsCount,
		"linked_conversations_count": counts.LinkedConversationsCount,
	})
}

// LinkNodeProject links a project to a node
func (h *Handlers) LinkNodeProject(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	nodeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid node ID"})
		return
	}

	var req struct {
		ProjectID string `json:"project_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	projectID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	queries := sqlc.New(h.pool)

	// Verify node ownership
	_, err = queries.GetNode(c.Request.Context(), sqlc.GetNodeParams{
		ID:     pgtype.UUID{Bytes: nodeID, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Node not found"})
		return
	}

	err = queries.LinkNodeProject(c.Request.Context(), sqlc.LinkNodeProjectParams{
		NodeID:    pgtype.UUID{Bytes: nodeID, Valid: true},
		ProjectID: pgtype.UUID{Bytes: projectID, Valid: true},
		LinkedBy:  &user.ID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to link project"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Project linked"})
}

// UnlinkNodeProject unlinks a project from a node
func (h *Handlers) UnlinkNodeProject(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	nodeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid node ID"})
		return
	}

	projectID, err := uuid.Parse(c.Param("projectId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	queries := sqlc.New(h.pool)

	// Verify node ownership
	_, err = queries.GetNode(c.Request.Context(), sqlc.GetNodeParams{
		ID:     pgtype.UUID{Bytes: nodeID, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Node not found"})
		return
	}

	err = queries.UnlinkNodeProject(c.Request.Context(), sqlc.UnlinkNodeProjectParams{
		NodeID:    pgtype.UUID{Bytes: nodeID, Valid: true},
		ProjectID: pgtype.UUID{Bytes: projectID, Valid: true},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unlink project"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Project unlinked"})
}

// LinkNodeContext links a context to a node
func (h *Handlers) LinkNodeContext(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	nodeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid node ID"})
		return
	}

	var req struct {
		ContextID string `json:"context_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	contextID, err := uuid.Parse(req.ContextID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid context ID"})
		return
	}

	queries := sqlc.New(h.pool)

	// Verify node ownership
	_, err = queries.GetNode(c.Request.Context(), sqlc.GetNodeParams{
		ID:     pgtype.UUID{Bytes: nodeID, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Node not found"})
		return
	}

	err = queries.LinkNodeContext(c.Request.Context(), sqlc.LinkNodeContextParams{
		NodeID:    pgtype.UUID{Bytes: nodeID, Valid: true},
		ContextID: pgtype.UUID{Bytes: contextID, Valid: true},
		LinkedBy:  &user.ID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to link context"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Context linked"})
}

// UnlinkNodeContext unlinks a context from a node
func (h *Handlers) UnlinkNodeContext(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	nodeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid node ID"})
		return
	}

	contextID, err := uuid.Parse(c.Param("contextId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid context ID"})
		return
	}

	queries := sqlc.New(h.pool)

	// Verify node ownership
	_, err = queries.GetNode(c.Request.Context(), sqlc.GetNodeParams{
		ID:     pgtype.UUID{Bytes: nodeID, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Node not found"})
		return
	}

	err = queries.UnlinkNodeContext(c.Request.Context(), sqlc.UnlinkNodeContextParams{
		NodeID:    pgtype.UUID{Bytes: nodeID, Valid: true},
		ContextID: pgtype.UUID{Bytes: contextID, Valid: true},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unlink context"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Context unlinked"})
}

// LinkNodeConversation links a conversation to a node
func (h *Handlers) LinkNodeConversation(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	nodeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid node ID"})
		return
	}

	var req struct {
		ConversationID string `json:"conversation_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conversationID, err := uuid.Parse(req.ConversationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conversation ID"})
		return
	}

	queries := sqlc.New(h.pool)

	// Verify node ownership
	_, err = queries.GetNode(c.Request.Context(), sqlc.GetNodeParams{
		ID:     pgtype.UUID{Bytes: nodeID, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Node not found"})
		return
	}

	err = queries.LinkNodeConversation(c.Request.Context(), sqlc.LinkNodeConversationParams{
		NodeID:         pgtype.UUID{Bytes: nodeID, Valid: true},
		ConversationID: pgtype.UUID{Bytes: conversationID, Valid: true},
		LinkedBy:       &user.ID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to link conversation"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Conversation linked"})
}

// UnlinkNodeConversation unlinks a conversation from a node
func (h *Handlers) UnlinkNodeConversation(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	nodeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid node ID"})
		return
	}

	conversationID, err := uuid.Parse(c.Param("conversationId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conversation ID"})
		return
	}

	queries := sqlc.New(h.pool)

	// Verify node ownership
	_, err = queries.GetNode(c.Request.Context(), sqlc.GetNodeParams{
		ID:     pgtype.UUID{Bytes: nodeID, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Node not found"})
		return
	}

	err = queries.UnlinkNodeConversation(c.Request.Context(), sqlc.UnlinkNodeConversationParams{
		NodeID:         pgtype.UUID{Bytes: nodeID, Valid: true},
		ConversationID: pgtype.UUID{Bytes: conversationID, Valid: true},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unlink conversation"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Conversation unlinked"})
}

// Transform functions for linked items
func transformNodeLinkedProjects(projects []sqlc.GetNodeLinkedProjectsRow) []gin.H {
	result := make([]gin.H, 0, len(projects))
	for _, p := range projects {
		item := gin.H{
			"id":         uuid.UUID(p.ID.Bytes).String(),
			"name":       p.Name,
			"linked_at":  p.LinkedAt.Time.Format("2006-01-02T15:04:05Z"),
		}
		if p.Description != nil {
			item["description"] = *p.Description
		}
		if p.Status.Valid {
			item["status"] = string(p.Status.Projectstatus)
		}
		if p.Priority.Valid {
			item["priority"] = string(p.Priority.Projectpriority)
		}
		result = append(result, item)
	}
	return result
}

func transformNodeLinkedContexts(contexts []sqlc.GetNodeLinkedContextsRow) []gin.H {
	result := make([]gin.H, 0, len(contexts))
	for _, ctx := range contexts {
		item := gin.H{
			"id":        uuid.UUID(ctx.ID.Bytes).String(),
			"name":      ctx.Name,
			"linked_at": ctx.LinkedAt.Time.Format("2006-01-02T15:04:05Z"),
		}
		if ctx.Type.Valid {
			item["type"] = string(ctx.Type.Contexttype)
		}
		if ctx.Icon != nil {
			item["icon"] = *ctx.Icon
		}
		if ctx.WordCount != nil {
			item["word_count"] = *ctx.WordCount
		}
		result = append(result, item)
	}
	return result
}

func transformNodeLinkedConversations(conversations []sqlc.GetNodeLinkedConversationsRow) []gin.H {
	result := make([]gin.H, 0, len(conversations))
	for _, conv := range conversations {
		item := gin.H{
			"id":         uuid.UUID(conv.ID.Bytes).String(),
			"linked_at":  conv.LinkedAt.Time.Format("2006-01-02T15:04:05Z"),
			"created_at": conv.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
		}
		if conv.Title != nil {
			item["title"] = *conv.Title
		} else {
			item["title"] = "New Conversation"
		}
		result = append(result, item)
	}
	return result
}
