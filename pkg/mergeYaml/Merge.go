package mergeYaml

// Merge takes two unmarshalled yaml files, and merges src into dst while overriding existing values.
// Enabling override mode will override existing keys in dst.
// Disabling override mode will merge existing maps recursively & append to existing lists in dst.
func Merge(src, dst interface{}, override bool) interface{} {
	switch dst := dst.(type) {
	case map[interface{}]interface{}: // If dst is map
		if src, ok := src.(map[interface{}]interface{}); ok { // If src is map
			return mergeMaps(src, dst, override) // Merge maps
		}
	case []interface{}: // If dst is list
		if src, ok := src.([]interface{}); ok { // If src is list
			return mergeLists(src, dst, override) // Merge lists
		}
	}

	if override {
		return src // Replace the previous value with the new one if the type doesn't match
	} else {
		return dst // Leave previous value be if the type doesn't match
	}
}
