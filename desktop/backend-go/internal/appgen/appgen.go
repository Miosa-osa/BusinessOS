package appgen

import (
	"time"
)

// AgentType defines the type of agent in app generation
type AgentType string

const (
	AgentFrontend  AgentType = "frontend"
	AgentBackend   AgentType = "backend"
	AgentDatabase  AgentType = "database"
	AgentTest      AgentType = "test"
)

// ProgressEvent represents progress in app generation
type ProgressEvent struct {
	TaskID    string
	AgentType AgentType
	Progress  int
	Message   string
	Status    string
	Timestamp time.Time
}

// Plan represents an app generation plan
type Plan struct {
	ID    string
	Name  string
	Tasks []PlanTask
}

// PlanTask represents a single task in a plan
type PlanTask struct {
	ID          string
	Type        AgentType
	AgentType   AgentType
	Title       string
	Description string
	Status      string
}

// AgentResult represents the result of an agent's work
type AgentResult struct {
	AgentType AgentType
	Status    string
	Output    map[string]string
	Error     string
}

// GeneratedApp represents a complete generated application
type GeneratedApp struct {
	AppID   string
	Name    string
	Results []AgentResult
}
