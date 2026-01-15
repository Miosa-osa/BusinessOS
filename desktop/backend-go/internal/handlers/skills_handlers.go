package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/services"
)

// SkillsHandler handles skill-related API endpoints
type SkillsHandler struct {
	loader *services.SkillsLoader
}

// NewSkillsHandler creates a new skills handler
func NewSkillsHandler(loader *services.SkillsLoader) *SkillsHandler {
	return &SkillsHandler{loader: loader}
}

// ListSkills returns all enabled skills with metadata
// GET /api/skills
func (h *SkillsHandler) ListSkills(c *gin.Context) {
	if h.loader == nil || !h.loader.IsLoaded() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Skills system not initialized",
		})
		return
	}

	skills := h.loader.GetEnabledSkills()
	
	response := make([]gin.H, 0, len(skills))
	for _, skill := range skills {
		response = append(response, gin.H{
			"name":        skill.Name,
			"description": skill.Description,
			"version":     skill.Version,
			"priority":    skill.Priority,
			"tools_used":  skill.ToolsUsed,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"skills": response,
		"count":  len(response),
	})
}

// GetSkill returns a specific skill's full content
// GET /api/skills/:name
func (h *SkillsHandler) GetSkill(c *gin.Context) {
	name := c.Param("name")

	if h.loader == nil || !h.loader.IsLoaded() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Skills system not initialized",
		})
		return
	}

	metadata := h.loader.GetSkillMetadata(name)
	if metadata == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Skill not found",
			"name":  name,
		})
		return
	}

	content, err := h.loader.GetSkillContent(name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to load skill content",
		})
		return
	}

	refs, _ := h.loader.ListSkillReferences(name)

	c.JSON(http.StatusOK, gin.H{
		"name":        metadata.Name,
		"description": metadata.Description,
		"version":     metadata.Version,
		"priority":    metadata.Priority,
		"tools_used":  metadata.ToolsUsed,
		"content":     content,
		"references":  refs,
	})
}

// GetSkillReference returns a specific reference file
// GET /api/skills/:name/references/:ref
func (h *SkillsHandler) GetSkillReference(c *gin.Context) {
	name := c.Param("name")
	ref := c.Param("ref")

	if h.loader == nil || !h.loader.IsLoaded() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Skills system not initialized",
		})
		return
	}

	content, err := h.loader.GetSkillReference(name, ref)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":     "Reference not found",
			"skill":     name,
			"reference": ref,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"skill":     name,
		"reference": ref,
		"content":   content,
	})
}

// GetSkillSchema returns the JSON schema for a skill
// GET /api/skills/:name/schemas/:schema
func (h *SkillsHandler) GetSkillSchema(c *gin.Context) {
	name := c.Param("name")
	schema := c.Param("schema")

	if h.loader == nil || !h.loader.IsLoaded() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Skills system not initialized",
		})
		return
	}

	content, err := h.loader.GetSkillSchema(name, schema)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":  "Schema not found",
			"skill":  name,
			"schema": schema,
		})
		return
	}

	c.Data(http.StatusOK, "application/json", []byte(content))
}

// ValidateSkill checks a skill for issues
// GET /api/skills/:name/validate
func (h *SkillsHandler) ValidateSkill(c *gin.Context) {
	name := c.Param("name")

	if h.loader == nil || !h.loader.IsLoaded() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Skills system not initialized",
		})
		return
	}

	issues := h.loader.ValidateSkill(name)
	
	valid := len(issues) == 0
	c.JSON(http.StatusOK, gin.H{
		"skill":  name,
		"valid":  valid,
		"issues": issues,
	})
}

// ReloadSkills reloads all skills from disk
// POST /api/skills/reload
func (h *SkillsHandler) ReloadSkills(c *gin.Context) {
	if h.loader == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Skills system not initialized",
		})
		return
	}

	if err := h.loader.Reload(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to reload skills",
		})
		return
	}

	skills := h.loader.GetEnabledSkills()
	c.JSON(http.StatusOK, gin.H{
		"message": "Skills reloaded",
		"count":   len(skills),
	})
}

// GetSkillsPrompt returns the XML prompt for agent integration
// GET /api/skills/prompt
func (h *SkillsHandler) GetSkillsPrompt(c *gin.Context) {
	if h.loader == nil || !h.loader.IsLoaded() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Skills system not initialized",
		})
		return
	}

	xml := h.loader.GetSkillsPromptXML()
	instructions := h.loader.GetSkillsPromptInstructions()

	c.JSON(http.StatusOK, gin.H{
		"skills_xml":   xml,
		"instructions": instructions,
	})
}

// GetSkillGroups returns available skill groups
// GET /api/skills/groups
func (h *SkillsHandler) GetSkillGroups(c *gin.Context) {
	if h.loader == nil || !h.loader.IsLoaded() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Skills system not initialized",
		})
		return
	}

	settings := h.loader.GetSettings()
	
	groups := make(map[string][]string)
	for _, groupName := range []string{"productivity", "intelligence", "system"} {
		if skills := h.loader.GetSkillGroup(groupName); len(skills) > 0 {
			groups[groupName] = skills
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"groups":   groups,
		"settings": settings,
	})
}

// RegisterSkillsRoutes registers skill routes on the given router group
func (h *SkillsHandler) RegisterRoutes(rg *gin.RouterGroup) {
	skills := rg.Group("/skills")
	{
		skills.GET("", h.ListSkills)
		skills.GET("/prompt", h.GetSkillsPrompt)
		skills.GET("/groups", h.GetSkillGroups)
		skills.POST("/reload", h.ReloadSkills)
		skills.GET("/:name", h.GetSkill)
		skills.GET("/:name/validate", h.ValidateSkill)
		skills.GET("/:name/references/:ref", h.GetSkillReference)
		skills.GET("/:name/schemas/:schema", h.GetSkillSchema)
	}
}
