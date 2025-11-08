package mergeYaml

import "log/slog"

// mergeMaps merges two maps recursively with option for overriding or merging/extending
func mergeMaps(src, dst map[interface{}]interface{}, override bool) map[interface{}]interface{} {
	for k, v := range src { // For each key-value pair in src
		if override { // If overriding
			// Log at debug/verbose level that we're setting dst[k]
			slog.Debug("Overriding merged map value", "key", k, "value", v)
			dst[k] = v // Override dst[key] with value from src
		} else { // If not overriding
			if dstVal, ok := dst[k]; ok { // Retrieve dst[key] if exists
				dst[k] = Merge(v, dstVal, override) // Merge srv[key] and dst[key] (recusion) and store result in dst[key]
			} else { // If dst[key] doesn't exist
				dst[k] = v // Set dst[key] to value from src
			}
		}
	}
	return dst
}
