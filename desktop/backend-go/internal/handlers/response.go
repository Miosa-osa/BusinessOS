package handlers

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rhl/businessos-backend/internal/database/sqlc"
)

// Helper functions for converting pgtype values to JSON-friendly formats

func pgtypeUUIDToString(u pgtype.UUID) *string {
	if !u.Valid {
		return nil
	}
	s := uuid.UUID(u.Bytes).String()
	return &s
}

func pgtypeUUIDToStringRequired(u pgtype.UUID) string {
	if !u.Valid {
		return ""
	}
	return uuid.UUID(u.Bytes).String()
}

func pgtypeTimestampToString(t pgtype.Timestamp) *string {
	if !t.Valid {
		return nil
	}
	s := t.Time.Format(time.RFC3339)
	return &s
}

func pgtypeTimestamptzToString(t pgtype.Timestamptz) *string {
	if !t.Valid {
		return nil
	}
	s := t.Time.Format(time.RFC3339)
	return &s
}

func pgtypeDateToString(d pgtype.Date) *string {
	if !d.Valid {
		return nil
	}
	s := d.Time.Format("2006-01-02")
	return &s
}

func pgtypeNumericToFloat(n pgtype.Numeric) *float64 {
	if !n.Valid {
		return nil
	}
	f, _ := n.Float64Value()
	if !f.Valid {
		return nil
	}
	return &f.Float64
}

// Task response transformation
type TaskResponse struct {
	ID          string  `json:"id"`
	UserID      string  `json:"user_id"`
	Title       string  `json:"title"`
	Description *string `json:"description"`
	Status      string  `json:"status"`
	Priority    string  `json:"priority"`
	DueDate     *string `json:"due_date"`
	CompletedAt *string `json:"completed_at"`
	ProjectID   *string `json:"project_id"`
	AssigneeID  *string `json:"assignee_id"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

func TransformTask(t sqlc.Task) TaskResponse {
	status := "todo"
	if t.Status.Valid {
		status = string(t.Status.Taskstatus)
	}

	priority := "medium"
	if t.Priority.Valid {
		priority = string(t.Priority.Taskpriority)
	}

	return TaskResponse{
		ID:          pgtypeUUIDToStringRequired(t.ID),
		UserID:      t.UserID,
		Title:       t.Title,
		Description: t.Description,
		Status:      status,
		Priority:    priority,
		DueDate:     pgtypeTimestampToString(t.DueDate),
		CompletedAt: pgtypeTimestampToString(t.CompletedAt),
		ProjectID:   pgtypeUUIDToString(t.ProjectID),
		AssigneeID:  pgtypeUUIDToString(t.AssigneeID),
		CreatedAt:   t.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:   t.UpdatedAt.Time.Format(time.RFC3339),
	}
}

func TransformTasks(tasks []sqlc.Task) []TaskResponse {
	result := make([]TaskResponse, len(tasks))
	for i, t := range tasks {
		result[i] = TransformTask(t)
	}
	return result
}

// TeamMember response transformation
type TeamMemberResponse struct {
	ID         string   `json:"id"`
	UserID     string   `json:"user_id"`
	Name       string   `json:"name"`
	Email      string   `json:"email"`
	Role       string   `json:"role"`
	AvatarUrl  *string  `json:"avatar_url"`
	Status     string   `json:"status"`
	Capacity   int32    `json:"capacity"`
	ManagerID  *string  `json:"manager_id"`
	Skills     []string `json:"skills"`
	HourlyRate *float64 `json:"hourly_rate"`
	JoinedAt   string   `json:"joined_at"`
	CreatedAt  string   `json:"created_at"`
	UpdatedAt  string   `json:"updated_at"`
}

func TransformTeamMember(m sqlc.TeamMember) TeamMemberResponse {
	status := "available"
	if m.Status.Valid {
		status = strings.ToLower(string(m.Status.Memberstatus))
	}

	capacity := int32(100)
	if m.Capacity != nil {
		capacity = *m.Capacity
	}

	var skills []string
	if m.Skills != nil {
		json.Unmarshal(m.Skills, &skills)
	}
	if skills == nil {
		skills = []string{}
	}

	return TeamMemberResponse{
		ID:         pgtypeUUIDToStringRequired(m.ID),
		UserID:     m.UserID,
		Name:       m.Name,
		Email:      m.Email,
		Role:       m.Role,
		AvatarUrl:  m.AvatarUrl,
		Status:     status,
		Capacity:   capacity,
		ManagerID:  pgtypeUUIDToString(m.ManagerID),
		Skills:     skills,
		HourlyRate: pgtypeNumericToFloat(m.HourlyRate),
		JoinedAt:   m.JoinedAt.Time.Format(time.RFC3339),
		CreatedAt:  m.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:  m.UpdatedAt.Time.Format(time.RFC3339),
	}
}

func TransformTeamMembers(members []sqlc.TeamMember) []TeamMemberResponse {
	result := make([]TeamMemberResponse, len(members))
	for i, m := range members {
		result[i] = TransformTeamMember(m)
	}
	return result
}

// TeamMemberListResponse for list view with additional computed fields
type TeamMemberListResponse struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Email           string   `json:"email"`
	Role            string   `json:"role"`
	AvatarUrl       *string  `json:"avatar_url"`
	Status          string   `json:"status"`
	Capacity        int32    `json:"capacity"`
	ManagerID       *string  `json:"manager_id"`
	ActiveProjects  int      `json:"active_projects"`
	OpenTasks       int      `json:"open_tasks"`
	JoinedAt        string   `json:"joined_at"`
}

func TransformTeamMemberListRow(m sqlc.ListTeamMembersRow) TeamMemberListResponse {
	status := "available"
	if m.Status.Valid {
		status = strings.ToLower(string(m.Status.Memberstatus))
	}

	capacity := int32(100)
	if m.Capacity != nil {
		capacity = *m.Capacity
	}

	return TeamMemberListResponse{
		ID:              pgtypeUUIDToStringRequired(m.ID),
		Name:            m.Name,
		Email:           m.Email,
		Role:            m.Role,
		AvatarUrl:       m.AvatarUrl,
		Status:          status,
		Capacity:        capacity,
		ManagerID:       pgtypeUUIDToString(m.ManagerID),
		ActiveProjects:  0, // Not computed in the query
		OpenTasks:       int(m.ActiveTaskCount),
		JoinedAt:        m.JoinedAt.Time.Format(time.RFC3339),
	}
}

func TransformTeamMemberListRows(members []sqlc.ListTeamMembersRow) []TeamMemberListResponse {
	result := make([]TeamMemberListResponse, len(members))
	for i, m := range members {
		result[i] = TransformTeamMemberListRow(m)
	}
	return result
}

// Project response transformation
type ProjectResponse struct {
	ID              string                 `json:"id"`
	UserID          string                 `json:"user_id"`
	Name            string                 `json:"name"`
	Description     *string                `json:"description"`
	Status          string                 `json:"status"`
	Priority        string                 `json:"priority"`
	ClientName      *string                `json:"client_name"`
	ProjectType     string                 `json:"project_type"`
	ProjectMetadata map[string]interface{} `json:"project_metadata"`
	Notes           []ProjectNoteResponse  `json:"notes"`
	CreatedAt       string                 `json:"created_at"`
	UpdatedAt       string                 `json:"updated_at"`
}

type ProjectNoteResponse struct {
	ID        string `json:"id"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

func TransformProject(p sqlc.Project) ProjectResponse {
	status := "active"
	if p.Status.Valid {
		status = strings.ToLower(string(p.Status.Projectstatus))
	}

	priority := "medium"
	if p.Priority.Valid {
		priority = strings.ToLower(string(p.Priority.Projectpriority))
	}

	projectType := "other"
	if p.ProjectType != nil {
		projectType = *p.ProjectType
	}

	var metadata map[string]interface{}
	if p.ProjectMetadata != nil {
		json.Unmarshal(p.ProjectMetadata, &metadata)
	}

	return ProjectResponse{
		ID:              pgtypeUUIDToStringRequired(p.ID),
		UserID:          p.UserID,
		Name:            p.Name,
		Description:     p.Description,
		Status:          status,
		Priority:        priority,
		ClientName:      p.ClientName,
		ProjectType:     projectType,
		ProjectMetadata: metadata,
		Notes:           []ProjectNoteResponse{}, // Will be populated separately
		CreatedAt:       p.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:       p.UpdatedAt.Time.Format(time.RFC3339),
	}
}

func TransformProjects(projects []sqlc.Project) []ProjectResponse {
	result := make([]ProjectResponse, len(projects))
	for i, p := range projects {
		result[i] = TransformProject(p)
	}
	return result
}

func TransformProjectNote(n sqlc.ProjectNote) ProjectNoteResponse {
	return ProjectNoteResponse{
		ID:        pgtypeUUIDToStringRequired(n.ID),
		Content:   n.Content,
		CreatedAt: n.CreatedAt.Time.Format(time.RFC3339),
	}
}

// Client response transformation
type ClientResponse struct {
	ID              string                 `json:"id"`
	UserID          string                 `json:"user_id"`
	Name            string                 `json:"name"`
	Type            string                 `json:"type"`
	Email           *string                `json:"email"`
	Phone           *string                `json:"phone"`
	Website         *string                `json:"website"`
	Industry        *string                `json:"industry"`
	CompanySize     *string                `json:"company_size"`
	Address         *string                `json:"address"`
	City            *string                `json:"city"`
	State           *string                `json:"state"`
	ZipCode         *string                `json:"zip_code"`
	Country         *string                `json:"country"`
	Status          string                 `json:"status"`
	Source          *string                `json:"source"`
	AssignedTo      *string                `json:"assigned_to"`
	LifetimeValue   *float64               `json:"lifetime_value"`
	Tags            []string               `json:"tags"`
	CustomFields    map[string]interface{} `json:"custom_fields"`
	Notes           *string                `json:"notes"`
	CreatedAt       string                 `json:"created_at"`
	UpdatedAt       string                 `json:"updated_at"`
	LastContactedAt *string                `json:"last_contacted_at"`
}

func TransformClient(c sqlc.Client) ClientResponse {
	clientType := "company"
	if c.Type.Valid {
		clientType = string(c.Type.Clienttype)
	}

	status := "lead"
	if c.Status.Valid {
		status = string(c.Status.Clientstatus)
	}

	var tags []string
	if c.Tags != nil {
		json.Unmarshal(c.Tags, &tags)
	}
	if tags == nil {
		tags = []string{}
	}

	var customFields map[string]interface{}
	if c.CustomFields != nil {
		json.Unmarshal(c.CustomFields, &customFields)
	}

	return ClientResponse{
		ID:              pgtypeUUIDToStringRequired(c.ID),
		UserID:          c.UserID,
		Name:            c.Name,
		Type:            clientType,
		Email:           c.Email,
		Phone:           c.Phone,
		Website:         c.Website,
		Industry:        c.Industry,
		CompanySize:     c.CompanySize,
		Address:         c.Address,
		City:            c.City,
		State:           c.State,
		ZipCode:         c.ZipCode,
		Country:         c.Country,
		Status:          status,
		Source:          c.Source,
		AssignedTo:      c.AssignedTo,
		LifetimeValue:   pgtypeNumericToFloat(c.LifetimeValue),
		Tags:            tags,
		CustomFields:    customFields,
		Notes:           c.Notes,
		CreatedAt:       c.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:       c.UpdatedAt.Time.Format(time.RFC3339),
		LastContactedAt: pgtypeTimestamptzToString(c.LastContactedAt),
	}
}

func TransformClients(clients []sqlc.Client) []ClientResponse {
	result := make([]ClientResponse, len(clients))
	for i, c := range clients {
		result[i] = TransformClient(c)
	}
	return result
}

// Context response transformation
type ContextResponse struct {
	ID                   string                   `json:"id"`
	UserID               string                   `json:"user_id"`
	Name                 string                   `json:"name"`
	Type                 string                   `json:"type"`
	Content              *string                  `json:"content"`
	StructuredData       map[string]interface{}   `json:"structured_data"`
	SystemPromptTemplate *string                  `json:"system_prompt_template"`
	Blocks               []map[string]interface{} `json:"blocks"`
	CoverImage           *string                  `json:"cover_image"`
	Icon                 *string                  `json:"icon"`
	ParentID             *string                  `json:"parent_id"`
	IsTemplate           bool                     `json:"is_template"`
	IsArchived           bool                     `json:"is_archived"`
	LastEditedAt         *string                  `json:"last_edited_at"`
	WordCount            int32                    `json:"word_count"`
	IsPublic             bool                     `json:"is_public"`
	ShareID              *string                  `json:"share_id"`
	PropertySchema       []map[string]interface{} `json:"property_schema"`
	Properties           map[string]interface{}   `json:"properties"`
	ClientID             *string                  `json:"client_id"`
	CreatedAt            string                   `json:"created_at"`
	UpdatedAt            string                   `json:"updated_at"`
}

func TransformContext(ctx sqlc.Context) ContextResponse {
	contextType := "document"
	if ctx.Type.Valid {
		contextType = strings.ToLower(string(ctx.Type.Contexttype))
	}

	isTemplate := false
	if ctx.IsTemplate != nil {
		isTemplate = *ctx.IsTemplate
	}

	isArchived := false
	if ctx.IsArchived != nil {
		isArchived = *ctx.IsArchived
	}

	wordCount := int32(0)
	if ctx.WordCount != nil {
		wordCount = *ctx.WordCount
	}

	isPublic := false
	if ctx.IsPublic != nil {
		isPublic = *ctx.IsPublic
	}

	var structuredData map[string]interface{}
	if ctx.StructuredData != nil {
		json.Unmarshal(ctx.StructuredData, &structuredData)
	}

	var blocks []map[string]interface{}
	if ctx.Blocks != nil {
		json.Unmarshal(ctx.Blocks, &blocks)
	}

	var propertySchema []map[string]interface{}
	if ctx.PropertySchema != nil {
		json.Unmarshal(ctx.PropertySchema, &propertySchema)
	}

	var properties map[string]interface{}
	if ctx.Properties != nil {
		json.Unmarshal(ctx.Properties, &properties)
	}

	return ContextResponse{
		ID:                   pgtypeUUIDToStringRequired(ctx.ID),
		UserID:               ctx.UserID,
		Name:                 ctx.Name,
		Type:                 contextType,
		Content:              ctx.Content,
		StructuredData:       structuredData,
		SystemPromptTemplate: ctx.SystemPromptTemplate,
		Blocks:               blocks,
		CoverImage:           ctx.CoverImage,
		Icon:                 ctx.Icon,
		ParentID:             pgtypeUUIDToString(ctx.ParentID),
		IsTemplate:           isTemplate,
		IsArchived:           isArchived,
		LastEditedAt:         pgtypeTimestampToString(ctx.LastEditedAt),
		WordCount:            wordCount,
		IsPublic:             isPublic,
		ShareID:              ctx.ShareID,
		PropertySchema:       propertySchema,
		Properties:           properties,
		ClientID:             pgtypeUUIDToString(ctx.ClientID),
		CreatedAt:            ctx.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:            ctx.UpdatedAt.Time.Format(time.RFC3339),
	}
}

func TransformContexts(contexts []sqlc.Context) []ContextResponse {
	result := make([]ContextResponse, len(contexts))
	for i, ctx := range contexts {
		result[i] = TransformContext(ctx)
	}
	return result
}

// Node response transformation
type NodeResponse struct {
	ID              string   `json:"id"`
	UserID          string   `json:"user_id"`
	ParentID        *string  `json:"parent_id"`
	ContextID       *string  `json:"context_id"`
	Name            string   `json:"name"`
	Type            string   `json:"type"`
	Health          string   `json:"health"`
	Purpose         *string  `json:"purpose"`
	CurrentStatus   *string  `json:"current_status"`
	ThisWeekFocus   []string `json:"this_week_focus"`
	DecisionQueue   []string `json:"decision_queue"`
	DelegationReady []string `json:"delegation_ready"`
	IsActive        bool     `json:"is_active"`
	IsArchived      bool     `json:"is_archived"`
	SortOrder       int32    `json:"sort_order"`
	CreatedAt       string   `json:"created_at"`
	UpdatedAt       string   `json:"updated_at"`
}

func TransformNode(n sqlc.Node) NodeResponse {
	nodeType := "business"
	nodeType = strings.ToLower(string(n.Type))

	health := "not_started"
	if n.Health.Valid {
		health = strings.ToLower(string(n.Health.Nodehealth))
	}

	isActive := false
	if n.IsActive != nil {
		isActive = *n.IsActive
	}

	isArchived := false
	if n.IsArchived != nil {
		isArchived = *n.IsArchived
	}

	sortOrder := int32(0)
	if n.SortOrder != nil {
		sortOrder = *n.SortOrder
	}

	var thisWeekFocus []string
	if n.ThisWeekFocus != nil {
		json.Unmarshal(n.ThisWeekFocus, &thisWeekFocus)
	}
	if thisWeekFocus == nil {
		thisWeekFocus = []string{}
	}

	var decisionQueue []string
	if n.DecisionQueue != nil {
		json.Unmarshal(n.DecisionQueue, &decisionQueue)
	}
	if decisionQueue == nil {
		decisionQueue = []string{}
	}

	var delegationReady []string
	if n.DelegationReady != nil {
		json.Unmarshal(n.DelegationReady, &delegationReady)
	}
	if delegationReady == nil {
		delegationReady = []string{}
	}

	return NodeResponse{
		ID:              pgtypeUUIDToStringRequired(n.ID),
		UserID:          n.UserID,
		ParentID:        pgtypeUUIDToString(n.ParentID),
		ContextID:       pgtypeUUIDToString(n.ContextID),
		Name:            n.Name,
		Type:            nodeType,
		Health:          health,
		Purpose:         n.Purpose,
		CurrentStatus:   n.CurrentStatus,
		ThisWeekFocus:   thisWeekFocus,
		DecisionQueue:   decisionQueue,
		DelegationReady: delegationReady,
		IsActive:        isActive,
		IsArchived:      isArchived,
		SortOrder:       sortOrder,
		CreatedAt:       n.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:       n.UpdatedAt.Time.Format(time.RFC3339),
	}
}

func TransformNodes(nodes []sqlc.Node) []NodeResponse {
	result := make([]NodeResponse, len(nodes))
	for i, n := range nodes {
		result[i] = TransformNode(n)
	}
	return result
}

// Deal response transformation
type DealResponse struct {
	ID                string  `json:"id"`
	ClientID          string  `json:"client_id"`
	Name              string  `json:"name"`
	Value             float64 `json:"value"`
	Stage             string  `json:"stage"`
	Probability       int32   `json:"probability"`
	ExpectedCloseDate *string `json:"expected_close_date"`
	Notes             *string `json:"notes"`
	CreatedAt         string  `json:"created_at"`
	UpdatedAt         string  `json:"updated_at"`
	ClosedAt          *string `json:"closed_at"`
}

func TransformDeal(d sqlc.ClientDeal) DealResponse {
	stage := "qualification"
	if d.Stage.Valid {
		stage = string(d.Stage.Dealstage)
	}

	value := float64(0)
	if v := pgtypeNumericToFloat(d.Value); v != nil {
		value = *v
	}

	probability := int32(0)
	if d.Probability != nil {
		probability = *d.Probability
	}

	return DealResponse{
		ID:                pgtypeUUIDToStringRequired(d.ID),
		ClientID:          pgtypeUUIDToStringRequired(d.ClientID),
		Name:              d.Name,
		Value:             value,
		Stage:             stage,
		Probability:       probability,
		ExpectedCloseDate: pgtypeDateToString(d.ExpectedCloseDate),
		Notes:             d.Notes,
		CreatedAt:         d.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:         d.UpdatedAt.Time.Format(time.RFC3339),
		ClosedAt:          pgtypeTimestamptzToString(d.ClosedAt),
	}
}

func TransformDeals(deals []sqlc.ClientDeal) []DealResponse {
	result := make([]DealResponse, len(deals))
	for i, d := range deals {
		result[i] = TransformDeal(d)
	}
	return result
}

// FocusItem response transformation
type FocusItemResponse struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Text      string `json:"text"`
	Completed bool   `json:"completed"`
	FocusDate string `json:"focus_date"`
	CreatedAt string `json:"created_at"`
}

func TransformFocusItem(f sqlc.FocusItem) FocusItemResponse {
	completed := false
	if f.Completed != nil {
		completed = *f.Completed
	}

	return FocusItemResponse{
		ID:        pgtypeUUIDToStringRequired(f.ID),
		UserID:    f.UserID,
		Text:      f.Text,
		Completed: completed,
		FocusDate: f.FocusDate.Time.Format("2006-01-02"),
		CreatedAt: f.CreatedAt.Time.Format(time.RFC3339),
	}
}

func TransformFocusItems(items []sqlc.FocusItem) []FocusItemResponse {
	result := make([]FocusItemResponse, len(items))
	for i, f := range items {
		result[i] = TransformFocusItem(f)
	}
	return result
}

// Contact response transformation
type ContactResponse struct {
	ID        string  `json:"id"`
	ClientID  string  `json:"client_id"`
	Name      string  `json:"name"`
	Email     *string `json:"email"`
	Phone     *string `json:"phone"`
	Role      *string `json:"role"`
	IsPrimary bool    `json:"is_primary"`
	Notes     *string `json:"notes"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

func TransformContact(c sqlc.ClientContact) ContactResponse {
	isPrimary := false
	if c.IsPrimary != nil {
		isPrimary = *c.IsPrimary
	}

	return ContactResponse{
		ID:        pgtypeUUIDToStringRequired(c.ID),
		ClientID:  pgtypeUUIDToStringRequired(c.ClientID),
		Name:      c.Name,
		Email:     c.Email,
		Phone:     c.Phone,
		Role:      c.Role,
		IsPrimary: isPrimary,
		Notes:     c.Notes,
		CreatedAt: c.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt: c.UpdatedAt.Time.Format(time.RFC3339),
	}
}

func TransformContacts(contacts []sqlc.ClientContact) []ContactResponse {
	result := make([]ContactResponse, len(contacts))
	for i, c := range contacts {
		result[i] = TransformContact(c)
	}
	return result
}

// Interaction response transformation
type InteractionResponse struct {
	ID          string  `json:"id"`
	ClientID    string  `json:"client_id"`
	ContactID   *string `json:"contact_id"`
	Type        string  `json:"type"`
	Subject     string  `json:"subject"`
	Description *string `json:"description"`
	Outcome     *string `json:"outcome"`
	OccurredAt  string  `json:"occurred_at"`
	CreatedAt   string  `json:"created_at"`
}

func TransformInteraction(i sqlc.ClientInteraction) InteractionResponse {
	// InteractionType is not nullable in the model
	interactionType := string(i.Type)

	return InteractionResponse{
		ID:          pgtypeUUIDToStringRequired(i.ID),
		ClientID:    pgtypeUUIDToStringRequired(i.ClientID),
		ContactID:   pgtypeUUIDToString(i.ContactID),
		Type:        interactionType,
		Subject:     i.Subject,
		Description: i.Description,
		Outcome:     i.Outcome,
		OccurredAt:  i.OccurredAt.Time.Format(time.RFC3339),
		CreatedAt:   i.CreatedAt.Time.Format(time.RFC3339),
	}
}

func TransformInteractions(interactions []sqlc.ClientInteraction) []InteractionResponse {
	result := make([]InteractionResponse, len(interactions))
	for i, inter := range interactions {
		result[i] = TransformInteraction(inter)
	}
	return result
}

// Artifact response transformation
type ArtifactResponse struct {
	ID             string  `json:"id"`
	UserID         string  `json:"user_id"`
	Title          string  `json:"title"`
	Content        string  `json:"content"`
	Type           string  `json:"type"`
	Language       *string `json:"language"`
	Summary        *string `json:"summary"`
	ConversationID *string `json:"conversation_id"`
	MessageID      *string `json:"message_id"`
	ProjectID      *string `json:"project_id"`
	ContextID      *string `json:"context_id"`
	ContextName    *string `json:"context_name"`
	Version        int32   `json:"version"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

func TransformArtifact(a sqlc.Artifact) ArtifactResponse {
	// ArtifactType is not nullable in the model
	artifactType := strings.ToLower(string(a.Type))

	version := int32(1)
	if a.Version != nil {
		version = *a.Version
	}

	return ArtifactResponse{
		ID:             pgtypeUUIDToStringRequired(a.ID),
		UserID:         a.UserID,
		Title:          a.Title,
		Content:        a.Content,
		Type:           artifactType,
		Language:       a.Language,
		Summary:        a.Summary,
		ConversationID: pgtypeUUIDToString(a.ConversationID),
		MessageID:      pgtypeUUIDToString(a.MessageID),
		ProjectID:      pgtypeUUIDToString(a.ProjectID),
		ContextID:      pgtypeUUIDToString(a.ContextID),
		ContextName:    nil, // Populated separately if needed
		Version:        version,
		CreatedAt:      a.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:      a.UpdatedAt.Time.Format(time.RFC3339),
	}
}

func TransformArtifacts(artifacts []sqlc.Artifact) []ArtifactResponse {
	result := make([]ArtifactResponse, len(artifacts))
	for i, a := range artifacts {
		result[i] = TransformArtifact(a)
	}
	return result
}

// DealListResponse for deals with client name
type DealListResponse struct {
	ID                string  `json:"id"`
	ClientID          string  `json:"client_id"`
	ClientName        string  `json:"client_name"`
	Name              string  `json:"name"`
	Value             float64 `json:"value"`
	Stage             string  `json:"stage"`
	Probability       int32   `json:"probability"`
	ExpectedCloseDate *string `json:"expected_close_date"`
	Notes             *string `json:"notes"`
	CreatedAt         string  `json:"created_at"`
	UpdatedAt         string  `json:"updated_at"`
	ClosedAt          *string `json:"closed_at"`
}

func TransformDealListRow(d sqlc.ListDealsRow) DealListResponse {
	stage := "qualification"
	if d.Stage.Valid {
		stage = string(d.Stage.Dealstage)
	}

	value := float64(0)
	if v := pgtypeNumericToFloat(d.Value); v != nil {
		value = *v
	}

	probability := int32(0)
	if d.Probability != nil {
		probability = *d.Probability
	}

	return DealListResponse{
		ID:                pgtypeUUIDToStringRequired(d.ID),
		ClientID:          pgtypeUUIDToStringRequired(d.ClientID),
		ClientName:        d.ClientName,
		Name:              d.Name,
		Value:             value,
		Stage:             stage,
		Probability:       probability,
		ExpectedCloseDate: pgtypeDateToString(d.ExpectedCloseDate),
		Notes:             d.Notes,
		CreatedAt:         d.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:         d.UpdatedAt.Time.Format(time.RFC3339),
		ClosedAt:          pgtypeTimestamptzToString(d.ClosedAt),
	}
}

func TransformDealListRows(deals []sqlc.ListDealsRow) []DealListResponse {
	result := make([]DealListResponse, len(deals))
	for i, d := range deals {
		result[i] = TransformDealListRow(d)
	}
	return result
}

// Calendar Event Response
type CalendarEventResponse struct {
	ID            string                 `json:"id"`
	UserID        string                 `json:"user_id"`
	GoogleEventID *string                `json:"google_event_id"`
	CalendarID    *string                `json:"calendar_id"`
	Title         *string                `json:"title"`
	Description   *string                `json:"description"`
	StartTime     string                 `json:"start_time"`
	EndTime       string                 `json:"end_time"`
	AllDay        bool                   `json:"all_day"`
	Location      *string                `json:"location"`
	Attendees     []map[string]any       `json:"attendees"`
	Status        *string                `json:"status"`
	Visibility    *string                `json:"visibility"`
	HtmlLink      *string                `json:"html_link"`
	Source        *string                `json:"source"`
	MeetingType   string                 `json:"meeting_type"`
	ContextID     *string                `json:"context_id"`
	ProjectID     *string                `json:"project_id"`
	ClientID      *string                `json:"client_id"`
	RecordingURL  *string                `json:"recording_url"`
	MeetingLink   *string                `json:"meeting_link"`
	ExternalLinks []string               `json:"external_links"`
	MeetingNotes  *string                `json:"meeting_notes"`
	ActionItems   []string               `json:"action_items"`
	SyncedAt      string                 `json:"synced_at"`
	CreatedAt     string                 `json:"created_at"`
	UpdatedAt     string                 `json:"updated_at"`
}

func TransformCalendarEvent(e sqlc.CalendarEvent) CalendarEventResponse {
	meetingType := "other"
	if e.MeetingType.Valid {
		meetingType = strings.ToLower(string(e.MeetingType.Meetingtype))
	}

	allDay := false
	if e.AllDay != nil {
		allDay = *e.AllDay
	}

	var attendees []map[string]any
	if e.Attendees != nil {
		json.Unmarshal(e.Attendees, &attendees)
	}
	if attendees == nil {
		attendees = []map[string]any{}
	}

	var externalLinks []string
	if e.ExternalLinks != nil {
		json.Unmarshal(e.ExternalLinks, &externalLinks)
	}
	if externalLinks == nil {
		externalLinks = []string{}
	}

	var actionItems []string
	if e.ActionItems != nil {
		json.Unmarshal(e.ActionItems, &actionItems)
	}
	if actionItems == nil {
		actionItems = []string{}
	}

	return CalendarEventResponse{
		ID:            pgtypeUUIDToStringRequired(e.ID),
		UserID:        e.UserID,
		GoogleEventID: e.GoogleEventID,
		CalendarID:    e.CalendarID,
		Title:         e.Title,
		Description:   e.Description,
		StartTime:     e.StartTime.Time.Format(time.RFC3339),
		EndTime:       e.EndTime.Time.Format(time.RFC3339),
		AllDay:        allDay,
		Location:      e.Location,
		Attendees:     attendees,
		Status:        e.Status,
		Visibility:    e.Visibility,
		HtmlLink:      e.HtmlLink,
		Source:        e.Source,
		MeetingType:   meetingType,
		ContextID:     pgtypeUUIDToString(e.ContextID),
		ProjectID:     pgtypeUUIDToString(e.ProjectID),
		ClientID:      pgtypeUUIDToString(e.ClientID),
		RecordingURL:  e.RecordingUrl,
		MeetingLink:   e.MeetingLink,
		ExternalLinks: externalLinks,
		MeetingNotes:  e.MeetingNotes,
		ActionItems:   actionItems,
		SyncedAt:      e.SyncedAt.Time.Format(time.RFC3339),
		CreatedAt:     e.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:     e.UpdatedAt.Time.Format(time.RFC3339),
	}
}

func TransformCalendarEvents(events []sqlc.CalendarEvent) []CalendarEventResponse {
	result := make([]CalendarEventResponse, len(events))
	for i, e := range events {
		result[i] = TransformCalendarEvent(e)
	}
	return result
}

// Message response transformation
type MessageResponse struct {
	ID             string  `json:"id"`
	ConversationID string  `json:"conversation_id"`
	Role           string  `json:"role"`
	Content        string  `json:"content"`
	CreatedAt      string  `json:"created_at"`
}

func TransformMessage(m sqlc.Message) MessageResponse {
	return MessageResponse{
		ID:             pgtypeUUIDToStringRequired(m.ID),
		ConversationID: pgtypeUUIDToStringRequired(m.ConversationID),
		Role:           strings.ToLower(string(m.Role)),
		Content:        m.Content,
		CreatedAt:      m.CreatedAt.Time.Format(time.RFC3339),
	}
}

func TransformMessages(messages []sqlc.Message) []MessageResponse {
	result := make([]MessageResponse, len(messages))
	for i, m := range messages {
		result[i] = TransformMessage(m)
	}
	return result
}

// Conversation response transformation
type ConversationResponse struct {
	ID           string  `json:"id"`
	UserID       string  `json:"user_id"`
	Title        string  `json:"title"`
	ContextID    *string `json:"context_id"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
	MessageCount int64   `json:"message_count"`
}

func TransformConversation(c sqlc.Conversation) ConversationResponse {
	title := "New Conversation"
	if c.Title != nil {
		title = *c.Title
	}

	return ConversationResponse{
		ID:        pgtypeUUIDToStringRequired(c.ID),
		UserID:    c.UserID,
		Title:     title,
		ContextID: pgtypeUUIDToString(c.ContextID),
		CreatedAt: c.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt: c.UpdatedAt.Time.Format(time.RFC3339),
	}
}

func TransformConversationListRow(c sqlc.ListConversationsRow) ConversationResponse {
	title := "New Conversation"
	if c.Title != nil {
		title = *c.Title
	}

	return ConversationResponse{
		ID:           pgtypeUUIDToStringRequired(c.ID),
		UserID:       c.UserID,
		Title:        title,
		ContextID:    pgtypeUUIDToString(c.ContextID),
		CreatedAt:    c.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:    c.UpdatedAt.Time.Format(time.RFC3339),
		MessageCount: c.MessageCount,
	}
}

func TransformConversationListRows(conversations []sqlc.ListConversationsRow) []ConversationResponse {
	result := make([]ConversationResponse, len(conversations))
	for i, c := range conversations {
		result[i] = TransformConversationListRow(c)
	}
	return result
}

func TransformConversationByContextRow(c sqlc.ListConversationsByContextRow) ConversationResponse {
	title := "New Conversation"
	if c.Title != nil {
		title = *c.Title
	}

	return ConversationResponse{
		ID:           pgtypeUUIDToStringRequired(c.ID),
		UserID:       c.UserID,
		Title:        title,
		ContextID:    pgtypeUUIDToString(c.ContextID),
		CreatedAt:    c.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:    c.UpdatedAt.Time.Format(time.RFC3339),
		MessageCount: c.MessageCount,
	}
}

func TransformConversationsByContextRows(conversations []sqlc.ListConversationsByContextRow) []ConversationResponse {
	result := make([]ConversationResponse, len(conversations))
	for i, c := range conversations {
		result[i] = TransformConversationByContextRow(c)
	}
	return result
}
