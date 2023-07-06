package markdown2html

import (
	"bytes"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

func Convert(markdown []byte) ([]byte, error) {
	var (
		err error
		buf bytes.Buffer
	)

	converter := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
		),
	)

	err = converter.Convert([]byte(markdown), &buf)

	return buf.Bytes(), err
}
