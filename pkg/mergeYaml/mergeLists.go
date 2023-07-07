package mergeYaml

// Merge two arrays recursively with option for overriding or merging/extending
func mergeLists(dst, src []interface{}, override bool) []interface{} {
	if override {
		return src
	}

	dst = append(dst, src...)
	return dst
}
