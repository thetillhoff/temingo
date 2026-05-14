package prettifyhtml

import (
	"bytes"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// Format parses s as HTML and returns it with consistent 2-space indentation.
// Returns s unchanged on parse error or if s is empty.
func Format(s string) string {
	if s == "" {
		return s
	}
	doc, err := html.Parse(strings.NewReader(s))
	if err != nil {
		return s
	}
	var buf bytes.Buffer
	renderNode(&buf, doc, 0)
	return buf.String()
}

func renderNode(buf *bytes.Buffer, n *html.Node, depth int) {
	switch n.Type {
	case html.DocumentNode:
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			renderNode(buf, c, depth)
		}
	case html.DoctypeNode:
		buf.WriteString("<!DOCTYPE " + n.Data + ">\n")
	case html.CommentNode:
		buf.WriteString(pad(depth) + "<!--" + n.Data + "-->\n")
	case html.TextNode:
		if text := strings.TrimSpace(n.Data); text != "" {
			if n.Parent != nil && isBlock(n.Parent.DataAtom) {
				buf.WriteString(pad(depth) + text + "\n")
			} else {
				buf.WriteString(text)
			}
		}
	case html.ElementNode:
		block := isBlock(n.DataAtom)
		if block {
			buf.WriteString(pad(depth))
		}
		buf.WriteByte('<')
		buf.WriteString(n.Data)
		for _, a := range n.Attr {
			buf.WriteByte(' ')
			if a.Namespace != "" {
				buf.WriteString(a.Namespace + ":" + a.Key)
			} else {
				buf.WriteString(a.Key)
			}
			buf.WriteString(`="` + html.EscapeString(a.Val) + `"`)
		}
		if isVoid(n.DataAtom) {
			buf.WriteByte('>')
			if block {
				buf.WriteByte('\n')
			}
			return
		}
		buf.WriteByte('>')
		if block && n.FirstChild != nil {
			buf.WriteByte('\n')
		}
		childDepth := depth
		if block {
			childDepth++
		}
		if isRaw(n.DataAtom) {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				var raw bytes.Buffer
				_ = html.Render(&raw, c)
				buf.Write(raw.Bytes())
			}
		} else {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				renderNode(buf, c, childDepth)
			}
		}
		if block {
			buf.WriteString(pad(depth) + "</" + n.Data + ">\n")
		} else {
			buf.WriteString("</" + n.Data + ">")
		}
	}
}

func pad(depth int) string {
	return strings.Repeat("  ", depth)
}

func isBlock(a atom.Atom) bool {
	switch a {
	case atom.Address, atom.Article, atom.Aside, atom.Blockquote,
		atom.Body, atom.Canvas, atom.Dd, atom.Details, atom.Dialog,
		atom.Div, atom.Dl, atom.Dt, atom.Fieldset, atom.Figcaption,
		atom.Figure, atom.Footer, atom.Form,
		atom.H1, atom.H2, atom.H3, atom.H4, atom.H5, atom.H6,
		atom.Head, atom.Header, atom.Hr, atom.Html,
		atom.Legend, atom.Li, atom.Link, atom.Main, atom.Menu,
		atom.Meta, atom.Nav, atom.Noscript, atom.Ol,
		atom.P, atom.Pre, atom.Script, atom.Section, atom.Style,
		atom.Summary, atom.Table, atom.Tbody, atom.Td, atom.Tfoot,
		atom.Th, atom.Thead, atom.Title, atom.Tr, atom.Ul:
		return true
	}
	return false
}

func isVoid(a atom.Atom) bool {
	switch a {
	case atom.Area, atom.Base, atom.Br, atom.Col, atom.Embed,
		atom.Hr, atom.Img, atom.Input, atom.Link, atom.Meta,
		atom.Param, atom.Source, atom.Track, atom.Wbr:
		return true
	}
	return false
}

func isRaw(a atom.Atom) bool {
	switch a {
	case atom.Script, atom.Style, atom.Pre, atom.Textarea:
		return true
	}
	return false
}
