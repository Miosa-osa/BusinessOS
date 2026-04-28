package wiki

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestPageFromMarkdownExtractsFrontmatter(t *testing.T) {
	md := "---\nslug: q2-plan\naudience: exec-brief\nversion: 3\n---\n\n# Q2 Plan\n\nBody here."
	page, err := FromMarkdown(md, nil)
	if err != nil {
		t.Fatal(err)
	}
	if page.Slug != "q2-plan" {
		t.Fatalf("slug: %q", page.Slug)
	}
	if page.Audience != "exec-brief" {
		t.Fatalf("audience: %q", page.Audience)
	}
	if page.Version != 3 {
		t.Fatalf("version: %d", page.Version)
	}
	if !strings.Contains(page.Body, "Q2 Plan") {
		t.Fatalf("body lost: %q", page.Body)
	}
}

func TestPageMissingSlug(t *testing.T) {
	if _, err := FromMarkdown("# no frontmatter", nil); !errors.Is(err, ErrMissingSlug) {
		t.Fatalf("expected ErrMissingSlug, got %v", err)
	}
}

func TestPageRoundTrip(t *testing.T) {
	page := NewPage("tenant-1", "rfc-001", "engineering")
	page.Body = "Body."
	page.Frontmatter["status"] = "draft"
	md := page.ToMarkdown()
	got, err := FromMarkdown(md, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got.Slug != "rfc-001" || got.Audience != "engineering" {
		t.Fatalf("round trip lost identity: %+v", got)
	}
}

func TestParseDirectives(t *testing.T) {
	body := "intro {{cite: optimal://a/b}} and {{include: x.md}} plus [[some-slug]] {{bogus: nope}}"
	out := ParseDirectives(body)
	if len(out) != 4 {
		t.Fatalf("expected 4 directives, got %d", len(out))
	}
	verbs := map[string]bool{}
	for _, d := range out {
		verbs[d.Verb] = true
	}
	for _, want := range []string{"cite", "include", "wikilink", "bogus"} {
		if !verbs[want] {
			t.Fatalf("missing verb %s in %+v", want, out)
		}
	}
}

func TestRenderMarkdownEmitsFootnotes(t *testing.T) {
	body := "Earnings doubled {{cite: optimal://earnings/q2}} this quarter."
	resolver := func(d Directive) (Resolved, error) {
		return Resolved{Source: d.Argument}, nil
	}
	r := Render(body, resolver, FormatMarkdown)
	if !strings.Contains(r.Body, "[^1]") {
		t.Fatalf("footnote marker missing: %q", r.Body)
	}
	if !strings.Contains(r.Body, "[^1]: optimal://earnings/q2") {
		t.Fatalf("footnote target missing: %q", r.Body)
	}
}

func TestRenderUnknownVerbWarnsButPreservesContent(t *testing.T) {
	body := "before {{bogus: x}} after"
	r := Render(body, func(Directive) (Resolved, error) { return Resolved{}, nil }, FormatMarkdown)
	if !strings.Contains(r.Body, "wiki:warn") {
		t.Fatalf("warn comment missing: %q", r.Body)
	}
	if !strings.Contains(r.Body, "{{bogus: x}}") {
		t.Fatalf("original directive should be preserved: %q", r.Body)
	}
	if len(r.Errors) != 1 {
		t.Fatalf("expected 1 error, got %v", r.Errors)
	}
}

func TestIntegrityFlagsBrokenCitations(t *testing.T) {
	page, _ := FromMarkdown("---\nslug: x\n---\n\n{{cite: optimal://broken}}", nil)
	resolver := func(uri string) error { return errors.New("not found") }
	r := Check(page, CheckOptions{ResolveURI: resolver})
	if r.OK {
		t.Fatal("expected broken citation to fail check")
	}
	hasBroken := false
	for _, i := range r.Issues {
		if i.Kind == "broken_citation" {
			hasBroken = true
		}
	}
	if !hasBroken {
		t.Fatalf("missing broken_citation: %+v", r.Issues)
	}
}

func TestIntegritySchemaRequiredSection(t *testing.T) {
	page, _ := FromMarkdown("---\nslug: x\n---\n\nBody only.", nil)
	r := AgainstSchema(page, SchemaRules{RequiredSections: []string{"Summary"}}, CheckOptions{})
	if r.OK {
		t.Fatal("expected missing-section to fail schema")
	}
}

func TestMemoryStoreRoundTrip(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	page := NewPage("tenant-A", "doc-1", DefaultAudience)
	page.Body = "first version"
	if err := store.Put(ctx, page); err != nil {
		t.Fatal(err)
	}
	got, err := store.Latest(ctx, "tenant-A", "doc-1", "")
	if err != nil {
		t.Fatal(err)
	}
	if got.Body != "first version" {
		t.Fatalf("latest body: %q", got.Body)
	}
}

func TestFacadeCurateBumpsVersionAndStoresCitations(t *testing.T) {
	ctx := context.Background()
	w := New(NewMemoryStore())

	curate := func(_ context.Context, existing *Page, _ []Citation) (string, error) {
		if existing == nil {
			return "v1 body", nil
		}
		return "v2 body", nil
	}

	cits := []Citation{{ChunkID: "optimal://chunk/1"}}
	first, err := w.Curate(ctx, "tenant-X", "intro", cits, curate, CurateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if first.Version != 1 || first.Body != "v1 body" {
		t.Fatalf("first: %+v", first)
	}

	second, err := w.Curate(ctx, "tenant-X", "intro", cits, curate, CurateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if second.Version != 2 {
		t.Fatalf("expected version=2, got %d", second.Version)
	}

	stored, err := w.Store.ListCitations(ctx, "tenant-X", "intro", "")
	if err != nil {
		t.Fatal(err)
	}
	// Idempotent storage — same citation set replayed must not duplicate.
	if len(stored) != 1 {
		t.Fatalf("expected 1 citation, got %d", len(stored))
	}
	if stored[0].ClaimHash == "" {
		t.Fatal("ClaimHash should be auto-derived")
	}
}
