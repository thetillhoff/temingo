package mergeYaml

// Merge two maps recursively with option for overriding or merging/extending
func mergeMaps(dst, src map[interface{}]interface{}, override bool) map[interface{}]interface{} {
	for k, v := range src {
		if dstVal, ok := dst[k]; ok {
			dst[k] = Merge(dstVal, v, override)
		} else {
			dst[k] = v
		}
	}
	return dst
}
