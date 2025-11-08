package mergeYaml

// mergeLists merges two arrays recursively with option for overriding or merging/extending
func mergeLists(src, dst []interface{}, override bool) []interface{} {
	if override { // If overriding
		return src // Return source instead
	} else { // If not overriding
		dst = append(dst, src...) // Append src to dst
		return dst                // Return dst
	}
}
