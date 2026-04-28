package tenancy

import (
	"context"
	"errors"
	"testing"
)

func TestResolveDefaults(t *testing.T) {
	if Resolve("") != DefaultID {
		t.Fatal("empty should resolve to default")
	}
	if Resolve("tenant-x") != "tenant-x" {
		t.Fatal("non-empty should pass through")
	}
}

func TestIsDefault(t *testing.T) {
	if !IsDefault("") || !IsDefault(DefaultID) {
		t.Fatal("empty and DefaultID must be considered default")
	}
	if IsDefault("tenant-1") {
		t.Fatal("a real tenant id is not default")
	}
}

func TestMemoryStoreEnsureDefaultIsIdempotent(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()
	if err := s.EnsureDefault(ctx); err != nil {
		t.Fatal(err)
	}
	if err := s.EnsureDefault(ctx); err != nil {
		t.Fatal(err)
	}
	got, err := s.Get(ctx, DefaultID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Plan != "system" {
		t.Fatalf("default plan: %q", got.Plan)
	}
}

func TestMemoryStoreCreateAndGet(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()
	tenant := &Tenant{ID: "tenant-A", Name: "Alpha", Plan: "pro", OwnerID: "user:alice"}
	if err := s.Create(ctx, tenant); err != nil {
		t.Fatal(err)
	}
	got, err := s.Get(ctx, "tenant-A")
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Alpha" || got.Plan != "pro" {
		t.Fatalf("round-trip: %+v", got)
	}
}

func TestMemoryStoreGetNotFound(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()
	if _, err := s.Get(ctx, "nope"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
