package identity

import (
	"context"
	"errors"
	"testing"
)

func TestFormatAndParseID(t *testing.T) {
	id := FormatID(KindAgent, "alpha")
	if id != "agent:alpha" {
		t.Fatalf("format: %q", id)
	}
	kind, raw := ParseID(id)
	if kind != KindAgent || raw != "alpha" {
		t.Fatalf("parse: %s %s", kind, raw)
	}
	// Bare values default to KindUser.
	k, r := ParseID("ada")
	if k != KindUser || r != "ada" {
		t.Fatalf("default kind: %s %s", k, r)
	}
}

func TestPrincipalUpsertAndGet(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryPrincipalStore()
	p := &Principal{
		ID:          "agent:alpha",
		TenantID:    "tenant-1",
		Kind:        KindAgent,
		DisplayName: "Alpha Agent",
	}
	if err := store.Upsert(ctx, p); err != nil {
		t.Fatal(err)
	}
	got, err := store.Get(ctx, "agent:alpha", "tenant-1")
	if err != nil {
		t.Fatal(err)
	}
	if got.DisplayName != "Alpha Agent" {
		t.Fatalf("display: %q", got.DisplayName)
	}
}

func TestPrincipalCrossTenantIsolation(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryPrincipalStore()
	store.Upsert(ctx, &Principal{ID: "agent:a", TenantID: "tenant-1", Kind: KindAgent, DisplayName: "A"})
	if _, err := store.Get(ctx, "agent:a", "tenant-2"); !errors.Is(err, ErrPrincipalNotFound) {
		t.Fatalf("cross-tenant access should fail, got %v", err)
	}
}

func TestACLValidatePrincipalXorGroup(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryACLStore()
	bad := &ACL{ResourceURI: "wiki:foo:default", PrincipalID: "user:a", GroupID: "g1", Permission: PermRead}
	if err := store.Grant(ctx, bad); !errors.Is(err, ErrACLConflict) {
		t.Fatalf("expected conflict, got %v", err)
	}
	neither := &ACL{ResourceURI: "wiki:foo:default", Permission: PermRead}
	if err := store.Grant(ctx, neither); !errors.Is(err, ErrACLConflict) {
		t.Fatalf("expected conflict for neither, got %v", err)
	}
}

func TestACLPermissiveDefault(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryACLStore()
	// No ACLs for the resource → permissive default returns true.
	ok, err := store.Can(ctx, "user:a", "wiki:no-acls:default", PermRead, CanOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("permissive default should allow read when no ACLs exist")
	}
}

func TestACLDenyWithoutGrant(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryACLStore()
	store.Grant(ctx, &ACL{ResourceURI: "wiki:secret:default", PrincipalID: "user:owner", Permission: PermAdmin})
	ok, err := store.Can(ctx, "user:other", "wiki:secret:default", PermRead, CanOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("non-granted user should be denied once ACLs exist")
	}
}

func TestACLAdminImpliesWriteAndRead(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryACLStore()
	store.Grant(ctx, &ACL{ResourceURI: "wiki:doc:default", PrincipalID: "user:owner", Permission: PermAdmin})
	for _, want := range []Permission{PermRead, PermWrite, PermAdmin} {
		ok, err := store.Can(ctx, "user:owner", "wiki:doc:default", want, CanOptions{})
		if err != nil || !ok {
			t.Fatalf("admin should imply %s: ok=%v err=%v", want, ok, err)
		}
	}
}

func TestACLGroupGrant(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryACLStore()
	store.Grant(ctx, &ACL{ResourceURI: "wiki:eng:default", GroupID: "engineers", Permission: PermWrite})
	// Member of group should pass.
	ok, _ := store.Can(ctx, "user:dev", "wiki:eng:default", PermRead, CanOptions{Groups: []string{"engineers"}})
	if !ok {
		t.Fatal("group member should be granted read via write ACL")
	}
	// Non-member should fail.
	ok, _ = store.Can(ctx, "user:other", "wiki:eng:default", PermRead, CanOptions{Groups: []string{"sales"}})
	if ok {
		t.Fatal("non-member should be denied")
	}
}

func TestACLRevoke(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryACLStore()
	store.Grant(ctx, &ACL{ResourceURI: "wiki:x:default", PrincipalID: "user:a", Permission: PermWrite})
	if err := store.Revoke(ctx, "", "wiki:x:default", "user:a", "", PermWrite); err != nil {
		t.Fatal(err)
	}
	// After revoke + with no other ACLs, permissive default kicks back in.
	ok, _ := store.Can(ctx, "user:a", "wiki:x:default", PermRead, CanOptions{})
	if !ok {
		t.Fatal("after final revoke, permissive default should allow read")
	}
}
