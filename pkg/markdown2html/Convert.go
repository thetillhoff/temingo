package markdown2html

import (
	"bytes"

	"github.com/yuin/goldmark/v2"
	"github.com/yuin/goldmark/v2/extension"
	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/renderer/html"
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
