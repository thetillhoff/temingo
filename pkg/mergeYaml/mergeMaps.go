package mergeYaml

import "fmt"

// Merge two maps recursively with option for overriding or merging/extending
func mergeMaps(src, dst map[interface{}]interface{}, override bool) map[interface{}]interface{} {
	fmt.Println("src:", src)
	fmt.Println("dst:", dst)
	for k, v := range src { // For each key-value pair in src
		if override { // If overriding
			fmt.Println("setting dst[", k, "] to", v)
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
