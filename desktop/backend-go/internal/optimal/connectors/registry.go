package connectors

import (
	"errors"
	"sort"
	"sync"
)

// Registry is the process-wide map of connector kind → Connector.
// Adapters call Register() from an init() func; the facade calls Lookup().
var (
	registryMu sync.RWMutex
	registry   = map[string]Connector{}
)

// ErrAdapterNotFound is returned when Lookup can't find a registered kind.
var ErrAdapterNotFound = errors.New("connectors: adapter not registered")

// Register installs an adapter under its Kind(). Safe to call from init().
// Re-registering the same Kind replaces the previous binding — the last one
// wins, which matches how Elixir modules are recompiled and re-loaded.
func Register(c Connector) {
	registryMu.Lock()
	defer registryMu.Unlock()
	registry[c.Kind()] = c
}

// Lookup fetches an adapter by kind. Returns ErrAdapterNotFound on miss.
func Lookup(kind string) (Connector, error) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	if c, ok := registry[kind]; ok {
		return c, nil
	}
	return nil, ErrAdapterNotFound
}

// Summary is one entry of the list every UI and mix task needs.
type Summary struct {
	Kind        string
	DisplayName string
	AuthScheme  AuthScheme
}

// List returns one Summary per registered adapter, sorted by Kind for
// deterministic output.
func List() []Summary {
	registryMu.RLock()
	defer registryMu.RUnlock()
	out := make([]Summary, 0, len(registry))
	for _, c := range registry {
		out = append(out, Summary{Kind: c.Kind(), DisplayName: c.DisplayName(), AuthScheme: c.AuthScheme()})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Kind < out[j].Kind })
	return out
}

// Kinds is the bare list of registered adapter kinds, sorted.
func Kinds() []string {
	sums := List()
	out := make([]string, len(sums))
	for i, s := range sums {
		out[i] = s.Kind
	}
	return out
}
