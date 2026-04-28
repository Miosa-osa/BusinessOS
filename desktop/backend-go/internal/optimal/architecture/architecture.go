package architecture

import (
	"errors"
	"fmt"
)

// Granularity is one decomposition scale a piece of content gets sliced at.
// A text_signal is decomposed at document / section / paragraph / sentence;
// a code_commit is decomposed at commit / file / hunk / line.
type Granularity string

// Architecture is the typed schema for one class of data point. Multiple
// fields, multiple granularities, one canonical id — model-agnostic.
type Architecture struct {
	ID              string
	Name            string
	Version         int
	Description     string
	ModalityPrimary Modality
	Fields          []Field
	Granularity     []Granularity
	Retention       string
	Metadata        map[string]any
}

// New builds an Architecture with sensible defaults. Required fields:
// Name, ModalityPrimary. Version defaults to 1; ID defaults to
// "<name>.v<version>".
func New(name string, primary Modality, opts ...Option) (*Architecture, error) {
	if name == "" {
		return nil, errors.New("architecture: name is required")
	}
	a := &Architecture{
		Name:            name,
		Version:         1,
		ModalityPrimary: primary,
		Granularity:     []Granularity{"document"},
		Retention:       "default",
		Metadata:        map[string]any{},
	}
	for _, opt := range opts {
		opt(a)
	}
	if a.ID == "" {
		a.ID = fmt.Sprintf("%s.v%d", a.Name, a.Version)
	}
	return a, nil
}

// Option configures an Architecture.
type Option func(*Architecture)

// WithVersion overrides the version (default 1). Bumping the version forces
// a new ID so downstream caches don't blend old + new specs.
func WithVersion(v int) Option { return func(a *Architecture) { a.Version = v } }

// WithDescription sets the human-readable doc string.
func WithDescription(d string) Option { return func(a *Architecture) { a.Description = d } }

// WithFields appends fields in declaration order.
func WithFields(fields ...Field) Option {
	return func(a *Architecture) { a.Fields = append(a.Fields, fields...) }
}

// WithGranularity overrides the default [document] granularity list.
func WithGranularity(g ...Granularity) Option {
	return func(a *Architecture) { a.Granularity = g }
}

// WithRetention sets the retention class ("default", "compliance",
// "ephemeral", …) — the actual TTL math lives in compliance/retention.
func WithRetention(r string) Option { return func(a *Architecture) { a.Retention = r } }

// WithMetadata merges arbitrary metadata into the Architecture.
func WithMetadata(m map[string]any) Option {
	return func(a *Architecture) {
		for k, v := range m {
			a.Metadata[k] = v
		}
	}
}

// Field looks up a field by name, returning nil when absent.
func (a *Architecture) Field(name string) *Field {
	for i := range a.Fields {
		if a.Fields[i].Name == name {
			return &a.Fields[i]
		}
	}
	return nil
}

// RequiredFields returns the names of fields marked Required.
func (a *Architecture) RequiredFields() []string {
	out := []string{}
	for _, f := range a.Fields {
		if f.Required {
			out = append(out, f.Name)
		}
	}
	return out
}

// ToSpec produces the JSON-friendly map the Elixir source persists in
// `architectures.spec`. Useful for storing the architecture definition
// alongside the rows produced under it.
func (a *Architecture) ToSpec() map[string]any {
	fields := make([]map[string]any, 0, len(a.Fields))
	for _, f := range a.Fields {
		fmap := map[string]any{
			"name":        f.Name,
			"modality":    string(f.Modality),
			"required":    f.Required,
			"description": f.Description,
		}
		if f.Processor != "" {
			fmap["processor"] = f.Processor
		}
		if len(f.Dims) > 0 {
			dims := make([]any, len(f.Dims))
			for i, d := range f.Dims {
				if d == DimAny {
					dims[i] = "any"
				} else {
					dims[i] = int(d)
				}
			}
			fmap["dims"] = dims
		}
		fields = append(fields, fmap)
	}
	gran := make([]string, len(a.Granularity))
	for i, g := range a.Granularity {
		gran[i] = string(g)
	}
	return map[string]any{
		"id":               a.ID,
		"name":             a.Name,
		"version":          a.Version,
		"description":      a.Description,
		"modality_primary": string(a.ModalityPrimary),
		"granularity":      gran,
		"retention":        a.Retention,
		"fields":           fields,
		"metadata":         a.Metadata,
	}
}
