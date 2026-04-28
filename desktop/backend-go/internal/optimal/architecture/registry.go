package architecture

import (
	"errors"
	"sort"
	"sync"
)

// Registry holds every known Architecture and Processor in one place.
// Built-in architectures and processor stubs register themselves via init();
// callers can register additional ones at runtime (e.g. tenant-defined
// schemas loaded from disk).
//
// The registry is intentionally split between architectures (data shapes)
// and processors (active components) so a tenant can swap the embedder
// without redefining the schema.
var (
	regMu          sync.RWMutex
	archByID       = map[string]*Architecture{}
	archByName     = map[string]*Architecture{}
	procByID       = map[string]Processor{}
	procByModality = map[Modality]Processor{} // first registered wins as default
)

// ErrArchNotFound / ErrProcessorNotFound are returned by Lookup* misses.
var (
	ErrArchNotFound      = errors.New("architecture: not found")
	ErrProcessorNotFound = errors.New("architecture: processor not found")
)

// RegisterArchitecture installs an Architecture by id (and name). Re-registering
// the same id replaces the previous binding (last wins) so hot-reload works.
func RegisterArchitecture(a *Architecture) {
	if a == nil {
		return
	}
	regMu.Lock()
	defer regMu.Unlock()
	archByID[a.ID] = a
	archByName[a.Name] = a
}

// RegisterProcessor installs a Processor under its ID and, when no default
// for its modality exists yet, sets it as the modality default.
func RegisterProcessor(p Processor) {
	if p == nil {
		return
	}
	regMu.Lock()
	defer regMu.Unlock()
	procByID[p.ID()] = p
	if _, ok := procByModality[p.Modality()]; !ok {
		procByModality[p.Modality()] = p
	}
}

// LookupArchitecture finds an architecture by ID OR name.
func LookupArchitecture(idOrName string) (*Architecture, error) {
	regMu.RLock()
	defer regMu.RUnlock()
	if a, ok := archByID[idOrName]; ok {
		return a, nil
	}
	if a, ok := archByName[idOrName]; ok {
		return a, nil
	}
	return nil, ErrArchNotFound
}

// LookupProcessor finds a processor by id.
func LookupProcessor(id string) (Processor, error) {
	regMu.RLock()
	defer regMu.RUnlock()
	if p, ok := procByID[id]; ok {
		return p, nil
	}
	return nil, ErrProcessorNotFound
}

// DefaultProcessorFor returns the modality's default processor, or nil if
// none registered. Used by the field router when a Field has no processor
// hint.
func DefaultProcessorFor(m Modality) Processor {
	regMu.RLock()
	defer regMu.RUnlock()
	return procByModality[m]
}

// Architectures returns every registered architecture, sorted by id.
func Architectures() []*Architecture {
	regMu.RLock()
	defer regMu.RUnlock()
	out := make([]*Architecture, 0, len(archByID))
	for _, a := range archByID {
		out = append(out, a)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// Processors returns every registered processor, sorted by id.
func Processors() []Processor {
	regMu.RLock()
	defer regMu.RUnlock()
	out := make([]Processor, 0, len(procByID))
	for _, p := range procByID {
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID() < out[j].ID() })
	return out
}
