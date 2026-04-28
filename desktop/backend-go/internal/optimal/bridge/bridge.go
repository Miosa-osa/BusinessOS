// Package bridge ports OptimalEngine.Bridge.* (Elixir) to Go.
//
// The bridge layer is the thin glue between OptimalEngine subsystems
// (knowledge graph, memory, signal). Each Elixir bridge module is one file
// here. The bridges are deliberately stateless — they delegate to the
// subsystem packages we already shipped in Waves 1-2.
//
// Why three submodules instead of three packages: the Elixir source treats
// these as one logical layer. Keeping them in the same Go package lets a
// caller `import "...optimal/bridge"` and call Bridge.GraphBoost,
// Bridge.RecordEvent, Bridge.AuditSignal without juggling imports.
package bridge
