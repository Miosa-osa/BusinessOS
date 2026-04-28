// Package tenancy ports OptimalEngine.Tenancy (Elixir) to Go.
//
// In BusinessOS the tenant model bridges to the existing `workspaces` table:
// every workspace IS a tenant. A reserved "default" tenant covers anonymous
// or pre-workspace activity (legacy contexts, system events) so the audit
// log and wiki always have a non-empty tenant_id even before workspace
// onboarding completes.
package tenancy

import (
	"context"
	"errors"
	"time"
)

// DefaultID is the reserved tenant id for system / pre-workspace activity.
// Mirrors OptimalEngine.Tenancy.Tenant.default_id/0 — DO NOT change without
// also updating audit.DefaultTenantID and the wiki facade.
const DefaultID = "default"

// Tenant is the canonical struct passed across subsystems. The `Plan` field
// gates feature visibility (free / pro / enterprise) and `Region` is reserved
// for residency-aware compliance workflows.
type Tenant struct {
	ID        string
	Name      string
	Plan      string
	Region    string
	OwnerID   string // BusinessOS workspaces.owner_id
	CreatedAt time.Time
	Metadata  map[string]any
}

// Store is the persistence contract.
type Store interface {
	Get(ctx context.Context, id string) (*Tenant, error)
	Create(ctx context.Context, t *Tenant) error
	List(ctx context.Context) ([]*Tenant, error)
	EnsureDefault(ctx context.Context) error
}

// ErrNotFound is returned when Get can't find the requested tenant.
var ErrNotFound = errors.New("tenancy: tenant not found")

// IsDefault reports whether id refers to the reserved system tenant.
func IsDefault(id string) bool { return id == DefaultID || id == "" }

// Resolve picks the effective tenant id: explicit when non-empty, fallback
// to DefaultID otherwise. Helps call sites avoid sprinkling DefaultID checks
// everywhere.
func Resolve(id string) string {
	if id == "" {
		return DefaultID
	}
	return id
}
