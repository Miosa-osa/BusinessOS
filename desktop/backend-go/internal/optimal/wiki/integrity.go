package wiki

import (
	"fmt"
	"regexp"
	"strings"
)

// Severity ranks integrity issues — error blocks publication, warning is
// flagged, info is purely informational.
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
	SeverityInfo    Severity = "info"
)

// Issue is one finding from an integrity check.
type Issue struct {
	Severity Severity
	Kind     string
	Message  string
	Location string // anchor / section name; empty when global
}

// Report is the bundle of issues returned by Check / AgainstSchema.
type Report struct {
	PageSlug string
	OK       bool
	Issues   []Issue
}

// URIResolver is used to validate that `{{cite: uri}}` URIs resolve.
// Returning a non-nil error promotes the citation to an error issue.
type URIResolver func(uri string) error

// CheckOptions tune Check; zero values use sensible defaults.
type CheckOptions struct {
	ResolveURI URIResolver // nil = don't validate citations live
	MaxBytes   int         // soft size ceiling; 0 disables
}

// Check runs the full integrity suite against a page body.
// The OK field is true when no error-severity issues are found.
func Check(page *Page, opts CheckOptions) Report {
	r := Report{PageSlug: page.Slug, OK: true}
	r.Issues = append(r.Issues, checkVerbs(page.Body)...)
	r.Issues = append(r.Issues, checkCitations(page.Body, opts.ResolveURI)...)
	r.Issues = append(r.Issues, checkClaimDensity(page.Body)...)
	r.Issues = append(r.Issues, checkPageSize(page.Body, opts.MaxBytes)...)
	for _, i := range r.Issues {
		if i.Severity == SeverityError {
			r.OK = false
			break
		}
	}
	return r
}

// SchemaRules are the machine-enforceable subset of `.wiki/SCHEMA.md`.
type SchemaRules struct {
	RequiredSections []string // top-level `## Section` headings that must be present
	MinCitations     int      // minimum number of `{{cite:}}` directives
	MaxBytes         int      // 0 disables
}

// AgainstSchema layers schema rules on top of the base Check report. Useful
// for tenant-defined or audience-defined publication policies.
func AgainstSchema(page *Page, schema SchemaRules, opts CheckOptions) Report {
	if opts.MaxBytes == 0 && schema.MaxBytes > 0 {
		opts.MaxBytes = schema.MaxBytes
	}
	r := Check(page, opts)
	r.Issues = append(r.Issues, checkRequiredSections(page.Body, schema.RequiredSections)...)
	r.Issues = append(r.Issues, checkCitationCount(page.Body, schema.MinCitations)...)
	for _, i := range r.Issues {
		if i.Severity == SeverityError {
			r.OK = false
			break
		}
	}
	return r
}

func checkVerbs(body string) []Issue {
	var out []Issue
	for _, d := range ParseDirectives(body) {
		if d.Verb == "wikilink" {
			continue
		}
		if !Whitelist[d.Verb] {
			out = append(out, Issue{
				Severity: SeverityWarning,
				Kind:     "unknown_verb",
				Message:  fmt.Sprintf("verb %q is not whitelisted", d.Verb),
				Location: d.Raw,
			})
		}
	}
	return out
}

func checkCitations(body string, resolve URIResolver) []Issue {
	if resolve == nil {
		return nil
	}
	var out []Issue
	for _, d := range ParseDirectives(body) {
		if d.Verb != "cite" {
			continue
		}
		if err := resolve(d.Argument); err != nil {
			out = append(out, Issue{
				Severity: SeverityError,
				Kind:     "broken_citation",
				Message:  fmt.Sprintf("citation %q does not resolve: %v", d.Argument, err),
				Location: d.Raw,
			})
		}
	}
	return out
}

// claimRe is a heuristic for "this paragraph asserts a fact": presence of
// numbers, percentages, named entities. Matches the Elixir source's
// permissive heuristic — false positives are fine; false negatives lose
// audit value.
var claimRe = regexp.MustCompile(`\b\d+(?:\.\d+)?%?\b|\b[A-Z][a-z]+(?:\s[A-Z][a-z]+){1,3}\b`)

func checkClaimDensity(body string) []Issue {
	var out []Issue
	paragraphs := strings.Split(body, "\n\n")
	for i, p := range paragraphs {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if !claimRe.MatchString(p) {
			continue
		}
		if !strings.Contains(p, "{{cite:") && !strings.Contains(p, "[[") {
			out = append(out, Issue{
				Severity: SeverityWarning,
				Kind:     "uncited_claim",
				Message:  "paragraph contains a claim but no nearby citation",
				Location: fmt.Sprintf("paragraph %d", i+1),
			})
		}
	}
	return out
}

func checkPageSize(body string, max int) []Issue {
	if max <= 0 {
		return nil
	}
	if len(body) > max {
		return []Issue{{
			Severity: SeverityWarning,
			Kind:     "page_too_large",
			Message:  fmt.Sprintf("page is %d bytes (max %d)", len(body), max),
		}}
	}
	return nil
}

func checkRequiredSections(body string, required []string) []Issue {
	var out []Issue
	for _, name := range required {
		needle := "## " + name
		if !strings.Contains(body, needle) {
			out = append(out, Issue{
				Severity: SeverityError,
				Kind:     "missing_section",
				Message:  fmt.Sprintf("required section %q is missing", name),
			})
		}
	}
	return out
}

func checkCitationCount(body string, min int) []Issue {
	if min <= 0 {
		return nil
	}
	count := 0
	for _, d := range ParseDirectives(body) {
		if d.Verb == "cite" {
			count++
		}
	}
	if count < min {
		return []Issue{{
			Severity: SeverityWarning,
			Kind:     "low_citation_density",
			Message:  fmt.Sprintf("page has %d citations (min %d)", count, min),
		}}
	}
	return nil
}
