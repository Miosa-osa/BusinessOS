# Agent Skills & Tools Implementation Task List

**Created:** January 5, 2026
**Updated:** January 6, 2026
**Status:** Ready for Implementation
**Assigned:** Javaris

---

## Architecture Reminder

> **Skills are instructions, Tools are execution.**

```
┌────────────────────────────────────────────────────────────────────┐
│  SKILL (Markdown)              │  TOOL (Go Code)                   │
│  skills/dashboard-management/  │  internal/tools/dashboard_tool.go │

│  └── SKILL.md                  │  └── ConfigureDashboardTool       │
│                                │                                    │
│  Teaches WHEN/HOW to use tool  │  Actually DOES the work           │
└────────────────────────────────────────────────────────────────────┘
```

**Build both together** - The tool is useless without the skill, and vice versa.

---

## Phase Summary

| Phase | Description | What You're Building |
|-------|-------------|---------------------|
| 1 | Infrastructure | Skills loader, API endpoints |
| 2 | Dashboard Tool | MCP tool (Go code) |
| 3 | Dashboard Skill | SKILL.md (Markdown instructions) |
| 4 | Other Skills | Task, project, notification skills |
| 5 | Integration | Connect to agent system prompt |
| 6 | Testing | End-to-end validation |

---

## Phase 1: Skills Infrastructure

### 1.1 Directory Structure

- [ ] **1.1.1** Create skills root folder:
  ```
  desktop/backend-go/skills/
  ```

- [ ] **1.1.2** Create skills config file:
  ```yaml
  # desktop/backend-go/skills/skills.yaml
  version: "1.0"
  skills_directory: "./skills"

  skills:
    - name: dashboard-management
      path: ./dashboard-management
      enabled: true

    - name: analytics-insights
      path: ./analytics-insights
      enabled: true

    - name: task-management
      path: ./task-management
      enabled: true

    - name: project-management
      path: ./project-management
      enabled: true

    - name: notification-management
      path: ./notification-management
      enabled: true
  ```

### 1.2 Skills Loader Service

- [ ] **1.2.1** Create `internal/services/skills_loader.go`:
  ```go
  type SkillsLoader struct {
      configPath string
      skills     []SkillMetadata
  }

  type SkillMetadata struct {
      Name        string   `yaml:"name"`
      Description string   `yaml:"description"`
      Path        string   `yaml:"path"`
      Enabled     bool     `yaml:"enabled"`
      ToolsUsed   []string `yaml:"tools_used"`
  }

  func NewSkillsLoader(configPath string) *SkillsLoader
  func (l *SkillsLoader) LoadConfig() error
  func (l *SkillsLoader) ParseSkillFrontmatter(skillPath string) (*SkillMetadata, error)
  func (l *SkillsLoader) GetEnabledSkills() []SkillMetadata
  func (l *SkillsLoader) GetSkillContent(name string) (string, error)
  func (l *SkillsLoader) GetSkillReference(name, refName string) (string, error)
  func (l *SkillsLoader) GetSkillsPromptXML() string
  ```

- [ ] **1.2.2** Implement YAML frontmatter parser:
  ```go
  // Parse SKILL.md frontmatter (between --- markers)
  func ParseFrontmatter(content string) (map[string]interface{}, string, error)
  ```

- [ ] **1.2.3** Initialize loader in `main.go`:
  ```go
  skillsLoader := services.NewSkillsLoader("./skills/skills.yaml")
  if err := skillsLoader.LoadConfig(); err != nil {
      log.Warn("Skills not loaded: ", err)
  }
  ```

### 1.3 Skills API Endpoints

- [ ] **1.3.1** Create `internal/handlers/skills_handlers.go`:
  ```go
  type SkillsHandler struct {
      loader *services.SkillsLoader
  }

  // GET /api/agent/skills - List enabled skills (metadata only)
  func (h *SkillsHandler) ListSkills(c *gin.Context)

  // GET /api/agent/skills/:name - Get full SKILL.md content
  func (h *SkillsHandler) GetSkill(c *gin.Context)

  // GET /api/agent/skills/:name/references/:ref - Get reference file
  func (h *SkillsHandler) GetSkillReference(c *gin.Context)

  // GET /api/agent/skills/prompt - Get <available_skills> XML
  func (h *SkillsHandler) GetSkillsPrompt(c *gin.Context)
  ```

- [ ] **1.3.2** Register routes in `handlers.go`:
  ```go
  // Agent Skills
  agent := api.Group("/agent")
  {
      agent.GET("/skills", h.ListSkills)
      agent.GET("/skills/prompt", h.GetSkillsPrompt)
      agent.GET("/skills/:name", h.GetSkill)
      agent.GET("/skills/:name/references/:ref", h.GetSkillReference)
  }
  ```

---

## Phase 2: Dashboard MCP Tool (Go Code)

This is the **execution layer** - actual functions that do database work.

### 2.1 Create Tool File

- [ ] **2.1.1** Create `internal/tools/dashboard_tool.go`

- [ ] **2.1.2** Define tool struct:
  ```go
  type ConfigureDashboardTool struct {
      queries *sqlc.Queries
      pool    *pgxpool.Pool
  }

  func (t *ConfigureDashboardTool) Name() string
  func (t *ConfigureDashboardTool) Description() string
  func (t *ConfigureDashboardTool) InputSchema() json.RawMessage
  func (t *ConfigureDashboardTool) Execute(ctx context.Context, userID string, input json.RawMessage) (json.RawMessage, error)
  ```

### 2.2 Implement Actions

- [ ] **2.2.1** `list_dashboards` - Get user's dashboards
- [ ] **2.2.2** `list_widgets` - Get available widget types
- [ ] **2.2.3** `get_dashboard` - Get specific dashboard with layout
- [ ] **2.2.4** `create_dashboard` - Create new dashboard
- [ ] **2.2.5** `add_widget` - Add single widget
- [ ] **2.2.6** `add_widgets` - Batch add widgets
- [ ] **2.2.7** `remove_widget` - Remove widget
- [ ] **2.2.8** `update_widget` - Update widget config
- [ ] **2.2.9** `set_default` - Set default dashboard

### 2.3 Supporting Functions

- [ ] **2.3.1** `findNextPosition()` - Auto-place algorithm
- [ ] **2.3.2** `validateWidgetType()` - With suggestions on error
- [ ] **2.3.3** `buildResponseContext()` - Include dashboards list, widget types
- [ ] **2.3.4** `getDefaultWidgetSize()` - Default dimensions per type

### 2.4 Register Tool

- [ ] **2.4.1** Register in MCP server or tool registry

---

## Phase 3: Dashboard Skill (Markdown Instructions)

This is the **instruction layer** - teaches agent when/how to use the tool.

### 3.1 Create Skill Directory

- [ ] **3.1.1** Create folder:
  ```
  desktop/backend-go/skills/dashboard-management/
  ├── SKILL.md
  └── references/
      ├── WIDGETS.md
      ├── CONFIGS.md
      └── EXAMPLES.md
  ```

### 3.2 Write SKILL.md

- [ ] **3.2.1** Frontmatter:
  ```yaml
  ---
  name: dashboard-management
  description: Create and configure custom dashboards with widgets. Add task summaries, burndown charts, project progress, upcoming deadlines, and metric cards.
  metadata:
    version: "1.0.0"
    author: businessos
    tools_used:
      - configure_dashboard
  ---
  ```

- [ ] **3.2.2** Sections to include:
  - When to Use This Skill
  - Available Tool (configure_dashboard)
  - Request → Tool Mapping (table)
  - Common Examples (5-6)
  - Error Handling guidance
  - Limitations

### 3.3 Write Reference Files

- [ ] **3.3.1** `references/WIDGETS.md`:
  - All 12 widget types
  - Description, default size
  - Config options per widget
  - ASCII mockups

- [ ] **3.3.2** `references/CONFIGS.md`:
  - Common config options
  - Validation rules
  - Default values

- [ ] **3.3.3** `references/EXAMPLES.md`:
  - Complex multi-widget examples
  - Role-specific dashboards
  - Troubleshooting scenarios

---

## Phase 4: Additional Skills

### 4.1 Analytics Insights Skill

- [ ] **4.1.1** Create `skills/analytics-insights/SKILL.md`:
  - Tool: `query_analytics` (or use existing analytics endpoints)
  - Actions: get_summary, get_burndown, get_workload, explain_metric

- [ ] **4.1.2** Create `references/METRICS.md`:
  - All metrics with calculations
  - Good vs bad values
  - Interpretation guidance

### 4.2 Task Management Skill

- [ ] **4.2.1** Create `skills/task-management/SKILL.md`:
  - Tool: `manage_tasks` (or existing task endpoints)
  - Actions: create, update, complete, delete, filter, bulk operations

### 4.3 Project Management Skill

- [ ] **4.3.1** Create `skills/project-management/SKILL.md`:
  - Tool: `manage_projects`
  - Actions: create, update, manage members, status tracking

### 4.4 Notification Management Skill

- [ ] **4.4.1** Create `skills/notification-management/SKILL.md`:
  - Tool: `configure_notifications`
  - Actions: view, mark read, configure preferences, quiet hours

---

## Phase 5: Agent Integration

### 5.1 System Prompt Injection

- [ ] **5.1.1** Update agent system prompt builder:
  ```go
  func (a *Agent) BuildSystemPrompt() string {
      skillsXML := a.skillsLoader.GetSkillsPromptXML()

      return fmt.Sprintf(`
      You are the BusinessOS assistant.

      %s

      When a user request matches a skill description, use the view tool
      to read the full SKILL.md file for detailed instructions.
      `, skillsXML)
  }
  ```

- [ ] **5.1.2** Format for `<available_skills>`:
  ```xml
  <available_skills>
    <skill name="dashboard-management">
      Create and configure custom dashboards with widgets. Add task summaries,
      burndown charts, project progress, upcoming deadlines, and metric cards.
    </skill>
    <skill name="task-management">
      Create, update, complete, and organize tasks. Filter by status, project,
      priority, or due date. Bulk operations supported.
    </skill>
  </available_skills>
  ```

### 5.2 Skill Activation Flow

- [ ] **5.2.1** Implement skill loading when agent needs it:
  ```go
  // When agent requests skill content
  func (a *Agent) LoadSkill(skillName string) error {
      content, err := a.skillsLoader.GetSkillContent(skillName)
      if err != nil {
          return err
      }
      // Add to conversation context
      a.context.AddSystemMessage(content)
      return nil
  }
  ```

- [ ] **5.2.2** Implement reference loading:
  ```go
  func (a *Agent) LoadSkillReference(skillName, refName string) error {
      content, err := a.skillsLoader.GetSkillReference(skillName, refName)
      // Add to context
  }
  ```

---

## Phase 6: Testing & Validation

### 6.1 Skill Validation

- [ ] **6.1.1** Create validation utility:
  ```go
  // cmd/validate-skills/main.go
  func main() {
      loader := services.NewSkillsLoader("./skills/skills.yaml")
      errors := loader.ValidateAll()
      // Check: frontmatter valid, name matches directory, refs exist
  }
  ```

### 6.2 API Testing

- [ ] **6.2.1** Test endpoints:
  ```
  GET /api/agent/skills → returns 5 skills
  GET /api/agent/skills/dashboard-management → returns SKILL.md
  GET /api/agent/skills/dashboard-management/references/WIDGETS → returns WIDGETS.md
  GET /api/agent/skills/prompt → returns XML block
  ```

### 6.3 Tool Testing

- [ ] **6.3.1** Test `configure_dashboard` tool:
  - list_dashboards returns user's dashboards
  - create_dashboard creates with correct name
  - add_widget auto-places correctly
  - Error responses include suggestions

### 6.4 End-to-End Testing

- [ ] **6.4.1** Test full flow:
  ```
  User: "Create a dashboard with my tasks grouped by project"

  1. Agent matches "dashboard-management" skill
  2. Agent loads SKILL.md
  3. Agent calls configure_dashboard tool
  4. Tool creates dashboard + widget
  5. Agent responds with confirmation
  ```

- [ ] **6.4.2** Test error recovery:
  ```
  User: "Add a team capacity widget"

  1. Agent calls configure_dashboard with team_capacity
  2. Tool returns error with alternatives
  3. Agent suggests workload_heatmap instead
  ```

---

## Files to Create

| File | Layer | Purpose |
|------|-------|---------|
| `skills/skills.yaml` | Config | Skills configuration |
| `skills/dashboard-management/SKILL.md` | Skill | Dashboard instructions |
| `skills/dashboard-management/references/WIDGETS.md` | Skill | Widget details |
| `skills/dashboard-management/references/CONFIGS.md` | Skill | Config schemas |
| `skills/dashboard-management/references/EXAMPLES.md` | Skill | More examples |
| `skills/analytics-insights/SKILL.md` | Skill | Analytics instructions |
| `skills/task-management/SKILL.md` | Skill | Task instructions |
| `skills/project-management/SKILL.md` | Skill | Project instructions |
| `skills/notification-management/SKILL.md` | Skill | Notification instructions |
| `internal/tools/dashboard_tool.go` | Tool | Dashboard MCP tool |
| `internal/services/skills_loader.go` | Service | Skills discovery |
| `internal/handlers/skills_handlers.go` | API | Skills endpoints |

---

## Implementation Order (Recommended)

1. **Phase 1.1-1.2** - Create directories and skills loader
2. **Phase 2.1-2.2** - Build dashboard tool (Go code)
3. **Phase 3.1-3.2** - Write dashboard SKILL.md
4. **Phase 1.3** - Add API endpoints
5. **Phase 5** - Connect to agent
6. **Phase 6** - Test everything
7. **Phase 3.3 + 4** - Reference files and other skills

---

## Notes

- Dashboard is the most complex - other skills follow the same pattern
- Skills can be updated without code redeploy (just edit .md files)
- Keep SKILL.md focused (~400 lines max), put details in references/
- Tool responses should always include context for agent decision-making