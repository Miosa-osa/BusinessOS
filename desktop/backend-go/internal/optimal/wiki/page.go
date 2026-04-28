// Package wiki ports OptimalEngine.Wiki (Elixir) to Go.
//
// In BusinessOS the wiki layer is delivered through the existing Pages module
// (contexts table). This package provides the wiki API surface — Page struct,
// directives, integrity, citations — and the storage backend bridges to the
// pages tables instead of a parallel wiki_pages store. See
// store_pages_adapter.go for the bridge.
package wiki

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Page is the atomic unit of Tier-3 (the LLM-curated wiki).
//
// Mirrors the Elixir struct except `version` is monotonic per (tenant_id,
// slug, audience). The `Body` carries unresolved directive syntax —
// `{{cite: optimal://...}}`, `{{include: ...}}` — rendered at READ time
// via Directives.Render.
type Page struct {
	TenantID    string
	Slug        string
	Audience    string
	Version     int
	Frontmatter map[string]any
	Body        string
	LastCurated string // RFC3339; empty = never
	CuratedBy   string
}

// DefaultAudience is the per-page default when callers don't specialize.
const DefaultAudience = "default"

// ErrMissingSlug is returned when FromMarkdown can't derive a slug.
var ErrMissingSlug = errors.New("wiki: page is missing a slug")

// frontmatterRe matches a YAML-ish frontmatter block at the very top of the
// file. Identical capture behaviour to the Elixir source's permissive parser.
var frontmatterRe = regexp.MustCompile(`(?s)\A---\r?\n(.*?)\r?\n---\r?\n?(.*)\z`)

// FromMarkdown parses a markdown document (optionally with `---` frontmatter)
// into a Page. Frontmatter is parsed permissively — plain `key: value` pairs.
//
// Options propagate via the opts map: tenant_id, slug, audience override
// whatever the frontmatter declares.
func FromMarkdown(markdown string, opts map[string]string) (*Page, error) {
	frontmatter, body := parseFrontmatter(markdown)

	page := &Page{
		TenantID:    coalesce(opts["tenant_id"], stringFromAny(frontmatter["tenant_id"]), "default"),
		Slug:        coalesce(opts["slug"], stringFromAny(frontmatter["slug"])),
		Audience:    coalesce(opts["audience"], stringFromAny(frontmatter["audience"]), DefaultAudience),
		Version:     intFromAny(frontmatter["version"], 1),
		Frontmatter: frontmatter,
		Body:        body,
		LastCurated: stringFromAny(frontmatter["last_curated"]),
		CuratedBy:   stringFromAny(frontmatter["curated_by"]),
	}
	if page.Slug == "" {
		return nil, ErrMissingSlug
	}
	return page, nil
}

// ToMarkdown serializes a Page back to a frontmatter + body markdown string.
// The frontmatter is emitted with stable key ordering so two equivalent
// pages produce byte-identical output (good for diffs and content hashing).
func (p *Page) ToMarkdown() string {
	if p.Frontmatter == nil {
		p.Frontmatter = map[string]any{}
	}
	fm := cloneMap(p.Frontmatter)
	fm["slug"] = p.Slug
	fm["audience"] = p.Audience
	fm["version"] = p.Version
	fm["tenant_id"] = p.TenantID
	if p.LastCurated != "" {
		fm["last_curated"] = p.LastCurated
	}
	if p.CuratedBy != "" {
		fm["curated_by"] = p.CuratedBy
	}

	var sb strings.Builder
	sb.WriteString("---\n")
	sb.WriteString(renderFrontmatter(fm))
	sb.WriteString("---\n\n")
	sb.WriteString(p.Body)
	return sb.String()
}

// NewPage builds an empty Page with reasonable defaults. Most call sites
// should prefer FromMarkdown when ingesting external content.
func NewPage(tenantID, slug, audience string) *Page {
	if audience == "" {
		audience = DefaultAudience
	}
	return &Page{
		TenantID:    tenantID,
		Slug:        slug,
		Audience:    audience,
		Version:     1,
		Frontmatter: map[string]any{},
	}
}

// parseFrontmatter splits a markdown doc into (frontmatter map, body string).
// Missing frontmatter returns an empty map and the original body.
func parseFrontmatter(md string) (map[string]any, string) {
	m := frontmatterRe.FindStringSubmatch(md)
	if m == nil {
		return map[string]any{}, md
	}
	return parseYAMLish(m[1]), m[2]
}

// parseYAMLish handles `key: value` lines and `key: [a, b, c]` lists. Anything
// fancier degrades to the raw string. This matches the Elixir permissive
// parser; complex YAML would require a dedicated dependency we don't need
// for the wiki frontmatter shapes that exist today.
func parseYAMLish(text string) map[string]any {
	out := map[string]any{}
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.Index(line, ":")
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		out[key] = parseYAMLValue(val)
	}
	return out
}

func parseYAMLValue(v string) any {
	if v == "" {
		return ""
	}
	if strings.HasPrefix(v, "[") && strings.HasSuffix(v, "]") {
		inner := strings.TrimSpace(v[1 : len(v)-1])
		if inner == "" {
			return []any{}
		}
		parts := strings.Split(inner, ",")
		out := make([]any, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			p = strings.Trim(p, "\"'")
			out = append(out, p)
		}
		return out
	}
	if v == "true" {
		return true
	}
	if v == "false" {
		return false
	}
	if n, err := strconv.Atoi(v); err == nil {
		return n
	}
	v = strings.Trim(v, "\"'")
	return v
}

func renderFrontmatter(fm map[string]any) string {
	keys := make([]string, 0, len(fm))
	for k := range fm {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var sb strings.Builder
	for _, k := range keys {
		fmt.Fprintf(&sb, "%s: %s\n", k, renderFrontmatterValue(fm[k]))
	}
	return sb.String()
}

func renderFrontmatterValue(v any) string {
	switch x := v.(type) {
	case string:
		if needsQuote(x) {
			return strconv.Quote(x)
		}
		return x
	case []any:
		parts := make([]string, len(x))
		for i, p := range x {
			parts[i] = renderFrontmatterValue(p)
		}
		return "[" + strings.Join(parts, ", ") + "]"
	default:
		return fmt.Sprintf("%v", v)
	}
}

func needsQuote(s string) bool {
	for _, r := range s {
		if r == ':' || r == '[' || r == ']' || r == '"' || r == '\'' {
			return true
		}
	}
	return false
}

func coalesce(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func stringFromAny(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

func intFromAny(v any, def int) int {
	switch x := v.(type) {
	case int:
		return x
	case int64:
		return int(x)
	case float64:
		return int(x)
	case string:
		if n, err := strconv.Atoi(x); err == nil {
			return n
		}
	}
	return def
}

func cloneMap(m map[string]any) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}
