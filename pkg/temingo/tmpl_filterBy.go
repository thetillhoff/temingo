package temingo

import "fmt"

// tmpl_filterBy returns a copy of the map keeping only entries where the given
// field equals value OR the field is absent (default-include).
// Usage: {{ range $k, $v := filterBy "publish" true .childMeta }}
func tmpl_filterBy(field string, value any, m map[string]any) map[string]any {
	want := fmt.Sprint(value)
	result := make(map[string]any, len(m))
	for k, v := range m {
		if vm, ok := v.(map[string]any); ok {
			if fv, exists := vm[field]; exists {
				if fmt.Sprint(fv) != want {
					continue
				}
			}
		}
		result[k] = v
	}
	return result
}
