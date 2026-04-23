package optimal

import (
	"strconv"
	"strings"
)

// FrontmatterData holds parsed YAML frontmatter from a markdown file.
// Handles both the nested signal: block written by ingest.go and the flat
// format used by older files.
type FrontmatterData struct {
	// Signal dimensions — from nested signal: block or flat keys.
	Mode      string  `json:"mode"`
	Genre     string  `json:"genre"`
	Type      string  `json:"type"`
	Format    string  `json:"format"`
	Structure string  `json:"structure"`
	SNRatio   float64 `json:"sn_ratio"`

	// Routing
	Node     string   `json:"node"`
	CrossRef []string `json:"cross_ref"`

	// Temporal
	ValidFrom  string `json:"valid_from"`
	ValidUntil string `json:"valid_until"`

	// Identity
	ID       string   `json:"id"`
	Title    string   `json:"title"`
	Entities []string `json:"entities"`
	Audience []string `json:"audience"`
	Intent   string   `json:"intent"`
}

// HasSignalDims reports whether any of the five core Signal Theory dimensions
// were populated by the parser. When true, callers may skip re-classification.
func (f FrontmatterData) HasSignalDims() bool {
	return f.Mode != "" || f.Genre != "" || f.Type != "" || f.Format != "" || f.Structure != ""
}

// ParseFrontmatter extracts YAML frontmatter from a markdown file. It returns
// the parsed data and the body text that follows the closing "---". If no
// frontmatter is found, an empty FrontmatterData and the full content are returned.
//
// Supported formats:
//
//	Nested signal: block (written by ingest.go):
//	  ---
//	  signal:
//	    mode: linguistic
//	    genre: transcript
//	    sn_ratio: 0.70
//	  node: 04
//	  cross_ref: 11, 06
//	  valid_from: "2026-04-13"
//	  ---
//
//	Flat format (older files):
//	  ---
//	  mode: linguistic
//	  genre: note
//	  type: inform
//	  ---
//
// Arrays are accepted as inline [a, b, c] or as YAML list items (  - value).
// No yaml library is imported; parsing is line-by-line matching the
// extractFrontmatterField pattern already in reader.go.
func ParseFrontmatter(content string) (FrontmatterData, string) {
	var data FrontmatterData

	// Must start with "---" (optionally followed by a newline).
	if !strings.HasPrefix(content, "---") {
		return data, content
	}

	// Find the closing "---" (must be on its own line after the opening).
	rest := content[3:]
	// Skip the newline after the opening delimiter.
	if strings.HasPrefix(rest, "\n") {
		rest = rest[1:]
	} else if strings.HasPrefix(rest, "\r\n") {
		rest = rest[2:]
	}

	end := strings.Index(rest, "\n---")
	if end < 0 {
		// No closing delimiter found — not valid frontmatter.
		return data, content
	}

	fmBlock := rest[:end]
	body := rest[end+4:] // skip past "\n---"
	// Trim the leading newline of the body.
	body = strings.TrimPrefix(body, "\r\n")
	body = strings.TrimPrefix(body, "\n")

	// Parse the frontmatter block line by line.
	lines := strings.Split(fmBlock, "\n")

	inSignalBlock := false // true while inside the nested "signal:" block

	for i := 0; i < len(lines); i++ {
		raw := lines[i]
		trimmed := strings.TrimSpace(raw)

		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Detect indented lines (inside a nested block).
		isIndented := len(raw) > 0 && (raw[0] == ' ' || raw[0] == '\t')

		// Entering or leaving the signal: nested block.
		if trimmed == "signal:" {
			inSignalBlock = true
			continue
		}
		// Any non-indented non-empty line ends the signal block.
		if inSignalBlock && !isIndented {
			inSignalBlock = false
		}

		key, val, ok := splitKV(trimmed)
		if !ok {
			continue
		}

		if inSignalBlock {
			// Keys inside signal: block.
			switch key {
			case "mode":
				data.Mode = unquote(val)
			case "genre":
				data.Genre = unquote(val)
			case "type":
				data.Type = unquote(val)
			case "format":
				data.Format = unquote(val)
			case "structure":
				data.Structure = unquote(val)
			case "sn_ratio":
				data.SNRatio = parseFloat(val)
			}
			continue
		}

		// Flat / top-level keys.
		switch key {
		case "mode":
			if data.Mode == "" {
				data.Mode = unquote(val)
			}
		case "genre":
			if data.Genre == "" {
				data.Genre = unquote(val)
			}
		case "type":
			if data.Type == "" {
				data.Type = unquote(val)
			}
		case "format":
			if data.Format == "" {
				data.Format = unquote(val)
			}
		case "structure":
			if data.Structure == "" {
				data.Structure = unquote(val)
			}
		case "sn_ratio":
			if data.SNRatio == 0 {
				data.SNRatio = parseFloat(val)
			}
		case "node":
			data.Node = unquote(val)
		case "cross_ref":
			if val != "" {
				data.CrossRef = parseInlineArray(val)
			} else {
				// Collect following indented list items.
				data.CrossRef, i = collectListItems(lines, i+1)
			}
		case "valid_from":
			data.ValidFrom = unquote(val)
		case "valid_until":
			data.ValidUntil = unquote(val)
		case "id":
			data.ID = unquote(val)
		case "title":
			data.Title = unquote(val)
		case "intent":
			data.Intent = unquote(val)
		case "entities":
			if val != "" {
				data.Entities = parseInlineArray(val)
			} else {
				data.Entities, i = collectListItems(lines, i+1)
			}
		case "audience":
			if val != "" {
				data.Audience = parseInlineArray(val)
			} else {
				data.Audience, i = collectListItems(lines, i+1)
			}
		}
	}

	return data, body
}

// ── helpers ───────────────────────────────────────────────────────────────────

// splitKV splits a trimmed YAML line into key and value at the first ": ".
// For keys with no value (e.g. "signal:") it returns key="signal", val="", ok=true.
func splitKV(line string) (key, val string, ok bool) {
	// Find first ":"
	idx := strings.IndexByte(line, ':')
	if idx < 0 {
		return "", "", false
	}
	key = strings.TrimSpace(line[:idx])
	if key == "" {
		return "", "", false
	}
	rest := line[idx+1:]
	val = strings.TrimSpace(rest)
	return key, val, true
}

// unquote strips surrounding double-quotes from a YAML scalar value.
func unquote(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	if len(s) >= 2 && s[0] == '\'' && s[len(s)-1] == '\'' {
		return s[1 : len(s)-1]
	}
	return s
}

// parseFloat converts a string like "0.70" or "0.7" to float64. Returns 0 on error.
func parseFloat(s string) float64 {
	s = strings.TrimSpace(s)
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return f
}

// parseInlineArray parses YAML inline array syntax  [a, b, c]  or a bare
// comma-separated string  a, b, c  into a slice of trimmed, unquoted strings.
func parseInlineArray(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	// Strip surrounding brackets.
	if len(s) >= 2 && s[0] == '[' && s[len(s)-1] == ']' {
		s = s[1 : len(s)-1]
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		v := unquote(strings.TrimSpace(p))
		if v != "" {
			result = append(result, v)
		}
	}
	return result
}

// collectListItems reads consecutive "  - value" lines starting at index start
// and returns them as a slice plus the last consumed index. Called when a key
// has an empty inline value and the list items are on subsequent lines.
func collectListItems(lines []string, start int) ([]string, int) {
	var result []string
	i := start
	for i < len(lines) {
		trimmed := strings.TrimSpace(lines[i])
		if strings.HasPrefix(trimmed, "- ") {
			val := unquote(strings.TrimSpace(trimmed[2:]))
			if val != "" {
				result = append(result, val)
			}
			i++
		} else {
			break
		}
	}
	// Return i-1 because the outer loop will call i++ after we return.
	if i > start {
		return result, i - 1
	}
	return result, start - 1
}
