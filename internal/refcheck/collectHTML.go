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
//
// base[href] is absent for a different reason: it is a resolution rule rather
// than a target, and it names a directory, which no output path can match. It is
// read separately, to suppress resolution in documents that declare one.
var urlAttrs = map[string][]string{
	"a": {"href"}, "area": {"href"}, "link": {"href"},
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

	// A base element changes what every relative reference in the document points
	// at, and it can appear after the references it governs, so it is found first.
	hasBase := false
	var findBase func(n *html.Node)
	findBase = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "base" && strings.TrimSpace(attrValue(n, "href")) != "" {
			hasBase = true
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findBase(c)
		}
	}
	findBase(doc)

	var walk func(n *html.Node)
	walk = func(n *html.Node) {
		// Only elements are inspected. Comment nodes carry no attributes, so a
		// commented-out link yields nothing.
		if n.Type == html.ElementNode {
			refs = append(refs, refsFromElement(file, n)...)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)

	if hasBase {
		for i := range refs {
			// Only relative references are affected: a root-relative or absolute
			// URL ignores the base entirely.
			if refs[i].Origin == OriginInternal && !strings.HasPrefix(refs[i].URL, "/") {
				refs[i].ResolutionUnknown = true
			}
		}
	}

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

// splitURLs yields each URL in an attribute value.
//
// srcset holds a candidate list where each entry is a URL optionally followed by
// a descriptor, and entries are comma-separated. Splitting on commas first is
// wrong: a URL may contain them - image-transform paths like
// /img/w_100,h_100/a.jpg do, and every data URI does - so the URL is taken as
// the run up to the next whitespace, and only the descriptor that follows it is
// scanned for the separating comma.
func splitURLs(attr, raw string) []string {
	if attr != "srcset" {
		return []string{strings.TrimSpace(raw)}
	}

	var out []string
	rest := raw
	for {
		rest = strings.TrimLeft(rest, " \t\r\n\f")
		if rest == "" {
			return out
		}

		// The URL ends at whitespace, or at a comma that terminates the whole
		// candidate when no descriptor follows.
		end := strings.IndexAny(rest, " \t\r\n\f")
		if end == -1 {
			out = append(out, strings.TrimSuffix(rest, ","))
			return out
		}

		url := rest[:end]
		rest = rest[end:]

		// A trailing comma on the URL itself means this candidate carried no
		// descriptor.
		if trimmed, cut := strings.CutSuffix(url, ","); cut {
			out = append(out, trimmed)
			continue
		}
		out = append(out, url)

		// Skip the descriptor, which runs to the next comma.
		if next := strings.Index(rest, ","); next == -1 {
			return out
		} else {
			rest = rest[next+1:]
		}
	}
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
