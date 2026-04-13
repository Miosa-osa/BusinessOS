package optimal

import (
	"fmt"
	"strings"
)

// Tier constants for LoadTiered.
const (
	TierL0   = "l0"   // ~100 tokens: title + first paragraph
	TierL1   = "l1"   // ~2K tokens: all headings + first paragraph of each section
	TierFull = "full" // complete file content
)

// L0Abstract returns a ~100-token summary of a markdown document.
// It extracts the document title (from frontmatter or first heading) and the
// first substantive paragraph — enough for the agent to decide whether to read
// more without loading the full file.
func L0Abstract(content string) string {
	content = strings.TrimSpace(content)
	if content == "" {
		return ""
	}

	// Strip YAML frontmatter if present.
	body := stripFrontmatter(content)

	// Extract the first heading as title.
	title := extractFirstHeading(body)

	// Extract the first non-empty paragraph (sequence of non-blank lines).
	firstPara := extractFirstParagraph(body)

	var sb strings.Builder
	if title != "" {
		sb.WriteString("# ")
		sb.WriteString(title)
		sb.WriteString("\n\n")
	}
	if firstPara != "" {
		sb.WriteString(firstPara)
	}

	// Clamp to ~400 characters (~100 tokens at 4 chars/token).
	out := sb.String()
	if len(out) > 400 {
		out = out[:400]
		// Snap back to the last word boundary.
		if idx := strings.LastIndexAny(out, " \n"); idx > 300 {
			out = out[:idx]
		}
		out += "..."
	}
	return strings.TrimSpace(out)
}

// L1Overview returns a ~2K-token overview of a markdown document.
// For each ## or ### section heading it includes the heading line and the
// first paragraph of that section. The result is capped at ~8000 characters
// (~2000 tokens). Use this tier when L0 signals the document is relevant.
func L1Overview(content string) string {
	content = strings.TrimSpace(content)
	if content == "" {
		return ""
	}

	body := stripFrontmatter(content)
	lines := strings.Split(body, "\n")

	var sb strings.Builder
	i := 0
	sectionCount := 0

	for i < len(lines) {
		line := lines[i]

		// Detect a heading (# ## ###).
		if isHeading(line) {
			if sectionCount > 0 {
				sb.WriteString("\n")
			}
			sb.WriteString(line)
			sb.WriteString("\n")
			i++
			sectionCount++

			// Collect the first non-blank paragraph after the heading.
			// Skip blank lines immediately after the heading.
			for i < len(lines) && strings.TrimSpace(lines[i]) == "" {
				i++
			}
			// Collect paragraph lines until a blank line or next heading.
			var para []string
			for i < len(lines) {
				l := lines[i]
				if strings.TrimSpace(l) == "" || isHeading(l) {
					break
				}
				para = append(para, l)
				i++
			}
			if len(para) > 0 {
				sb.WriteString(strings.Join(para, "\n"))
				sb.WriteString("\n")
			}
			continue
		}

		// Before the first heading: include as-is (preamble / intro).
		if sectionCount == 0 && strings.TrimSpace(line) != "" {
			sb.WriteString(line)
			sb.WriteString("\n")
		}
		i++
	}

	// Cap at ~8000 chars (~2000 tokens).
	out := sb.String()
	if len(out) > 8000 {
		out = out[:8000]
		if idx := strings.LastIndex(out, "\n"); idx > 7000 {
			out = out[:idx]
		}
		out += "\n..."
	}
	return strings.TrimSpace(out)
}

// L2Full returns the complete file content unmodified.
func L2Full(content string) string {
	return content
}

// LoadTiered reads a file from inside a node at the requested tier depth.
//
//   - tier "l0"   → L0Abstract  (~100 tokens)
//   - tier "l1"   → L1Overview  (~2K tokens)
//   - any other   → full content
//
// The slug and filePath arguments are forwarded to GetNodeFile — filePath is
// relative to the node folder (e.g. "context.md" or "signals/2026-04-01-foo.md").
func LoadTiered(nodesRoot, slug, filePath, tier string) (string, error) {
	content, err := GetNodeFile(nodesRoot, slug, filePath)
	if err != nil {
		return "", fmt.Errorf("optimal/tiered: read %s/%s: %w", slug, filePath, err)
	}
	switch strings.ToLower(tier) {
	case TierL0:
		return L0Abstract(content), nil
	case TierL1:
		return L1Overview(content), nil
	default:
		return L2Full(content), nil
	}
}

// ── internal helpers ──────────────────────────────────────────────────────────

// stripFrontmatter removes the leading YAML block (--- ... ---) if present.
func stripFrontmatter(content string) string {
	if !strings.HasPrefix(content, "---") {
		return content
	}
	rest := content[3:]
	end := strings.Index(rest, "\n---")
	if end < 0 {
		return content
	}
	// Skip the closing --- line.
	after := rest[end+4:]
	return strings.TrimLeft(after, "\n")
}

// extractFirstHeading returns the text of the first # heading in body (without
// the leading # characters and surrounding whitespace). Returns "" if none found.
func extractFirstHeading(body string) string {
	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimLeft(line, "#")
		if len(trimmed) < len(line) { // at least one # was removed
			return strings.TrimSpace(trimmed)
		}
	}
	return ""
}

// extractFirstParagraph returns the first block of consecutive non-blank lines.
func extractFirstParagraph(body string) string {
	lines := strings.Split(body, "\n")
	var para []string
	inPara := false

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			if inPara {
				break // end of first paragraph
			}
			continue
		}
		// Skip heading lines for the first-paragraph search.
		if isHeading(line) {
			if inPara {
				break
			}
			continue
		}
		inPara = true
		para = append(para, line)
	}
	return strings.Join(para, "\n")
}

// isHeading reports whether line starts with one or more '#' characters.
func isHeading(line string) bool {
	return len(line) > 0 && line[0] == '#'
}
