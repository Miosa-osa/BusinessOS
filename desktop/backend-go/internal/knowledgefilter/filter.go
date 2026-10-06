package knowledgefilter

import (
	"os"
	"path/filepath"
	"strings"
)

// Load reads workspace-relative knowledge paths that must not be projected to
// the shared BusinessOS cloud copy.
func Load(root string) []string {
	data, err := os.ReadFile(filepath.Join(root, ".businessosignore"))
	if err != nil {
		return nil
	}
	var rules []string
	for _, line := range strings.Split(string(data), "\n") {
		rule := filepath.ToSlash(strings.TrimSpace(line))
		rule = strings.TrimPrefix(rule, "./")
		if rule == "" || strings.HasPrefix(rule, "#") {
			continue
		}
		rules = append(rules, rule)
	}
	return rules
}

// Ignored matches exact files and directory prefixes ending in a slash.
func Ignored(rel string, rules []string) bool {
	rel = filepath.ToSlash(rel)
	for _, rule := range rules {
		if strings.HasSuffix(rule, "/") {
			if strings.HasPrefix(rel, rule) {
				return true
			}
			continue
		}
		if rel == rule {
			return true
		}
	}
	return false
}
