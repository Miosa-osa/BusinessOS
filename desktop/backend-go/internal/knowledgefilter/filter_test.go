package knowledgefilter

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadAndIgnore(t *testing.T) {
	root := t.TempDir()
	content := "# comment\nprivate/\nexact.md\n\n"
	if err := os.WriteFile(filepath.Join(root, ".businessosignore"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	rules := Load(root)
	if !reflect.DeepEqual(rules, []string{"private/", "exact.md"}) {
		t.Fatalf("rules = %#v", rules)
	}
	for _, path := range []string{"private/doc.md", "private/nested/doc.md", "exact.md"} {
		if !Ignored(path, rules) {
			t.Errorf("expected %q to be ignored", path)
		}
	}
	for _, path := range []string{"shared.md", "private-other/doc.md"} {
		if Ignored(path, rules) {
			t.Errorf("expected %q to remain visible", path)
		}
	}
}
