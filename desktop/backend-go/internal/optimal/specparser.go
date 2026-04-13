package optimal

import (
	"fmt"
	"regexp"
	"strings"
)

// Spec is the parsed representation of a .spec.md file.
type Spec struct {
	Name         string        `json:"name"`
	Version      string        `json:"version"`
	Requirements []Requirement `json:"requirements"`
	Status       string        `json:"status"` // draft, active, deprecated
}

// Requirement is a single numbered requirement extracted from a spec file.
type Requirement struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Priority    string `json:"priority"` // must, should, may
	Status      string `json:"status"`   // pending, verified, failed
}

// ParseSpec reads a .spec.md file content string and extracts the spec metadata
// and requirements. YAML frontmatter provides name/version/status. Requirements
// are parsed from a "## Requirements" section where each line matches:
//
//   - MUST: <description>
//   - SHOULD: <description>
//   - MAY: <description>
//
// The requirement ID is auto-generated as REQ-001, REQ-002, etc.
func ParseSpec(content string) (*Spec, error) {
	if strings.TrimSpace(content) == "" {
		return nil, fmt.Errorf("optimal/specparser: content is empty")
	}

	spec := &Spec{
		Status: "draft",
	}

	// ── parse YAML frontmatter ────────────────────────────────────────────────
	body := content
	if strings.HasPrefix(strings.TrimSpace(content), "---") {
		fm, rest, err := splitFrontmatter(content)
		if err == nil {
			body = rest
			parseFrontmatterFields(fm, spec)
		}
	}

	// ── find requirements section ─────────────────────────────────────────────
	spec.Requirements = parseRequirements(body)

	return spec, nil
}

// ── frontmatter helpers ───────────────────────────────────────────────────────

var (
	reFMName    = regexp.MustCompile(`(?m)^name:\s*(.+)$`)
	reFMVersion = regexp.MustCompile(`(?m)^version:\s*(.+)$`)
	reFMStatus  = regexp.MustCompile(`(?m)^status:\s*(.+)$`)
)

func splitFrontmatter(content string) (fm, body string, err error) {
	content = strings.TrimLeft(content, "\n")
	if !strings.HasPrefix(content, "---") {
		return "", content, fmt.Errorf("no frontmatter")
	}
	rest := content[3:]
	// Handle optional newline after opening ---
	if len(rest) > 0 && rest[0] == '\n' {
		rest = rest[1:]
	}
	end := strings.Index(rest, "\n---")
	if end < 0 {
		return "", content, fmt.Errorf("no closing frontmatter delimiter")
	}
	fm = rest[:end]
	body = rest[end+4:] // skip \n---
	if len(body) > 0 && body[0] == '\n' {
		body = body[1:]
	}
	return fm, body, nil
}

func parseFrontmatterFields(fm string, spec *Spec) {
	if m := reFMName.FindStringSubmatch(fm); len(m) > 1 {
		spec.Name = strings.TrimSpace(m[1])
	}
	if m := reFMVersion.FindStringSubmatch(fm); len(m) > 1 {
		spec.Version = strings.TrimSpace(m[1])
	}
	if m := reFMStatus.FindStringSubmatch(fm); len(m) > 1 {
		status := strings.ToLower(strings.TrimSpace(m[1]))
		switch status {
		case "active", "deprecated", "draft":
			spec.Status = status
		}
	}
}

// ── requirement parsing ───────────────────────────────────────────────────────

// reReqSection matches any heading that looks like a Requirements section.
var reReqSection = regexp.MustCompile(`(?i)^#{1,3}\s+requirements?\s*$`)

// reReqLine matches requirement list items:
//   - MUST: description [status]
//   - SHOULD: description
//   - MAY: description
var reReqLine = regexp.MustCompile(`(?i)^[-*]\s+(MUST|SHOULD|MAY)[:\s]\s*(.+)$`)

// reReqStatus matches an inline status tag at the end: [verified], [failed], [pending]
var reReqStatus = regexp.MustCompile(`(?i)\[(verified|failed|pending)\]\s*$`)

func parseRequirements(body string) []Requirement {
	lines := strings.Split(body, "\n")

	var inSection bool
	var reqs []Requirement
	counter := 0

	for _, line := range lines {
		line = strings.TrimRight(line, " \t\r")

		if reReqSection.MatchString(line) {
			inSection = true
			continue
		}

		// A new ## heading ends the requirements section.
		if inSection && isHeading(line) {
			break
		}

		if !inSection {
			continue
		}

		m := reReqLine.FindStringSubmatch(line)
		if m == nil {
			continue
		}

		priority := strings.ToLower(m[1])
		description := strings.TrimSpace(m[2])

		// Extract inline status if present.
		reqStatus := "pending"
		if sm := reReqStatus.FindStringSubmatch(description); sm != nil {
			reqStatus = strings.ToLower(sm[1])
			description = strings.TrimSpace(reReqStatus.ReplaceAllString(description, ""))
		}

		counter++
		reqs = append(reqs, Requirement{
			ID:          fmt.Sprintf("REQ-%03d", counter),
			Description: description,
			Priority:    priority,
			Status:      reqStatus,
		})
	}
	return reqs
}
