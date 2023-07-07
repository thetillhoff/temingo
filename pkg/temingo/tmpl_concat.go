package temingo

func tmpl_concat(args ...string) string {
	result := ""
	for _, str := range args {
		result += str
	}
	return result
}
