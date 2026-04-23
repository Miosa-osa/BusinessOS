package optimal

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// TopologyConfig is the full routing and node topology loaded from config.yaml.
type TopologyConfig struct {
	Nodes  map[string]NodeConfig   `yaml:"nodes"  json:"nodes"`
	People map[string]PersonConfig `yaml:"people" json:"people"`
	Routes []RoutingRule           `yaml:"routes" json:"routes"`
}

// NodeConfig describes a single numbered node folder.
type NodeConfig struct {
	Name     string   `yaml:"name"      json:"name"`
	Type     string   `yaml:"type"      json:"type"`
	CrossRef []string `yaml:"cross_ref" json:"cross_ref"`
}

// PersonConfig describes a person's preferred communication style.
type PersonConfig struct {
	Role    string `yaml:"role"    json:"role"`
	Genre   string `yaml:"genre"   json:"genre"`   // preferred output genre
	Channel string `yaml:"channel" json:"channel"` // preferred delivery channel
}

// RoutingRule maps trigger keywords to a target node (and optional cross-references).
type RoutingRule struct {
	Keywords []string `yaml:"keywords" json:"keywords"`
	Target   string   `yaml:"target"   json:"target"` // node slug
	CrossRef []string `yaml:"cross_ref" json:"cross_ref"`
}

// LoadTopology reads and parses a YAML topology config at configPath.
// Returns a non-nil *TopologyConfig (with empty maps) if the file is absent,
// so callers can always dereference the result safely.
func LoadTopology(configPath string) (*TopologyConfig, error) {
	cfg := &TopologyConfig{
		Nodes:  make(map[string]NodeConfig),
		People: make(map[string]PersonConfig),
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default empty topology — not an error.
			return cfg, nil
		}
		return nil, fmt.Errorf("optimal/topology: read config %q: %w", configPath, err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("optimal/topology: parse config %q: %w", configPath, err)
	}

	// Ensure maps are not nil after unmarshal.
	if cfg.Nodes == nil {
		cfg.Nodes = make(map[string]NodeConfig)
	}
	if cfg.People == nil {
		cfg.People = make(map[string]PersonConfig)
	}

	return cfg, nil
}

// ToInternalRules converts the YAML RoutingRule slice into the internal
// routingRule type used by RouteWithRules. Returns nil when there are no routes
// so callers can fall back to defaultRoutingRules with a simple nil check.
func (t *TopologyConfig) ToInternalRules() []routingRule {
	if len(t.Routes) == 0 {
		return nil
	}
	out := make([]routingRule, len(t.Routes))
	for i, r := range t.Routes {
		out[i] = routingRule{
			keywords:  r.Keywords,
			primary:   r.Target,
			crossRefs: r.CrossRef,
		}
	}
	return out
}

// GetPersonConfig returns the PersonConfig for a given name (case-insensitive).
// The lookup matches against People map keys first, then against Role values.
// Returns nil if not found.
func (t *TopologyConfig) GetPersonConfig(name string) *PersonConfig {
	lower := toLower(name)
	for key, pc := range t.People {
		if toLower(key) == lower || toLower(pc.Role) == lower {
			// Copy to avoid returning a pointer into a map value (map iteration
			// aliasing — the value would change on the next iteration).
			copy := pc
			return &copy
		}
	}
	return nil
}

// GetPersonGenre returns the preferred output genre for the named person.
// The lookup is case-insensitive against People map keys. Returns an empty
// string when the person is not found.
func (t *TopologyConfig) GetPersonGenre(name string) string {
	lname := toLower(name)
	for key, p := range t.People {
		if toLower(key) == lname {
			return p.Genre
		}
	}
	return ""
}

// RouteQuery returns the RoutingRule whose keywords best match query (case-insensitive).
// Returns nil when no rule matches.
func (t *TopologyConfig) RouteQuery(query string) *RoutingRule {
	lower := toLower(query)
	for i := range t.Routes {
		rule := &t.Routes[i]
		for _, kw := range rule.Keywords {
			if containsWord(lower, toLower(kw)) {
				return rule
			}
		}
	}
	return nil
}

// toLower is a package-level helper (strings.ToLower alias) to avoid import cycle.
func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 32
		}
		result[i] = c
	}
	return string(result)
}

// containsWord reports whether haystack contains needle as a substring.
func containsWord(haystack, needle string) bool {
	if needle == "" {
		return false
	}
	return len(haystack) >= len(needle) && indexOf(haystack, needle) >= 0
}

func indexOf(s, sub string) int {
	if len(sub) == 0 {
		return 0
	}
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
