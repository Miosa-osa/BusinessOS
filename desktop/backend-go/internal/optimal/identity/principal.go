// Package identity ports OptimalEngine.Identity (Elixir) to Go.
//
// In BusinessOS principals bridge to better-auth's `user` table for human
// users; agents and service accounts are represented in their own
// `optimal_principals` table since BusinessOS has no native model for them.
// ACLs are recorded in `optimal_acls` and consulted by retrieval + curation.
package identity

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// Kind enumerates principal categories. Strings match the Elixir source's
// atom values byte-for-byte so cross-system audit messages stay identical.
type Kind string

const (
	KindUser    Kind = "user"
	KindAgent   Kind = "agent"
	KindService Kind = "service"
)

// SystemPrincipalID is the reserved actor for engine-internal events.
const SystemPrincipalID = "system"

// Principal is a unified identity used by audit, ACL, and retrieval.
// `ID` is the durable identifier; `ExternalID` references the upstream
// identity provider's record (better-auth user.id, an agent UUID, etc.).
type Principal struct {
	ID          string
	TenantID    string
	Kind        Kind
	DisplayName string
	ExternalID  string
	Metadata    map[string]any
}

// ErrPrincipalNotFound mirrors the Elixir `:not_found` return for Get.
var ErrPrincipalNotFound = errors.New("identity: principal not found")

// FormatID composes the conventional "kind:identifier" identity string used
// across logs and audit metadata. Empty kind falls back to KindUser.
func FormatID(kind Kind, id string) string {
	if kind == "" {
		kind = KindUser
	}
	return fmt.Sprintf("%s:%s", kind, id)
}

// ParseID inverts FormatID. Returns (KindUser, raw) when no prefix is present.
func ParseID(formatted string) (Kind, string) {
	parts := strings.SplitN(formatted, ":", 2)
	if len(parts) != 2 {
		return KindUser, formatted
	}
	return Kind(parts[0]), parts[1]
}

// PrincipalStore is the persistence contract.
type PrincipalStore interface {
	Upsert(ctx context.Context, p *Principal) error
	Get(ctx context.Context, id, tenantID string) (*Principal, error)
	List(ctx context.Context, tenantID string) ([]*Principal, error)
}
