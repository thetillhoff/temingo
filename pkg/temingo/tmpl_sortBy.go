package temingo

import (
	"fmt"
	"sort"
)

// tmpl_sortBy sorts a map[string]any by a nested string field, returning
// a slice of {"key": folderName, "value": metaObject} maps for template iteration.
// Entries missing the sort field sort last.
// Usage: {{ range sortBy "date" .childMeta }}{{ .key }} {{ .value.title }}{{ end }}
func tmpl_sortBy(field string, m map[string]any) []any {
	type entry struct {
		key     string
		value   any
		sortKey string
	}

	entries := make([]entry, 0, len(m))
	for k, v := range m {
		sortKey := ""
		if vm, ok := v.(map[string]any); ok {
			if fv, ok := vm[field]; ok {
				sortKey = fmt.Sprint(fv)
			}
		}
		entries = append(entries, entry{key: k, value: v, sortKey: sortKey})
	}

	sort.SliceStable(entries, func(i, j int) bool {
		ei, ej := entries[i], entries[j]
		if ei.sortKey == "" && ej.sortKey == "" {
			return ei.key < ej.key
		}
		if ei.sortKey == "" {
			return false
		}
		if ej.sortKey == "" {
			return true
		}
		return ei.sortKey < ej.sortKey
	})

	result := make([]any, len(entries))
	for i, e := range entries {
		result[i] = map[string]any{
			"key":   e.key,
			"value": e.value,
		}
	}
	return result
}
