package wiki

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// Directives is the executable-directive layer of the Wiki.
//
// Wiki pages carry literal `{{verb: argument [key=value …]}}` tokens in the
// body. This file lexes those tokens into Directive structs, renders them
// into one of four output formats, and verifies the verb against a whitelist
// so no arbitrary execution is possible.

// Directive is one parsed token. Position is preserved so renderers can splice
// resolved content back into the body without re-parsing.
type Directive struct {
	Verb     string         // "cite" | "include" | "expand" | "search" | "table" | "trace" | "recent" | "wikilink"
	Argument string         // the URI / slug / query / etc.
	Options  map[string]any // any `key=value` pairs after the argument
	Raw      string         // the literal `{{...}}` (or `[[...]]`) substring
	Offset   int            // byte offset in the body
	Length   int            // byte length of Raw
}

// Whitelist is the set of verbs that are safe to render. Anything else is
// emitted verbatim with a warning so typos don't silently lose content.
var Whitelist = map[string]bool{
	"cite":    true,
	"include": true,
	"expand":  true,
	"search":  true,
	"table":   true,
	"trace":   true,
	"recent":  true,
}

var (
	directiveRe = regexp.MustCompile(`\{\{\s*([a-z_]+)\s*:\s*([^}]*?)\s*\}\}`)
	wikilinkRe  = regexp.MustCompile(`\[\[([^\]]+)\]\]`)
	optionRe    = regexp.MustCompile(`\s+([a-zA-Z_][a-zA-Z0-9_]*)=("(?:[^"]*)"|'(?:[^']*)'|\S+)`)
)

// ParseDirectives lexes a body into Directive tokens (curly form + wikilinks),
// in document order. Both regexes scan the full body; results are merged and
// re-sorted by offset so the renderer can splice from right to left safely.
func ParseDirectives(body string) []Directive {
	out := []Directive{}

	for _, m := range directiveRe.FindAllStringSubmatchIndex(body, -1) {
		full := body[m[0]:m[1]]
		verb := body[m[2]:m[3]]
		arg := body[m[4]:m[5]]
		argument, options := splitArgument(arg)
		out = append(out, Directive{
			Verb:     verb,
			Argument: argument,
			Options:  options,
			Raw:      full,
			Offset:   m[0],
			Length:   m[1] - m[0],
		})
	}

	for _, m := range wikilinkRe.FindAllStringSubmatchIndex(body, -1) {
		full := body[m[0]:m[1]]
		slug := body[m[2]:m[3]]
		out = append(out, Directive{
			Verb:     "wikilink",
			Argument: strings.TrimSpace(slug),
			Options:  map[string]any{},
			Raw:      full,
			Offset:   m[0],
			Length:   m[1] - m[0],
		})
	}

	sort.Slice(out, func(i, j int) bool { return out[i].Offset < out[j].Offset })
	return out
}

// splitArgument peels `key=value` options off the end of the argument string,
// matching the Elixir parser. Only the leading non-option text is the argument.
func splitArgument(s string) (string, map[string]any) {
	options := map[string]any{}
	matches := optionRe.FindAllStringSubmatchIndex(s, -1)
	if len(matches) == 0 {
		return strings.TrimSpace(s), options
	}
	first := matches[0][0]
	for _, m := range matches {
		key := s[m[2]:m[3]]
		val := s[m[4]:m[5]]
		val = strings.Trim(val, "\"'")
		options[key] = val
	}
	return strings.TrimSpace(s[:first]), options
}

// Format picks how the renderer composes resolved content.
type Format string

const (
	FormatPlain    Format = "plain"
	FormatMarkdown Format = "markdown"
	FormatClaude   Format = "claude"
	FormatOpenAI   Format = "openai"
)

// Resolved is what a Resolver returns. Content is the inline replacement;
// Source is an optional origin URI (used for footnotes and Claude/OpenAI
// metadata blocks).
type Resolved struct {
	Content string
	Source  string
}

// Resolver converts a directive into resolved content. The same resolver
// signature handles every verb — implementations switch on Directive.Verb.
type Resolver func(d Directive) (Resolved, error)

// RenderResult bundles the rewritten body with the citation list — the
// citation list lets callers build footnotes / Claude documents at the end
// of the rendered output.
type RenderResult struct {
	Body      string
	Citations []string
	Errors    []error
}

// Render walks every directive in body and replaces it with the Resolver's
// output. Format controls how `cite` is rendered:
//   - plain    → nothing left at the cite site, no footnotes
//   - markdown → `[^N]` footnote markers, footnotes appended to body
//   - claude   → leaves citations to be emitted as <document> blocks by caller
//   - openai   → leaves citations to be emitted as additional messages
//
// The renderer never panics; per-directive errors are collected in Errors
// and the offending directive is emitted verbatim with a `<!-- wiki:err -->`
// comment so callers can spot them.
func Render(body string, resolver Resolver, format Format) RenderResult {
	if format == "" {
		format = FormatMarkdown
	}
	directives := ParseDirectives(body)
	if len(directives) == 0 {
		return RenderResult{Body: body}
	}

	citations := []string{}
	citeIndex := map[string]int{}
	errs := []error{}

	// Right-to-left splice so offsets stay valid.
	rendered := body
	for i := len(directives) - 1; i >= 0; i-- {
		d := directives[i]
		if d.Verb != "wikilink" && !Whitelist[d.Verb] {
			rendered = splice(rendered, d.Offset, d.Length,
				fmt.Sprintf("<!-- wiki:warn unknown-verb=%s -->%s", d.Verb, d.Raw))
			errs = append(errs, fmt.Errorf("unknown verb %q at offset %d", d.Verb, d.Offset))
			continue
		}
		if resolver == nil {
			rendered = splice(rendered, d.Offset, d.Length, "")
			continue
		}
		resolved, err := resolver(d)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s %s: %w", d.Verb, d.Argument, err))
			rendered = splice(rendered, d.Offset, d.Length,
				fmt.Sprintf("<!-- wiki:err %s -->%s", err.Error(), d.Raw))
			continue
		}

		switch d.Verb {
		case "cite":
			source := resolved.Source
			if source == "" {
				source = d.Argument
			}
			n, exists := citeIndex[source]
			if !exists {
				citations = append(citations, source)
				n = len(citations)
				citeIndex[source] = n
			}
			rendered = splice(rendered, d.Offset, d.Length, citeMarker(format, n))
		default:
			rendered = splice(rendered, d.Offset, d.Length, resolved.Content)
		}
	}

	if format == FormatMarkdown && len(citations) > 0 {
		rendered = appendFootnotes(rendered, citations)
	}

	return RenderResult{Body: rendered, Citations: citations, Errors: errs}
}

// citeMarker renders a `cite` directive per format.
func citeMarker(format Format, n int) string {
	switch format {
	case FormatPlain:
		return ""
	case FormatClaude, FormatOpenAI:
		return fmt.Sprintf("[%d]", n)
	default:
		return fmt.Sprintf("[^%d]", n)
	}
}

func appendFootnotes(body string, citations []string) string {
	if len(citations) == 0 {
		return body
	}
	var sb strings.Builder
	sb.WriteString(body)
	sb.WriteString("\n\n")
	for i, src := range citations {
		fmt.Fprintf(&sb, "[^%d]: %s\n", i+1, src)
	}
	return sb.String()
}

func splice(s string, offset, length int, replacement string) string {
	if offset < 0 || offset+length > len(s) {
		return s
	}
	return s[:offset] + replacement + s[offset+length:]
}

// ErrUnknownVerb is returned when a Resolver receives a verb it can't service.
var ErrUnknownVerb = errors.New("wiki/directives: unknown verb")
