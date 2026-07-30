package refcheck

import (
	"bytes"
	"strings"

	"golang.org/x/net/html"
)

// urlAttrs maps an element name to the attributes on it that carry a URL.
//
// form[action] is deliberately absent: a form posts to a server route, which
// need not exist as an output file, so treating it as a reference would report
// a broken target on any site with a form.
var urlAttrs = map[string][]string{
	"a": {"href"}, "area": {"href"}, "link": {"href"}, "base": {"href"},
	"script": {"src"}, "iframe": {"src"}, "embed": {"src"}, "track": {"src"},
	"audio": {"src"}, "input": {"src"},
	"img":    {"src", "srcset"},
	"source": {"src", "srcset"},
	"video":  {"src", "poster"},
	"object": {"data"},
}

// integrityRels are the link relations a browser verifies an integrity hash for.
var integrityRels = map[string]bool{"stylesheet": true, "preload": true, "modulepreload": true}

// CollectHTML returns every reference in content. Text and comment nodes are
// never references: a URL in a code sample or a commented-out link addresses
// nothing, and reporting it would be advice the author cannot act on.
func CollectHTML(file string, content []byte) []Reference {
	refs := []Reference{}

	doc, err := html.Parse(bytes.NewReader(content))
	if err != nil {
		return refs
	}

	var walk func(n *html.Node)
	walk = func(n *html.Node) {
		// Only elements are inspected. That is what guarantees text and comment
		// nodes never yield references.
		if n.Type == html.ElementNode {
			refs = append(refs, refsFromElement(file, n)...)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)

	return refs
}

func refsFromElement(file string, n *html.Node) []Reference {
	var refs []Reference

	// A style attribute holds stylesheet content on any element.
	if styleAttr := attrValue(n, "style"); strings.TrimSpace(styleAttr) != "" {
		refs = append(refs, CollectCSS(file, []byte(styleAttr))...)
	}

	// A style element's content is a stylesheet.
	if n.Data == "style" {
		var sb strings.Builder
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.TextNode {
				sb.WriteString(c.Data)
			}
		}
		return append(refs, CollectCSS(file, []byte(sb.String()))...)
	}

	attrs, ok := urlAttrs[n.Data]
	if !ok {
		return refs
	}

	var (
		hasIntegrity   = attrValue(n, "integrity") != ""
		hasCrossOrigin = attrPresent(n, "crossorigin")
		rel            = strings.ToLower(strings.TrimSpace(attrValue(n, "rel")))
	)

	for _, attr := range attrs {
		raw := attrValue(n, attr)
		if strings.TrimSpace(raw) == "" {
			continue
		}

		role := n.Data + " " + attr
		canCarry := false
		switch {
		case n.Data == "script" && attr == "src":
			canCarry = true
		case n.Data == "link" && attr == "href":
			if rel != "" {
				role = n.Data + " " + rel + " " + attr
			}
			for _, r := range strings.Fields(rel) {
				if integrityRels[r] {
					canCarry = true
				}
			}
		}

		for _, u := range splitURLs(attr, raw) {
			origin := Classify(u)
			if origin == OriginIgnored {
				continue
			}
			refs = append(refs, Reference{
				File: file, URL: u, Role: role, Origin: origin,
				CanCarryIntegrity: canCarry,
				HasIntegrity:      hasIntegrity,
				HasCrossOrigin:    hasCrossOrigin,
			})
		}
	}

	return refs
}

// splitURLs yields each URL in an attribute value. srcset holds a
// comma-separated candidate list where each entry may carry a descriptor.
func splitURLs(attr, raw string) []string {
	if attr != "srcset" {
		return []string{strings.TrimSpace(raw)}
	}
	var out []string
	for _, candidate := range strings.Split(raw, ",") {
		if f := strings.Fields(candidate); len(f) > 0 {
			out = append(out, f[0])
		}
	}
	return out
}

func attrValue(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if strings.EqualFold(a.Key, key) {
			return a.Val
		}
	}
	return ""
}

func attrPresent(n *html.Node, key string) bool {
	for _, a := range n.Attr {
		if strings.EqualFold(a.Key, key) {
			return true
		}
	}
	return false
}
