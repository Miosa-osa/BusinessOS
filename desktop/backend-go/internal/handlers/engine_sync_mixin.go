package handlers

import (
	"context"

	"github.com/rhl/businessos-backend/internal/services"
)

// EngineSyncHook is an embeddable mixin every BusinessOS module handler
// can include to gain a one-line connection to the OptimalEngine
// knowledge graph. Embedding the struct gives the handler:
//
//	h.SetEngineSync(s)             — called once at registration
//	h.enqueue(ctx, signal)         — fire-and-forget Enqueue
//	h.enqueueDelete(ctx, mod, id)  — same shape for deletes
//
// The mixin keeps every per-module CRUD file tiny: one struct field
// (the embed) + one or two lines per write/delete handler.
//
// All methods are nil-safe — when the engine isn't configured (no
// OPTIMAL_NODES_ROOT/OPTIMAL_DB_PATH), enqueue is a no-op so module
// CRUD continues to work without an engine wired.
type EngineSyncHook struct {
	engineSync *services.EngineSync
}

// SetEngineSync wires the shared service. Called once from route
// registration right after the handler is constructed.
func (h *EngineSyncHook) SetEngineSync(s *services.EngineSync) {
	h.engineSync = s
}

// enqueue mirrors a write into the engine. Passes through silently when
// the sync service isn't wired so modules don't have to nil-check.
func (h *EngineSyncHook) enqueue(ctx context.Context, sig services.Signal) {
	if h == nil || h.engineSync == nil {
		return
	}
	h.engineSync.Enqueue(ctx, sig)
}

// enqueueDelete removes a module's entity from the graph. Same nil-safety
// as enqueue.
func (h *EngineSyncHook) enqueueDelete(ctx context.Context, mod services.Module, id string) {
	if h == nil || h.engineSync == nil {
		return
	}
	h.engineSync.EnqueueDelete(ctx, mod, id)
}
