package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseWorkspaceSpecs(t *testing.T) {
	specs, err := parseWorkspaceSpecs("businessos=BusinessOS,acme-labs=Acme Labs,businessos")
	if err != nil {
		t.Fatal(err)
	}
	if len(specs) != 2 {
		t.Fatalf("got %d specs, want 2", len(specs))
	}
	if specs[0].slug != "businessos" || specs[0].name != "BusinessOS" {
		t.Fatalf("unexpected first spec: %#v", specs[0])
	}
	if specs[1].slug != "acme-labs" || specs[1].name != "Acme Labs" {
		t.Fatalf("unexpected second spec: %#v", specs[1])
	}
}

func TestCollectDocumentsSkipsEngineInternals(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, ".wiki"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "AGENTS.md"), []byte("agents"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, ".optimal"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".optimal", "internal.md"), []byte("private index internals"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "AGENTS.abstract.md"), []byte("generated"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "AGENTS.overview.md"), []byte("generated"), 0o644); err != nil {
		t.Fatal(err)
	}

	docs, err := collectDocuments(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(docs) != 1 || docs[0].path != "AGENTS.md" {
		t.Fatalf("documents = %#v, want only AGENTS.md", docs)
	}
}

func TestCollectDocumentsHonorsBusinessOSIgnore(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "private", "nested"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".businessosignore"), []byte("private/\nexact.md\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for path, body := range map[string]string{
		"shared.md":             "shared",
		"exact.md":              "private",
		"private/nested/doc.md": "private",
	} {
		if err := os.WriteFile(filepath.Join(root, path), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	docs, err := collectDocuments(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(docs) != 1 || docs[0].path != "shared.md" {
		t.Fatalf("documents = %#v, want only shared.md", docs)
	}
}

func TestTitleFromNamePreservesProductInitialisms(t *testing.T) {
	if got := titleFromName("acme-labs"); got != "Acme Labs" {
		t.Fatalf("titleFromName = %q", got)
	}
	if got := titleFromName("miosa.md"); got != "MIOSA" {
		t.Fatalf("titleFromName = %q", got)
	}
}

func TestEmptyProjectionDoesNotProduceSyntheticDocuments(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, ".wiki"), 0o755); err != nil {
		t.Fatal(err)
	}
	docs, err := collectDocuments(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(docs) != 0 {
		t.Fatalf("got %d documents, want none", len(docs))
	}
}
