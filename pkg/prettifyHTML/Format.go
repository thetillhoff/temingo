package prettifyhtml

import "github.com/yosssi/gohtml"

func Format(s string) string {
	return gohtml.Format(s)
}
