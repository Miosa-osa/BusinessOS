package adapters

// Shared helpers used across adapters for nested JSON-like payload access.
// Kept in one spot so each adapter file focuses on its domain logic.

func stringField(m map[string]any, key string) string {
	v, ok := m[key]
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
}

// firstString returns the first non-empty string value among keys.
func firstString(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if s := stringField(m, k); s != "" {
			return s
		}
	}
	return ""
}

// nestedMap descends through map[string]any hops, returning nil if any step
// is missing or of the wrong shape. Matches Elixir's get_in/2 behavior.
func nestedMap(m map[string]any, path ...string) map[string]any {
	cur := m
	for _, p := range path {
		v, ok := cur[p]
		if !ok {
			return nil
		}
		sub, ok := v.(map[string]any)
		if !ok {
			return nil
		}
		cur = sub
	}
	return cur
}

// nestedString is nestedMap + stringField in one shot.
func nestedString(m map[string]any, path ...string) string {
	if len(path) < 2 {
		if len(path) == 1 {
			return stringField(m, path[0])
		}
		return ""
	}
	leaf := path[len(path)-1]
	sub := nestedMap(m, path[:len(path)-1]...)
	if sub == nil {
		return ""
	}
	return stringField(sub, leaf)
}

// anyList returns the value at key as []any, or nil if missing/wrong shape.
func anyList(m map[string]any, key string) []any {
	v, ok := m[key]
	if !ok {
		return nil
	}
	list, _ := v.([]any)
	return list
}
