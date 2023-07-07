package mergeYaml

// Takes two unmarshalled yaml files, and merges src into dst while overriding existing values
// Enabling override mode will override existing keys in dst
// Disabling override mode will merge existing maps recursively & append to existing lists in dst
func Merge(src, dst interface{}, override bool) interface{} {
	switch dst := dst.(type) {
	case map[interface{}]interface{}:
		if src, ok := src.(map[interface{}]interface{}); ok {
			return mergeMaps(dst, src, override)
		}
	case []interface{}:
		if src, ok := src.([]interface{}); ok {
			return mergeLists(dst, src, override)
		}
	}

	// Replace the previous value with the new one if the type doesn't match
	return src
}
