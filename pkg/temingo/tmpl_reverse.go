package temingo

// tmpl_reverse returns a slice in reverse order
// It accepts a single slice argument: reverse .slice
func tmpl_reverse(slice []interface{}) []interface{} {
	if slice == nil {
		return nil
	}
	if len(slice) == 0 {
		return []interface{}{}
	}

	result := make([]interface{}, len(slice))
	for i := 0; i < len(slice); i++ {
		result[i] = slice[len(slice)-1-i]
	}
	return result
}
