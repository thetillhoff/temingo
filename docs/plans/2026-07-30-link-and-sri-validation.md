# Link and SRI Validation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Check every reference in rendered output at build time - external URLs over the network, internal paths against the build's own output - and expose a template function that pins remote subresources to a content hash.

**Architecture:** A new `pkg/refcheck` package owns reference collection, classification, static checking, internal resolution, and the URL request cache. It has no dependency on `pkg/temingo`. `Render()` calls into it after beautify/minify and before writing, where both the rendered content and the full set of planned output paths are in memory - so checks work under `--dry-run` and need no filesystem reads. The `sri` template function reaches the same request cache through the Engine.

**Tech Stack:** Go 1.25.5, `golang.org/x/net/html` (already a direct dependency), `gopkg.in/yaml.v3`, stdlib `net/http`, `crypto/sha256|sha512`, `testing` + `net/http/httptest`.

**Spec:** `docs/specs/2026-07-30-link-and-sri-validation-design.md`. Every contract there is normative; this plan is one way to satisfy it.

## Global Constraints

- Reference collection walks parsed nodes. Text nodes and comment nodes are never references - a URL in a code sample or a commented-out link produces no finding, ever.
- `canCarryIntegrity` is a property of the reference, not the element. Only `script[src]` and `link[href]` whose `rel` contains `stylesheet`, `preload` or `modulepreload` may carry it. Never raise integrity or crossorigin findings against any other reference.
- A `crossorigin` attribute that is present but empty or unrecognised is equivalent to `anonymous` and is accepted. Only a fully absent attribute is a finding.
- Internal resolution runs against the in-memory set of planned output paths, never the filesystem, so it holds under `--dry-run`.
- Internal resolution reports only proven absence. Anything a server rewrite could plausibly satisfy produces no finding.
- Each distinct URL is requested at most once per build. A URL the allowlist fully covers is never requested.
- Findings never prevent output from being written. Under `strict`, any finding makes the process exit non-zero - including unreachable and unresolvable hosts.
- A failed remote hash is a hard render error, not a finding.
- Default hash algorithm is `sha384`.
- Tests are table-driven with named struct fields, matching `pkg/temingo/tmpl_filterBy_test.go`.

---

### Task 1: Reference type and HTML collection

**Files:**

- Create: `pkg/refcheck/reference.go`
- Create: `pkg/refcheck/collectHTML.go`
- Test: `pkg/refcheck/collectHTML_test.go`

**Interfaces:**

- Consumes: nothing.
- Produces:
  - `type Origin int` with `OriginInternal`, `OriginRemote`, `OriginIgnored`
  - `type Reference struct { File, URL, Role string; Origin Origin; CanCarryIntegrity, HasIntegrity, HasCrossOrigin bool }`
  - `func Classify(rawURL string) Origin`
  - `func CollectHTML(file string, content []byte) []Reference`

- [ ] **Step 1: Write the failing test**

```go
package refcheck

import (
 "reflect"
 "testing"
)

func TestCollectHTML(t *testing.T) {
 tests := []struct {
  name     string
  content  string
  expected []Reference
 }{
  {
   name:     "url in text content is not a reference",
   content:  `<p>see https://example.com/page for details</p>`,
   expected: []Reference{},
  },
  {
   name:     "url in code element is not a reference",
   content:  `<code>&lt;a href="https://example.com"&gt;x&lt;/a&gt;</code>`,
   expected: []Reference{},
  },
  {
   name:     "url in comment is not a reference",
   content:  `<nav><!-- <a href="/reading-notes/">Reading Notes</a> --></nav>`,
   expected: []Reference{},
  },
  {
   name:    "script src can carry integrity",
   content: `<script src="https://cdn.example/x.js"></script>`,
   expected: []Reference{{
    File: "index.html", URL: "https://cdn.example/x.js", Role: "script src",
    Origin: OriginRemote, CanCarryIntegrity: true,
   }},
  },
  {
   name:    "stylesheet link can carry integrity and records both attributes",
   content: `<link rel="stylesheet" href="https://cdn.example/a.css" integrity="sha384-x" crossorigin="anonymous">`,
   expected: []Reference{{
    File: "index.html", URL: "https://cdn.example/a.css", Role: "link stylesheet href",
    Origin: OriginRemote, CanCarryIntegrity: true, HasIntegrity: true, HasCrossOrigin: true,
   }},
  },
  {
   name:    "icon link cannot carry integrity",
   content: `<link rel="icon" href="/favicon.ico">`,
   expected: []Reference{{
    File: "index.html", URL: "/favicon.ico", Role: "link icon href",
    Origin: OriginInternal,
   }},
  },
  {
   name:    "img src cannot carry integrity",
   content: `<img src="images/a.jpg">`,
   expected: []Reference{{
    File: "index.html", URL: "images/a.jpg", Role: "img src", Origin: OriginInternal,
   }},
  },
  {
   name:    "srcset yields one reference per candidate without descriptors",
   content: `<img srcset="a.jpg 1x, b.jpg 2x">`,
   expected: []Reference{
    {File: "index.html", URL: "a.jpg", Role: "img srcset", Origin: OriginInternal},
    {File: "index.html", URL: "b.jpg", Role: "img srcset", Origin: OriginInternal},
   },
  },
  {
   name:     "non-fetchable schemes and fragments are ignored",
   content:  `<a href="mailto:x@example.com">m</a><a href="tel:+1">t</a><a href="#top">f</a>`,
   expected: []Reference{},
  },
  {
   name:    "empty crossorigin is recorded as present",
   content: `<script src="https://cdn.example/x.js" integrity="sha384-x" crossorigin></script>`,
   expected: []Reference{{
    File: "index.html", URL: "https://cdn.example/x.js", Role: "script src",
    Origin: OriginRemote, CanCarryIntegrity: true, HasIntegrity: true, HasCrossOrigin: true,
   }},
  },
 }

 for _, test := range tests {
  t.Run(test.name, func(t *testing.T) {
   got := CollectHTML("index.html", []byte(test.content))
   if !reflect.DeepEqual(got, test.expected) {
    t.Errorf("CollectHTML() = %+v, want %+v", got, test.expected)
   }
  })
 }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd ~/code/thetillhoff/temingo && go test ./pkg/refcheck/ -run TestCollectHTML -v`
Expected: FAIL to build - `undefined: CollectHTML`, `undefined: Reference`.

- [ ] **Step 3: Write `pkg/refcheck/reference.go`**

```go
// Package refcheck collects references from rendered output and reports the
// ones that are broken, unverifiable, or point at nothing the build produced.
package refcheck

import "strings"

// Origin classifies what a reference addresses.
type Origin int

const (
 // OriginInternal addresses the build's own output.
 OriginInternal Origin = iota
 // OriginRemote addresses an http(s) origin.
 OriginRemote
 // OriginIgnored addresses nothing fetchable: a fragment, or a scheme such
 // as mailto or tel.
 OriginIgnored
)

// Reference is one addressable target found in rendered output.
type Reference struct {
 // File is the output-relative path of the file the reference was found in.
 File string
 // URL is the target exactly as written.
 URL string
 // Role names the syntactic position, for reporting back to the author.
 Role string

 Origin Origin

 // CanCarryIntegrity reports whether a browser would verify an integrity
 // hash on this reference. It is a property of the reference, not of its
 // element: a stylesheet reached from inside CSS can never carry one.
 CanCarryIntegrity bool
 HasIntegrity      bool
 HasCrossOrigin    bool
}

// Classify determines what a reference addresses from the URL alone.
func Classify(rawURL string) Origin {
 u := strings.TrimSpace(rawURL)
 if u == "" || strings.HasPrefix(u, "#") {
  return OriginIgnored
 }
 lower := strings.ToLower(u)
 if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") {
  return OriginRemote
 }
 // A protocol-relative URL addresses a remote origin.
 if strings.HasPrefix(u, "//") {
  return OriginRemote
 }
 // Any other scheme is not fetchable as a build artifact.
 if i := strings.Index(u, ":"); i > 0 {
  if !strings.ContainsAny(u[:i], "/?#") {
   return OriginIgnored
  }
 }
 return OriginInternal
}
```

- [ ] **Step 4: Write `pkg/refcheck/collectHTML.go`**

```go
package refcheck

import (
 "bytes"
 "strings"

 "golang.org/x/net/html"
)

// urlAttrs maps an element name to the attributes on it that carry a URL.
var urlAttrs = map[string][]string{
 "a": {"href"}, "area": {"href"}, "link": {"href"}, "base": {"href"},
 "script": {"src"}, "iframe": {"src"}, "embed": {"src"}, "track": {"src"},
 "audio": {"src"}, "input": {"src"},
 "img":    {"src", "srcset"},
 "source": {"src", "srcset"},
 "video":  {"src", "poster"},
 "object": {"data"},
 "form":   {"action"},
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
  if n.Type == html.ElementNode {
   refs = append(refs, refsFromElement(file, n)...)
  }
  // CommentNode and TextNode children are deliberately not descended into
  // for references; walking only elements is what guarantees it.
  for c := n.FirstChild; c != nil; c = c.NextSibling {
   walk(c)
  }
 }
 walk(doc)

 return refs
}

func refsFromElement(file string, n *html.Node) []Reference {
 attrs, ok := urlAttrs[n.Data]
 if !ok {
  return nil
 }

 var (
  refs           []Reference
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
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd ~/code/thetillhoff/temingo && go test ./pkg/refcheck/ -run TestCollectHTML -v`
Expected: PASS, all ten subtests.

- [ ] **Step 6: Commit**

```bash
cd ~/code/thetillhoff/temingo && git add pkg/refcheck/reference.go pkg/refcheck/collectHTML.go pkg/refcheck/collectHTML_test.go && git commit -m "feat(refcheck): collect references from rendered markup"
```

---

### Task 2: Stylesheet collection

**Files:**

- Create: `pkg/refcheck/collectCSS.go`
- Modify: `pkg/refcheck/collectHTML.go` - call CSS collection for style blocks and style attributes
- Test: `pkg/refcheck/collectCSS_test.go`

**Interfaces:**

- Consumes: `Reference`, `Classify` from Task 1.
- Produces: `func CollectCSS(file, role string, css []byte) []Reference`, and `CollectHTML` additionally returning references found in `<style>` content and `style` attributes.

- [ ] **Step 1: Write the failing test**

```go
package refcheck

import (
 "reflect"
 "testing"
)

func TestCollectCSS(t *testing.T) {
 tests := []struct {
  name     string
  css      string
  expected []Reference
 }{
  {
   name:     "no urls",
   css:      `.a { color: red; }`,
   expected: []Reference{},
  },
  {
   name: "background image, quoted and unquoted",
   css:  `.a{background:url("images/a.jpg")}.b{background:url(images/b.jpg)}`,
   expected: []Reference{
    {File: "style.css", URL: "images/a.jpg", Role: "css url()", Origin: OriginInternal},
    {File: "style.css", URL: "images/b.jpg", Role: "css url()", Origin: OriginInternal},
   },
  },
  {
   name: "font face src is a url like any other",
   css:  `@font-face{src:url('https://cdn.example/f.woff2')}`,
   expected: []Reference{
    {File: "style.css", URL: "https://cdn.example/f.woff2", Role: "css url()", Origin: OriginRemote},
   },
  },
  {
   name: "import with and without url()",
   css:  `@import url("https://cdn.example/a.css");@import "b.css";`,
   expected: []Reference{
    {File: "style.css", URL: "https://cdn.example/a.css", Role: "css @import", Origin: OriginRemote},
    {File: "style.css", URL: "b.css", Role: "css @import", Origin: OriginInternal},
   },
  },
  {
   name:     "data uri is ignored",
   css:      `.a{background:url(data:image/gif;base64,R0lGOD)}`,
   expected: []Reference{},
  },
  {
   name:     "css references can never carry integrity",
   css:      `@import url("https://cdn.example/a.css");`,
   expected: []Reference{{File: "style.css", URL: "https://cdn.example/a.css", Role: "css @import", Origin: OriginRemote}},
  },
 }

 for _, test := range tests {
  t.Run(test.name, func(t *testing.T) {
   got := CollectCSS("style.css", "css", []byte(test.css))
   if !reflect.DeepEqual(got, test.expected) {
    t.Errorf("CollectCSS() = %+v, want %+v", got, test.expected)
   }
  })
 }
}

func TestCollectHTMLIncludesStyles(t *testing.T) {
 tests := []struct {
  name     string
  content  string
  expected []Reference
 }{
  {
   name:    "style block",
   content: `<style>.a{background:url(images/a.jpg)}</style>`,
   expected: []Reference{
    {File: "index.html", URL: "images/a.jpg", Role: "css url()", Origin: OriginInternal},
   },
  },
  {
   name:    "style attribute",
   content: `<div style="background:url(images/b.jpg)"></div>`,
   expected: []Reference{
    {File: "index.html", URL: "images/b.jpg", Role: "css url()", Origin: OriginInternal},
   },
  },
  {
   name:     "css in a code element is still not a reference",
   content:  `<code>.a{background:url(images/c.jpg)}</code>`,
   expected: []Reference{},
  },
 }

 for _, test := range tests {
  t.Run(test.name, func(t *testing.T) {
   got := CollectHTML("index.html", []byte(test.content))
   if !reflect.DeepEqual(got, test.expected) {
    t.Errorf("CollectHTML() = %+v, want %+v", got, test.expected)
   }
  })
 }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd ~/code/thetillhoff/temingo && go test ./pkg/refcheck/ -run 'TestCollectCSS|TestCollectHTMLIncludesStyles' -v`
Expected: FAIL to build - `undefined: CollectCSS`.

- [ ] **Step 3: Write `pkg/refcheck/collectCSS.go`**

```go
package refcheck

import (
 "regexp"
 "strings"
)

// ponytail: regexes, not a CSS tokenizer. These match every url() and @import
// form seen in practice, but can also match inside a comment or a string
// literal. False positives are silenced by the allowlist. Replace with a real
// tokenizer when one arrives for CSS minification/beautification - the
// roadmap already wants one.
var (
 cssURLRe    = regexp.MustCompile(`url\(\s*(?:"([^"]*)"|'([^']*)'|([^)'"\s]+))`)
 cssImportRe = regexp.MustCompile(`@import\s+(?:url\(\s*)?(?:"([^"]*)"|'([^']*)'|([^)'";\s]+))`)
)

// CollectCSS returns every reference in stylesheet content. Role identifies the
// syntactic position; no CSS reference can ever carry an integrity hash,
// because CSS has no syntax for one.
func CollectCSS(file, role string, css []byte) []Reference {
 refs := []Reference{}
 s := string(css)

 // @import first, so an imported URL is reported as an import rather than
 // as the url() it may be wrapped in.
 imports := map[string]bool{}
 for _, m := range cssImportRe.FindAllStringSubmatch(s, -1) {
  u := firstNonEmpty(m[1], m[2], m[3])
  imports[u] = true
  if r, ok := cssRef(file, "css @import", u); ok {
   refs = append(refs, r)
  }
 }

 for _, m := range cssURLRe.FindAllStringSubmatch(s, -1) {
  u := firstNonEmpty(m[1], m[2], m[3])
  if imports[u] {
   continue
  }
  if r, ok := cssRef(file, "css url()", u); ok {
   refs = append(refs, r)
  }
 }

 return refs
}

func cssRef(file, role, rawURL string) (Reference, bool) {
 u := strings.TrimSpace(rawURL)
 origin := Classify(u)
 if origin == OriginIgnored {
  return Reference{}, false
 }
 return Reference{File: file, URL: u, Role: role, Origin: origin}, true
}

func firstNonEmpty(vals ...string) string {
 for _, v := range vals {
  if v != "" {
   return v
  }
 }
 return ""
}
```

- [ ] **Step 4: Extend `refsFromElement` in `pkg/refcheck/collectHTML.go`**

Add `"style"` handling. Insert this at the top of `refsFromElement`, before the `urlAttrs` lookup:

```go
 // A style attribute holds stylesheet content on any element.
 if styleAttr := attrValue(n, "style"); strings.TrimSpace(styleAttr) != "" {
  if cssRefs := CollectCSS(file, "css", []byte(styleAttr)); len(cssRefs) > 0 {
   defer func() {}() // no-op; refs appended below
   return append(cssRefs, refsFromURLAttrs(file, n)...)
  }
 }
 if n.Data == "style" {
  var sb strings.Builder
  for c := n.FirstChild; c != nil; c = c.NextSibling {
   if c.Type == html.TextNode {
    sb.WriteString(c.Data)
   }
  }
  return CollectCSS(file, "css", []byte(sb.String()))
 }
 return refsFromURLAttrs(file, n)
}

func refsFromURLAttrs(file string, n *html.Node) []Reference {
```

Then the remainder of the original `refsFromElement` body becomes the body of `refsFromURLAttrs`. Remove the stray `defer func() {}()` line - it is shown only to mark where the original body was split; the final code must not contain it.

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd ~/code/thetillhoff/temingo && go test ./pkg/refcheck/ -v`
Expected: PASS. Task 1's tests must still pass unchanged.

- [ ] **Step 6: Commit**

```bash
cd ~/code/thetillhoff/temingo && git add pkg/refcheck/ && git commit -m "feat(refcheck): collect references from stylesheet content"
```

---

### Task 3: Findings, categories, severities

**Files:**

- Create: `pkg/refcheck/finding.go`
- Test: `pkg/refcheck/finding_test.go`

**Interfaces:**

- Consumes: `Reference` from Task 1.
- Produces: `type Category string` with the eight constants below, `type Severity int` with `SevInfo`/`SevNote`/`SevWarn`, `type Finding struct { Ref Reference; Category Category; Severity Severity; Reason string }`, `func (f Finding) String() string`, `func SortFindings(fs []Finding)`.

- [ ] **Step 1: Write the failing test**

```go
package refcheck

import "testing"

func TestFindingString(t *testing.T) {
 tests := []struct {
  name     string
  finding  Finding
  expected string
 }{
  {
   name: "includes file, url, category and reason",
   finding: Finding{
    Ref:      Reference{File: "index.html", URL: "https://x.dev/a", Role: "a href"},
    Category: CategoryStatus,
    Severity: SevWarn,
    Reason:   "responded 404",
   },
   expected: "index.html: a href https://x.dev/a: status: responded 404",
  },
 }

 for _, test := range tests {
  t.Run(test.name, func(t *testing.T) {
   if got := test.finding.String(); got != test.expected {
    t.Errorf("String() = %q, want %q", got, test.expected)
   }
  })
 }
}

func TestSortFindings(t *testing.T) {
 findings := []Finding{
  {Ref: Reference{File: "b.html"}, Severity: SevInfo, Category: CategoryRedirect},
  {Ref: Reference{File: "a.html"}, Severity: SevWarn, Category: CategoryStatus},
  {Ref: Reference{File: "c.html"}, Severity: SevNote, Category: CategoryGated},
 }

 SortFindings(findings)

 // Most severe first, then by file for stable output.
 want := []Severity{SevWarn, SevNote, SevInfo}
 for i, w := range want {
  if findings[i].Severity != w {
   t.Errorf("findings[%d].Severity = %v, want %v", i, findings[i].Severity, w)
  }
 }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd ~/code/thetillhoff/temingo && go test ./pkg/refcheck/ -run 'TestFinding|TestSortFindings' -v`
Expected: FAIL to build - `undefined: Finding`.

- [ ] **Step 3: Write `pkg/refcheck/finding.go`**

```go
package refcheck

import (
 "fmt"
 "sort"
)

// Category identifies a kind of finding. Values are stable: allowlist entries
// name them.
type Category string

const (
 // CategoryStatus is a reference whose target reported a client or server error.
 CategoryStatus Category = "status"
 // CategoryGated is a reference whose target requires authorisation.
 CategoryGated Category = "gated"
 // CategoryRedirect is a reference whose target redirects elsewhere.
 CategoryRedirect Category = "redirect"
 // CategoryUnreachable is a reference whose target could not be determined.
 CategoryUnreachable Category = "unreachable"
 // CategoryMissingTarget is a reference to the build's own output that
 // resolves to nothing the build produced.
 CategoryMissingTarget Category = "missing-target"
 // CategoryIntegrity is a verifiable subresource carrying no integrity hash.
 CategoryIntegrity Category = "integrity"
 // CategoryCrossOrigin is an integrity hash the browser cannot verify for
 // lack of a CORS opt-in.
 CategoryCrossOrigin Category = "crossorigin"
 // CategoryUnverifiedImport is a cross-origin stylesheet imported by a
 // stylesheet that carries an integrity hash.
 CategoryUnverifiedImport Category = "unverified-import"
 // CategoryCORS is a subresource with an integrity hash whose target does
 // not permit cross-origin reads.
 CategoryCORS Category = "cors"
)

// Severity orders findings for presentation. It does not gate strictness:
// under strict mode every finding is fatal regardless of severity.
type Severity int

const (
 SevInfo Severity = iota
 SevNote
 SevWarn
)

func (s Severity) String() string {
 switch s {
 case SevWarn:
  return "WARN"
 case SevNote:
  return "NOTE"
 default:
  return "INFO"
 }
}

// Finding is one problem with one reference.
type Finding struct {
 Ref      Reference
 Category Category
 Severity Severity
 Reason   string
}

func (f Finding) String() string {
 return fmt.Sprintf("%s: %s %s: %s: %s",
  f.Ref.File, f.Ref.Role, f.Ref.URL, f.Category, f.Reason)
}

// SortFindings orders findings most severe first, then by file and URL so
// output is stable across runs.
func SortFindings(fs []Finding) {
 sort.SliceStable(fs, func(i, j int) bool {
  if fs[i].Severity != fs[j].Severity {
   return fs[i].Severity > fs[j].Severity
  }
  if fs[i].Ref.File != fs[j].Ref.File {
   return fs[i].Ref.File < fs[j].Ref.File
  }
  return fs[i].Ref.URL < fs[j].Ref.URL
 })
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd ~/code/thetillhoff/temingo && go test ./pkg/refcheck/ -run 'TestFinding|TestSortFindings' -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
cd ~/code/thetillhoff/temingo && git add pkg/refcheck/finding.go pkg/refcheck/finding_test.go && git commit -m "feat(refcheck): add finding categories and severities"
```

---

### Task 4: Static checks

**Files:**

- Create: `pkg/refcheck/checkStatic.go`
- Test: `pkg/refcheck/checkStatic_test.go`

**Interfaces:**

- Consumes: `Reference`, `Finding`, `Category`, `Severity` from Tasks 1 and 3.
- Produces: `func CheckStatic(refs []Reference) []Finding`.

- [ ] **Step 1: Write the failing test**

```go
package refcheck

import "testing"

func TestCheckStatic(t *testing.T) {
 tests := []struct {
  name       string
  refs       []Reference
  wantCats   []Category
 }{
  {
   name: "cross-origin script without integrity",
   refs: []Reference{{
    File: "i.html", URL: "https://cdn.example/x.js", Role: "script src",
    Origin: OriginRemote, CanCarryIntegrity: true,
   }},
   wantCats: []Category{CategoryIntegrity},
  },
  {
   name: "integrity without crossorigin is breakage",
   refs: []Reference{{
    File: "i.html", URL: "https://cdn.example/x.js", Role: "script src",
    Origin: OriginRemote, CanCarryIntegrity: true, HasIntegrity: true,
   }},
   wantCats: []Category{CategoryCrossOrigin},
  },
  {
   name: "integrity with crossorigin is clean",
   refs: []Reference{{
    File: "i.html", URL: "https://cdn.example/x.js", Role: "script src",
    Origin: OriginRemote, CanCarryIntegrity: true, HasIntegrity: true, HasCrossOrigin: true,
   }},
   wantCats: nil,
  },
  {
   name: "same-origin script needs no integrity",
   refs: []Reference{{
    File: "i.html", URL: "/local.js", Role: "script src",
    Origin: OriginInternal, CanCarryIntegrity: true,
   }},
   wantCats: nil,
  },
  {
   name: "image is never asked for integrity",
   refs: []Reference{{
    File: "i.html", URL: "https://cdn.example/a.jpg", Role: "img src",
    Origin: OriginRemote,
   }},
   wantCats: nil,
  },
  {
   name: "css reference is never asked for integrity",
   refs: []Reference{{
    File: "s.css", URL: "https://cdn.example/f.woff2", Role: "css url()",
    Origin: OriginRemote,
   }},
   wantCats: nil,
  },
  {
   name: "cross-origin import from a verified stylesheet is a hole",
   refs: []Reference{
    {File: "i.html", URL: "https://cdn.example/a.css", Role: "link stylesheet href",
     Origin: OriginRemote, CanCarryIntegrity: true, HasIntegrity: true, HasCrossOrigin: true},
    {File: "a.css", URL: "https://other.example/b.css", Role: "css @import",
     Origin: OriginRemote},
   },
   wantCats: []Category{CategoryUnverifiedImport},
  },
 }

 for _, test := range tests {
  t.Run(test.name, func(t *testing.T) {
   got := CheckStatic(test.refs)
   if len(got) != len(test.wantCats) {
    t.Fatalf("CheckStatic() returned %d findings (%+v), want %d", len(got), got, len(test.wantCats))
   }
   for i, c := range test.wantCats {
    if got[i].Category != c {
     t.Errorf("findings[%d].Category = %q, want %q", i, got[i].Category, c)
    }
   }
  })
 }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd ~/code/thetillhoff/temingo && go test ./pkg/refcheck/ -run TestCheckStatic -v`
Expected: FAIL to build - `undefined: CheckStatic`.

- [ ] **Step 3: Write `pkg/refcheck/checkStatic.go`**

```go
package refcheck

// CheckStatic produces every finding derivable from references alone, with no
// network access. A build with no reachability still gets the complete static
// finding set.
func CheckStatic(refs []Reference) []Finding {
 var findings []Finding

 // Any stylesheet whose own reference carried an integrity hash implies a
 // protection that its imports do not inherit.
 verifiedSheets := map[string]bool{}
 for _, r := range refs {
  if r.CanCarryIntegrity && r.HasIntegrity {
   verifiedSheets[r.File] = true
  }
 }

 for _, r := range refs {
  if r.Role == "css @import" && r.Origin == OriginRemote && verifiedSheets[r.File] {
   findings = append(findings, Finding{
    Ref: r, Category: CategoryUnverifiedImport, Severity: SevNote,
    Reason: "imported by a stylesheet that carries an integrity hash; imported sheets are not covered by it",
   })
  }

  // Integrity and CORS findings apply only where a browser would verify
  // a hash. Raising them elsewhere is advice the author cannot act on.
  if !r.CanCarryIntegrity || r.Origin != OriginRemote {
   continue
  }

  switch {
  case r.HasIntegrity && !r.HasCrossOrigin:
   findings = append(findings, Finding{
    Ref: r, Category: CategoryCrossOrigin, Severity: SevWarn,
    Reason: "integrity hash on a cross-origin subresource with no CORS opt-in; the browser blocks it outright",
   })
  case !r.HasIntegrity:
   findings = append(findings, Finding{
    Ref: r, Category: CategoryIntegrity, Severity: SevWarn,
    Reason: "cross-origin subresource with no integrity hash",
   })
  }
 }

 return findings
}
```

Note: `verifiedSheets` is keyed by the file the reference was *found in*, which covers a stylesheet that both carries a hash and imports. A hash on a `<link>` in one document and an `@import` inside the fetched stylesheet is not connected, because the fetched stylesheet is not part of the build. That is correct: the spec only requires reporting imports in stylesheets the build produced.

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd ~/code/thetillhoff/temingo && go test ./pkg/refcheck/ -run TestCheckStatic -v`
Expected: PASS, all seven subtests.

- [ ] **Step 5: Commit**

```bash
cd ~/code/thetillhoff/temingo && git add pkg/refcheck/checkStatic.go pkg/refcheck/checkStatic_test.go && git commit -m "feat(refcheck): add static integrity and CORS opt-in checks"
```

---

### Task 5: Internal reference resolution

**Files:**

- Create: `pkg/refcheck/resolveInternal.go`
- Test: `pkg/refcheck/resolveInternal_test.go`

**Interfaces:**

- Consumes: `Reference`, `Finding` from Tasks 1 and 3.
- Produces: `func ResolveInternal(refs []Reference, outputPaths map[string]bool) []Finding`.

- [ ] **Step 1: Write the failing test**

```go
package refcheck

import "testing"

func TestResolveInternal(t *testing.T) {
 outputs := map[string]bool{
  "index.html":            true,
  "style.css":             true,
  "slides/slides.css":     true,
  "slides/index.html":     true,
  "blog/a-post/index.html": true,
  "about.html":            true,
  "images/a.jpg":          true,
 }

 tests := []struct {
  name     string
  refs     []Reference
  wantURLs []string // URLs expected to produce a missing-target finding
 }{
  {
   name:     "root-relative file that exists",
   refs:     []Reference{{File: "index.html", URL: "/slides/slides.css", Origin: OriginInternal}},
   wantURLs: nil,
  },
  {
   name:     "root-relative file that does not exist",
   refs:     []Reference{{File: "index.html", URL: "/slides/slide.css", Origin: OriginInternal}},
   wantURLs: []string{"/slides/slide.css"},
  },
  {
   name:     "directory served by an index document",
   refs:     []Reference{{File: "index.html", URL: "/slides/", Origin: OriginInternal}},
   wantURLs: nil,
  },
  {
   name:     "extensionless path served by a document",
   refs:     []Reference{{File: "index.html", URL: "/about", Origin: OriginInternal}},
   wantURLs: nil,
  },
  {
   name:     "document-relative path",
   refs:     []Reference{{File: "slides/index.html", URL: "slides.css", Origin: OriginInternal}},
   wantURLs: nil,
  },
  {
   name:     "parent-relative path",
   refs:     []Reference{{File: "blog/a-post/index.html", URL: "../../style.css", Origin: OriginInternal}},
   wantURLs: nil,
  },
  {
   name:     "query and fragment are stripped before resolving",
   refs:     []Reference{{File: "index.html", URL: "/style.css?v=1#x", Origin: OriginInternal}},
   wantURLs: nil,
  },
  {
   name:     "unknown directory with no index is reported",
   refs:     []Reference{{File: "index.html", URL: "/reading-notes/", Origin: OriginInternal}},
   wantURLs: []string{"/reading-notes/"},
  },
  {
   name:     "remote references are not resolved here",
   refs:     []Reference{{File: "index.html", URL: "https://x.dev/a", Origin: OriginRemote}},
   wantURLs: nil,
  },
  {
   name:     "escaping the output root is not a finding",
   refs:     []Reference{{File: "index.html", URL: "../../../etc/passwd", Origin: OriginInternal}},
   wantURLs: nil,
  },
 }

 for _, test := range tests {
  t.Run(test.name, func(t *testing.T) {
   got := ResolveInternal(test.refs, outputs)
   if len(got) != len(test.wantURLs) {
    t.Fatalf("ResolveInternal() = %+v, want %d findings", got, len(test.wantURLs))
   }
   for i, u := range test.wantURLs {
    if got[i].Ref.URL != u {
     t.Errorf("findings[%d].Ref.URL = %q, want %q", i, got[i].Ref.URL, u)
    }
    if got[i].Category != CategoryMissingTarget {
     t.Errorf("findings[%d].Category = %q, want %q", i, got[i].Category, CategoryMissingTarget)
    }
   }
  })
 }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd ~/code/thetillhoff/temingo && go test ./pkg/refcheck/ -run TestResolveInternal -v`
Expected: FAIL to build - `undefined: ResolveInternal`.

- [ ] **Step 3: Write `pkg/refcheck/resolveInternal.go`**

```go
package refcheck

import (
 "net/url"
 "path"
 "strings"
)

// ResolveInternal reports references addressing the build's own output that
// resolve to nothing the build produced.
//
// outputPaths is the set of slash-separated, output-root-relative paths the
// build will write. Resolution never touches the filesystem, so it holds under
// a dry run, and it reports only proven absence: anything a server rewrite
// could plausibly satisfy produces no finding.
func ResolveInternal(refs []Reference, outputPaths map[string]bool) []Finding {
 var findings []Finding

 for _, r := range refs {
  if r.Origin != OriginInternal {
   continue
  }

  target, ok := resolveTarget(r)
  if !ok {
   continue
  }

  // A static tree can serve a path as a file, as a directory holding an
  // index document, or as a document whose extension was omitted. A
  // generator cannot know which form a given server prefers, so any of
  // them resolving is enough.
  candidates := []string{
   target,
   path.Join(target, "index.html"),
   target + ".html",
  }
  found := false
  for _, c := range candidates {
   if outputPaths[c] {
    found = true
    break
   }
  }
  if found {
   continue
  }

  findings = append(findings, Finding{
   Ref: r, Category: CategoryMissingTarget, Severity: SevWarn,
   Reason: "no file in the build output answers this path",
  })
 }

 return findings
}

// resolveTarget turns a reference into an output-root-relative path. It reports
// false when the reference cannot be resolved to one, in which case no finding
// may be raised.
func resolveTarget(r Reference) (string, bool) {
 u, err := url.Parse(r.URL)
 if err != nil {
  return "", false
 }
 p, err := url.PathUnescape(u.Path)
 if err != nil {
  return "", false
 }
 if p == "" {
  return "", false
 }

 var target string
 if strings.HasPrefix(p, "/") {
  target = strings.TrimPrefix(p, "/")
 } else {
  target = path.Join(path.Dir(r.File), p)
 }
 target = path.Clean(target)

 // A path that climbs out of the output root is not something the build can
 // speak to.
 if target == ".." || strings.HasPrefix(target, "../") {
  return "", false
 }
 if target == "." || target == "" {
  return "", false
 }

 return target, true
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd ~/code/thetillhoff/temingo && go test ./pkg/refcheck/ -run TestResolveInternal -v`
Expected: PASS, all ten subtests. Note the `/slides/` case passes via the `index.html` candidate and `/about` via the `.html` candidate.

- [ ] **Step 5: Commit**

```bash
cd ~/code/thetillhoff/temingo && git add pkg/refcheck/resolveInternal.go pkg/refcheck/resolveInternal_test.go && git commit -m "feat(refcheck): resolve internal references against build output"
```

---

### Task 6: Allowlist

**Files:**

- Create: `pkg/refcheck/allowlist.go`
- Test: `pkg/refcheck/allowlist_test.go`

**Interfaces:**

- Consumes: `Category` from Task 3.
- Produces: `type AllowEntry struct { URL string; Checks []Category }`, `type Allowlist []AllowEntry`, `func (a Allowlist) Allows(rawURL string, c Category) bool`, `func (a Allowlist) AllowsEverything(rawURL string) bool`, `func (a Allowlist) Filter(fs []Finding) []Finding`.

- [ ] **Step 1: Write the failing test**

```go
package refcheck

import "testing"

func TestAllowlist(t *testing.T) {
 list := Allowlist{
  {URL: "https://paywalled.example/*"},
  {URL: "https://redirecting.example/*", Checks: []Category{CategoryRedirect}},
  {URL: "https://cdn.example/lib/*/x.js", Checks: []Category{CategoryIntegrity}},
 }

 tests := []struct {
  name          string
  url           string
  category      Category
  wantAllowed   bool
  wantEverything bool
 }{
  {name: "no entry matches", url: "https://other.example/a", category: CategoryStatus},
  {
   name: "entry without checks allows any category", url: "https://paywalled.example/a",
   category: CategoryStatus, wantAllowed: true, wantEverything: true,
  },
  {
   name: "entry without checks allows a different category too", url: "https://paywalled.example/a",
   category: CategoryGated, wantAllowed: true, wantEverything: true,
  },
  {
   name: "entry without checks covers nested paths too", url: "https://paywalled.example/a/b/c",
   category: CategoryStatus, wantAllowed: true, wantEverything: true,
  },
  {
   name: "narrowed entry allows the named category", url: "https://redirecting.example/docs",
   category: CategoryRedirect, wantAllowed: true,
  },
  {
   name: "narrowed entry does not allow other categories", url: "https://redirecting.example/docs",
   category: CategoryStatus,
  },
  {
   name: "glob matches a version segment", url: "https://cdn.example/lib/5.2.1/x.js",
   category: CategoryIntegrity, wantAllowed: true,
  },
  {
   name: "glob does not match a different file", url: "https://cdn.example/lib/5.2.1/y.js",
   category: CategoryIntegrity,
  },
 }

 for _, test := range tests {
  t.Run(test.name, func(t *testing.T) {
   if got := list.Allows(test.url, test.category); got != test.wantAllowed {
    t.Errorf("Allows(%q, %q) = %v, want %v", test.url, test.category, got, test.wantAllowed)
   }
   if got := list.AllowsEverything(test.url); got != test.wantEverything {
    t.Errorf("AllowsEverything(%q) = %v, want %v", test.url, got, test.wantEverything)
   }
  })
 }
}

func TestAllowlistFilter(t *testing.T) {
 list := Allowlist{{URL: "https://paywalled.example/*"}}
 findings := []Finding{
  {Ref: Reference{URL: "https://paywalled.example/a"}, Category: CategoryGated},
  {Ref: Reference{URL: "https://other.example/b"}, Category: CategoryStatus},
 }

 got := list.Filter(findings)

 if len(got) != 1 {
  t.Fatalf("Filter() returned %d findings, want 1", len(got))
 }
 if got[0].Ref.URL != "https://other.example/b" {
  t.Errorf("Filter() kept %q, want the unallowed one", got[0].Ref.URL)
 }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd ~/code/thetillhoff/temingo && go test ./pkg/refcheck/ -run TestAllowlist -v`
Expected: FAIL to build - `undefined: Allowlist`.

- [ ] **Step 3: Write `pkg/refcheck/allowlist.go`**

```go
package refcheck

import "path"

// AllowEntry accepts findings for URLs matching a pattern. Checks names the
// categories accepted; empty accepts all of them.
type AllowEntry struct {
 URL    string     `yaml:"url"`
 Checks []Category `yaml:"checks"`
}

// Allowlist is the configured set of accepted findings.
type Allowlist []AllowEntry

// Allows reports whether a finding of category c against rawURL is accepted.
func (a Allowlist) Allows(rawURL string, c Category) bool {
 for _, e := range a {
  if !matchURL(e.URL, rawURL) {
   continue
  }
  if len(e.Checks) == 0 {
   return true
  }
  for _, allowed := range e.Checks {
   if allowed == c {
    return true
   }
  }
 }
 return false
}

// AllowsEverything reports whether every category is accepted for rawURL. When
// it is, the URL need never be requested: the allowlist elides the request
// rather than filtering its result.
func (a Allowlist) AllowsEverything(rawURL string) bool {
 for _, e := range a {
  if matchURL(e.URL, rawURL) && len(e.Checks) == 0 {
   return true
  }
 }
 return false
}

// Filter drops findings the allowlist accepts.
func (a Allowlist) Filter(fs []Finding) []Finding {
 kept := make([]Finding, 0, len(fs))
 for _, f := range fs {
  if !a.Allows(f.Ref.URL, f.Category) {
   kept = append(kept, f)
  }
 }
 return kept
}

// matchURL matches by glob so a versioned URL does not need a fresh entry each
// time its version changes.
//
// A trailing * matches any suffix, including further path segments, because an
// entry written to cover a host is meant to cover everything under it. An
// interior * matches within one path segment, which is what makes a pattern
// like https://cdn.example/lib/*/x.js pin the filename while accepting any
// version.
func matchURL(pattern, rawURL string) bool {
 if pattern == rawURL {
  return true
 }
 if prefix, ok := strings.CutSuffix(pattern, "*"); ok {
  if strings.HasPrefix(rawURL, prefix) {
   return true
  }
 }
 matched, err := path.Match(pattern, rawURL)
 return err == nil && matched
}
```

Add `"strings"` to the file's imports alongside `"path"`.

Note: `path.Match` alone treats `*` as never crossing `/`, so `https://paywalled.example/*` would match `/a` but not `/a/b` - an entry meant to cover a host would silently miss every nested path. The trailing-`*` prefix case above exists to fix that; the `path.Match` fallback is kept only for interior wildcards. Verified: `path.Match("https://paywalled.example/*", "https://paywalled.example/a/b")` returns false, which is why the special case is needed rather than optional.

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd ~/code/thetillhoff/temingo && go test ./pkg/refcheck/ -run TestAllowlist -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
cd ~/code/thetillhoff/temingo && git add pkg/refcheck/allowlist.go pkg/refcheck/allowlist_test.go && git commit -m "feat(refcheck): add allowlist with per-category narrowing"
```

---

### Task 7: URL request cache

**Files:**

- Create: `pkg/refcheck/cache.go`
- Test: `pkg/refcheck/cache_test.go`

**Interfaces:**

- Consumes: `Allowlist` from Task 6.
- Produces: `type Result struct { Status int; FinalURL string; AllowsCORS bool; Hash string; Err error }`, `type Cache struct{...}`, `func NewCache(allow Allowlist, ttl time.Duration, concurrency int) *Cache`, `func (c *Cache) Fetch(rawURL, algorithm string) Result`.

`algorithm` is empty when no hash is wanted. A `Result` with `Err` set is indeterminate: unreachable or unresolvable.

- [ ] **Step 1: Write the failing test**

```go
package refcheck

import (
 "net/http"
 "net/http/httptest"
 "sync/atomic"
 "testing"
 "time"
)

func TestCacheFetch(t *testing.T) {
 var hits int64

 mux := http.NewServeMux()
 mux.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) {
  atomic.AddInt64(&hits, 1)
  w.Header().Set("Access-Control-Allow-Origin", "*")
  _, _ = w.Write([]byte("hello"))
 })
 mux.HandleFunc("/nocors", func(w http.ResponseWriter, r *http.Request) {
  _, _ = w.Write([]byte("hello"))
 })
 mux.HandleFunc("/missing", func(w http.ResponseWriter, r *http.Request) {
  w.WriteHeader(http.StatusNotFound)
 })
 mux.HandleFunc("/moved", func(w http.ResponseWriter, r *http.Request) {
  http.Redirect(w, r, "/ok", http.StatusMovedPermanently)
 })
 srv := httptest.NewServer(mux)
 defer srv.Close()

 c := NewCache(nil, time.Minute, 4)

 t.Run("200 with CORS", func(t *testing.T) {
  got := c.Fetch(srv.URL+"/ok", "")
  if got.Err != nil {
   t.Fatalf("Err = %v, want nil", got.Err)
  }
  if got.Status != 200 || !got.AllowsCORS {
   t.Errorf("Status=%d AllowsCORS=%v, want 200 true", got.Status, got.AllowsCORS)
  }
 })

 t.Run("200 without CORS", func(t *testing.T) {
  if got := c.Fetch(srv.URL+"/nocors", ""); got.AllowsCORS {
   t.Errorf("AllowsCORS = true, want false")
  }
 })

 t.Run("404", func(t *testing.T) {
  if got := c.Fetch(srv.URL+"/missing", ""); got.Status != 404 {
   t.Errorf("Status = %d, want 404", got.Status)
  }
 })

 t.Run("redirect reports final url", func(t *testing.T) {
  got := c.Fetch(srv.URL+"/moved", "")
  if got.Status != 301 {
   t.Errorf("Status = %d, want 301", got.Status)
  }
  if got.FinalURL != srv.URL+"/ok" {
   t.Errorf("FinalURL = %q, want %q", got.FinalURL, srv.URL+"/ok")
  }
 })

 t.Run("hash of body", func(t *testing.T) {
  got := c.Fetch(srv.URL+"/ok", "sha384")
  // sha384 of "hello", verified with:
  //   printf 'hello' | openssl dgst -sha384 -binary | openssl base64 -A
  want := "sha384-WeF0h3dEjGnea4ANejO7+5/xtGPkQ1TDVTvNucZm+pASWjx5+QOXvfX2oT3oKGhP"
  if got.Hash != want {
   t.Errorf("Hash = %q, want %q", got.Hash, want)
  }
 })

 t.Run("one request per url regardless of call count", func(t *testing.T) {
  before := atomic.LoadInt64(&hits)
  for i := 0; i < 5; i++ {
   c.Fetch(srv.URL+"/ok", "")
  }
  if after := atomic.LoadInt64(&hits); after != before {
   t.Errorf("server saw %d new requests, want 0 - results must be cached", after-before)
  }
 })

 t.Run("unreachable host is indeterminate", func(t *testing.T) {
  got := c.Fetch("https://this-host-does-not-exist.invalid/x", "")
  if got.Err == nil {
   t.Errorf("Err = nil, want an error for an unresolvable host")
  }
 })

 t.Run("allowlisted url is never requested", func(t *testing.T) {
  allowed := NewCache(Allowlist{{URL: srv.URL + "/*"}}, time.Minute, 4)
  before := atomic.LoadInt64(&hits)
  got := allowed.Fetch(srv.URL+"/ok", "")
  if after := atomic.LoadInt64(&hits); after != before {
   t.Errorf("server saw a request for an allowlisted url")
  }
  if got.Status != 0 || got.Err != nil {
   t.Errorf("Fetch() = %+v, want a zero result for an elided request", got)
  }
 })
}
```

- [ ] **Step 2: Verify the expected hash before running the suite**

The `sha384` literal above must be correct or the test is meaningless. Confirm it:

Run: `printf 'hello' | openssl dgst -sha384 -binary | openssl base64 -A`
Expected: the base64 string in the test, matching `sha384-<that>`. If it differs, replace the literal in the test with the command's output.

- [ ] **Step 3: Run test to verify it fails**

Run: `cd ~/code/thetillhoff/temingo && go test ./pkg/refcheck/ -run TestCacheFetch -v`
Expected: FAIL to build - `undefined: NewCache`.

- [ ] **Step 4: Write `pkg/refcheck/cache.go`**

```go
package refcheck

import (
 "crypto/sha256"
 "crypto/sha512"
 "encoding/base64"
 "fmt"
 "hash"
 "io"
 "net/http"
 "sync"
 "time"
)

// Result is the outcome of one URL request. A non-nil Err means the outcome is
// indeterminate - unreachable or unresolvable - rather than known-bad.
type Result struct {
 Status     int
 FinalURL   string
 AllowsCORS bool
 Hash       string
 Err        error
}

type cacheEntry struct {
 result    Result
 fetchedAt time.Time
 hashed    bool
}

// Cache requests each distinct URL at most once and retains the outcome for a
// bounded freshness window, so that repeated builds in one process - watch mode
// - do not re-request unchanged references.
type Cache struct {
 allow       Allowlist
 ttl         time.Duration
 client      *http.Client
 sem         chan struct{}
 mu          sync.Mutex
 entries     map[string]cacheEntry
 inflight    map[string]*sync.WaitGroup
 now         func() time.Time
}

// NewCache returns a cache that elides requests the allowlist fully covers and
// runs at most concurrency requests at once.
func NewCache(allow Allowlist, ttl time.Duration, concurrency int) *Cache {
 if concurrency < 1 {
  concurrency = 1
 }
 return &Cache{
  allow:    allow,
  ttl:      ttl,
  client:   &http.Client{Timeout: 15 * time.Second, CheckRedirect: stopAtFirstRedirect},
  sem:      make(chan struct{}, concurrency),
  entries:  map[string]cacheEntry{},
  inflight: map[string]*sync.WaitGroup{},
  now:      time.Now,
 }
}

// stopAtFirstRedirect keeps the first redirect visible so its target can be
// reported back to the author for repair.
func stopAtFirstRedirect(req *http.Request, via []*http.Request) error {
 return http.ErrUseLastResponse
}

// Fetch returns the outcome for rawURL, requesting it only if no fresh result
// is held. algorithm is empty when no content hash is wanted.
func (c *Cache) Fetch(rawURL, algorithm string) Result {
 // A URL the allowlist fully covers is never requested at all.
 if c.allow.AllowsEverything(rawURL) {
  return Result{}
 }

 for {
  c.mu.Lock()
  entry, ok := c.entries[rawURL]
  fresh := ok && c.now().Sub(entry.fetchedAt) < c.ttl
  // A cached result without a hash cannot answer a request that wants one.
  usable := fresh && (algorithm == "" || entry.hashed)
  if usable {
   c.mu.Unlock()
   return entry.result
  }
  if wg, running := c.inflight[rawURL]; running {
   c.mu.Unlock()
   wg.Wait()
   continue
  }
  wg := &sync.WaitGroup{}
  wg.Add(1)
  c.inflight[rawURL] = wg
  c.mu.Unlock()

  result := c.request(rawURL, algorithm)

  c.mu.Lock()
  c.entries[rawURL] = cacheEntry{result: result, fetchedAt: c.now(), hashed: algorithm != ""}
  delete(c.inflight, rawURL)
  c.mu.Unlock()
  wg.Done()

  return result
 }
}

func (c *Cache) request(rawURL, algorithm string) Result {
 c.sem <- struct{}{}
 defer func() { <-c.sem }()

 req, err := http.NewRequest(http.MethodGet, rawURL, nil)
 if err != nil {
  return Result{Err: err}
 }
 // An Origin header makes the response's CORS posture observable, which is
 // what an integrity hash on a cross-origin subresource depends on.
 req.Header.Set("Origin", "https://temingo.invalid")

 resp, err := c.client.Do(req)
 if err != nil {
  return Result{Err: err}
 }
 defer func() { _ = resp.Body.Close() }()

 result := Result{
  Status:     resp.StatusCode,
  AllowsCORS: resp.Header.Get("Access-Control-Allow-Origin") != "",
 }
 if loc := resp.Header.Get("Location"); loc != "" {
  if abs, err := resp.Request.URL.Parse(loc); err == nil {
   result.FinalURL = abs.String()
  }
 }

 if algorithm == "" {
  _, _ = io.Copy(io.Discard, resp.Body)
  return result
 }

 h, err := hasherFor(algorithm)
 if err != nil {
  return Result{Err: err}
 }
 if _, err := io.Copy(h, resp.Body); err != nil {
  return Result{Err: err}
 }
 result.Hash = algorithm + "-" + base64.StdEncoding.EncodeToString(h.Sum(nil))

 return result
}

func hasherFor(algorithm string) (hash.Hash, error) {
 switch algorithm {
 case "sha256":
  return sha256.New(), nil
 case "sha384":
  return sha512.New384(), nil
 case "sha512":
  return sha512.New(), nil
 default:
  return nil, fmt.Errorf("unsupported hash algorithm %q: use sha256, sha384 or sha512", algorithm)
 }
}
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd ~/code/thetillhoff/temingo && go test ./pkg/refcheck/ -run TestCacheFetch -v`
Expected: PASS. The "one request per url" subtest is the one that proves dedupe; if it fails, the cache is re-requesting.

- [ ] **Step 6: Commit**

```bash
cd ~/code/thetillhoff/temingo && git add pkg/refcheck/cache.go pkg/refcheck/cache_test.go && git commit -m "feat(refcheck): add deduplicating URL request cache"
```

---

### Task 8: Network checks

**Files:**

- Create: `pkg/refcheck/checkRemote.go`
- Test: `pkg/refcheck/checkRemote_test.go`

**Interfaces:**

- Consumes: `Reference`, `Finding`, `Cache` from Tasks 1, 3 and 7.
- Produces: `func CheckRemote(refs []Reference, c *Cache) []Finding`.

- [ ] **Step 1: Write the failing test**

```go
package refcheck

import (
 "net/http"
 "net/http/httptest"
 "testing"
 "time"
)

func TestCheckRemote(t *testing.T) {
 mux := http.NewServeMux()
 mux.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Access-Control-Allow-Origin", "*")
 })
 mux.HandleFunc("/nocors", func(w http.ResponseWriter, r *http.Request) {})
 mux.HandleFunc("/missing", func(w http.ResponseWriter, r *http.Request) {
  w.WriteHeader(http.StatusNotFound)
 })
 mux.HandleFunc("/gated", func(w http.ResponseWriter, r *http.Request) {
  w.WriteHeader(http.StatusForbidden)
 })
 mux.HandleFunc("/moved", func(w http.ResponseWriter, r *http.Request) {
  http.Redirect(w, r, "/ok", http.StatusMovedPermanently)
 })
 srv := httptest.NewServer(mux)
 defer srv.Close()

 tests := []struct {
  name     string
  path     string
  ref      Reference
  wantCat  Category
  wantNone bool
 }{
  {name: "200 is clean", path: "/ok", wantNone: true},
  {name: "404 is breakage", path: "/missing", wantCat: CategoryStatus},
  {name: "403 is gated", path: "/gated", wantCat: CategoryGated},
  {name: "301 reports the target", path: "/moved", wantCat: CategoryRedirect},
  {
   name: "200 without CORS breaks an integrity hash", path: "/nocors",
   ref:     Reference{Role: "script src", CanCarryIntegrity: true, HasIntegrity: true, HasCrossOrigin: true},
   wantCat: CategoryCORS,
  },
  {
   name: "200 without CORS is fine with no integrity hash", path: "/nocors",
   ref:      Reference{Role: "img src"},
   wantNone: true,
  },
 }

 for _, test := range tests {
  t.Run(test.name, func(t *testing.T) {
   ref := test.ref
   ref.File = "index.html"
   ref.URL = srv.URL + test.path
   ref.Origin = OriginRemote
   if ref.Role == "" {
    ref.Role = "a href"
   }

   got := CheckRemote([]Reference{ref}, NewCache(nil, time.Minute, 2))

   if test.wantNone {
    if len(got) != 0 {
     t.Fatalf("CheckRemote() = %+v, want no findings", got)
    }
    return
   }
   if len(got) != 1 {
    t.Fatalf("CheckRemote() = %+v, want 1 finding", got)
   }
   if got[0].Category != test.wantCat {
    t.Errorf("Category = %q, want %q", got[0].Category, test.wantCat)
   }
  })
 }
}

func TestCheckRemoteUnreachable(t *testing.T) {
 ref := Reference{File: "index.html", URL: "https://does-not-exist.invalid/x", Role: "a href", Origin: OriginRemote}

 got := CheckRemote([]Reference{ref}, NewCache(nil, time.Minute, 2))

 if len(got) != 1 || got[0].Category != CategoryUnreachable {
  t.Fatalf("CheckRemote() = %+v, want one unreachable finding", got)
 }
 if got[0].Severity != SevInfo {
  t.Errorf("Severity = %v, want SevInfo - indeterminate outcomes are informational, though strict still fails on them", got[0].Severity)
 }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd ~/code/thetillhoff/temingo && go test ./pkg/refcheck/ -run TestCheckRemote -v`
Expected: FAIL to build - `undefined: CheckRemote`.

- [ ] **Step 3: Write `pkg/refcheck/checkRemote.go`**

```go
package refcheck

import (
 "fmt"
 "net/http"
)

// CheckRemote requests every remote reference and reports what it finds. Each
// distinct URL is requested at most once, however many references share it.
func CheckRemote(refs []Reference, c *Cache) []Finding {
 var findings []Finding

 for _, r := range refs {
  if r.Origin != OriginRemote {
   continue
  }

  result := c.Fetch(r.URL, "")

  switch {
  case result.Err != nil:
   findings = append(findings, Finding{
    Ref: r, Category: CategoryUnreachable, Severity: SevInfo,
    Reason: fmt.Sprintf("could not be determined: %v", result.Err),
   })
   continue
  case result.Status == 0:
   // The request was elided by the allowlist.
   continue
  case result.Status == http.StatusUnauthorized || result.Status == http.StatusForbidden:
   findings = append(findings, Finding{
    Ref: r, Category: CategoryGated, Severity: SevNote,
    Reason: fmt.Sprintf("responded %d; expected for paywalled or login-gated targets", result.Status),
   })
  case result.Status >= 400:
   findings = append(findings, Finding{
    Ref: r, Category: CategoryStatus, Severity: SevWarn,
    Reason: fmt.Sprintf("responded %d", result.Status),
   })
  case result.Status >= 300:
   findings = append(findings, Finding{
    Ref: r, Category: CategoryRedirect, Severity: SevInfo,
    Reason: fmt.Sprintf("responded %d, redirecting to %s; replace the reference with that target", result.Status, result.FinalURL),
   })
  }

  // An integrity hash the browser cannot verify for lack of readable
  // bytes is breakage, not advice.
  if r.CanCarryIntegrity && r.HasIntegrity && result.Status < 300 && !result.AllowsCORS {
   findings = append(findings, Finding{
    Ref: r, Category: CategoryCORS, Severity: SevWarn,
    Reason: "carries an integrity hash but the target sends no Access-Control-Allow-Origin, so the browser cannot verify it and blocks the subresource",
   })
  }
 }

 return findings
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd ~/code/thetillhoff/temingo && go test ./pkg/refcheck/ -v`
Expected: PASS, whole package.

- [ ] **Step 5: Commit**

```bash
cd ~/code/thetillhoff/temingo && git add pkg/refcheck/checkRemote.go pkg/refcheck/checkRemote_test.go && git commit -m "feat(refcheck): add network checks for status, redirects and CORS"
```

---

### Task 9: Engine configuration and Render integration

**Files:**

- Modify: `pkg/temingo/Engine.go` - add `Strict`, `Allow`, `LinkCheckTTL`, `LinkCheckConcurrency`, and an unexported cache field
- Create: `pkg/temingo/checkReferences.go`
- Modify: `pkg/temingo/Render.go:167-172` - replace the `warnHTTPLinks` loop
- Delete: `pkg/temingo/warnHTTPLinks.go`, `pkg/temingo/warnHTTPLinks_test.go`
- Test: `pkg/temingo/checkReferences_test.go`

**Interfaces:**

- Consumes: everything from `pkg/refcheck`.
- Produces: `func (engine *Engine) checkReferences(rendered map[string][]byte, staticPaths []string) error`, returning a non-nil error only under `Strict` with at least one finding.

- [ ] **Step 1: Write the failing test**

```go
package temingo

import (
 "bytes"
 "log/slog"
 "strings"
 "testing"
)

func TestCheckReferences(t *testing.T) {
 tests := []struct {
  name        string
  strict      bool
  rendered    map[string][]byte
  staticPaths []string
  wantLogged  []string
  wantAbsent  []string
  wantErr     bool
 }{
  {
   name: "resolvable internal reference is silent",
   rendered: map[string][]byte{
    "index.html": []byte(`<link rel="stylesheet" href="/style.css">`),
   },
   staticPaths: []string{"style.css"},
   wantAbsent:  []string{"missing-target"},
  },
  {
   name: "mistyped internal reference is reported",
   rendered: map[string][]byte{
    "index.html": []byte(`<link rel="stylesheet" href="/styl.css">`),
   },
   staticPaths: []string{"style.css"},
   wantLogged:  []string{"missing-target", "/styl.css"},
  },
  {
   name: "directory reference resolves through an index document",
   rendered: map[string][]byte{
    "index.html":        []byte(`<a href="/slides/">Slides</a>`),
    "slides/index.html": []byte(`<p>decks</p>`),
   },
   wantAbsent: []string{"missing-target"},
  },
  {
   name: "commented-out reference is not reported",
   rendered: map[string][]byte{
    "index.html": []byte(`<nav><!-- <a href="/reading-notes/">Notes</a> --></nav>`),
   },
   wantAbsent: []string{"missing-target", "reading-notes"},
  },
  {
   name: "url in a code element is not reported",
   rendered: map[string][]byte{
    "index.html": []byte(`<code>&lt;a href="/nope/"&gt;x&lt;/a&gt;</code>`),
   },
   wantAbsent: []string{"missing-target"},
  },
  {
   name:   "strict fails the build on a finding",
   strict: true,
   rendered: map[string][]byte{
    "index.html": []byte(`<link rel="stylesheet" href="/styl.css">`),
   },
   staticPaths: []string{"style.css"},
   wantLogged:  []string{"missing-target"},
   wantErr:     true,
  },
  {
   name: "non-strict never fails the build",
   rendered: map[string][]byte{
    "index.html": []byte(`<link rel="stylesheet" href="/styl.css">`),
   },
   staticPaths: []string{"style.css"},
   wantLogged:  []string{"missing-target"},
  },
 }

 for _, test := range tests {
  t.Run(test.name, func(t *testing.T) {
   var buf bytes.Buffer
   engine := DefaultEngine()
   engine.Logger = slog.New(slog.NewTextHandler(&buf, nil))
   engine.Strict = test.strict

   err := engine.checkReferences(test.rendered, test.staticPaths)

   if test.wantErr && err == nil {
    t.Errorf("checkReferences() = nil, want an error under strict")
   }
   if !test.wantErr && err != nil {
    t.Errorf("checkReferences() = %v, want nil", err)
   }
   out := buf.String()
   for _, want := range test.wantLogged {
    if !strings.Contains(out, want) {
     t.Errorf("log does not contain %q; got:\n%s", want, out)
    }
   }
   for _, absent := range test.wantAbsent {
    if strings.Contains(out, absent) {
     t.Errorf("log unexpectedly contains %q; got:\n%s", absent, out)
    }
   }
  })
 }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd ~/code/thetillhoff/temingo && go test ./pkg/temingo/ -run TestCheckReferences -v`
Expected: FAIL to build - `engine.Strict undefined`, `engine.checkReferences undefined`.

- [ ] **Step 3: Add fields to `pkg/temingo/Engine.go`**

Add to the `Engine` struct after `Minify`:

```go
 // Strict makes any reference finding exit non-zero. It draws no
 // distinction between a definite failure and an indeterminate one: a
 // timeout is as fatal as a 404, and the remedy is to run again.
 Strict bool
 // Allow accepts findings for matching URLs.
 Allow refcheck.Allowlist
 // LinkCheckTTL bounds how long a request outcome is reused within one
 // process, so watch-mode rebuilds do not re-request every reference.
 LinkCheckTTL time.Duration
 // LinkCheckConcurrency caps simultaneous requests.
 LinkCheckConcurrency int

 linkCache *refcheck.Cache
```

Add to imports: `"time"` and `"github.com/thetillhoff/temingo/pkg/refcheck"`.

Add to the `DefaultEngine()` literal after `Minify: false,`:

```go
  Strict:               false,
  Allow:                nil,
  LinkCheckTTL:         10 * time.Minute,
  LinkCheckConcurrency: 8,
```

- [ ] **Step 4: Write `pkg/temingo/checkReferences.go`**

```go
package temingo

import (
 "fmt"
 "path"

 "github.com/thetillhoff/temingo/pkg/refcheck"
)

// checkReferences collects every reference in the rendered output and reports
// the ones that are broken, unverifiable, or address nothing the build
// produced.
//
// It runs on the in-memory rendered content and the set of paths the build will
// write, so it needs no filesystem reads and holds under a dry run. Findings
// never prevent output from being written; under Strict, any finding is
// returned as an error so the process exits non-zero.
func (engine *Engine) checkReferences(rendered map[string][]byte, staticPaths []string) error {
 var refs []refcheck.Reference

 outputPaths := make(map[string]bool, len(rendered)+len(staticPaths))
 for p := range rendered {
  outputPaths[path.Clean(p)] = true
 }
 for _, p := range staticPaths {
  outputPaths[path.Clean(p)] = true
 }

 for p, content := range rendered {
  switch path.Ext(p) {
  case ".html":
   refs = append(refs, refcheck.CollectHTML(p, content)...)
  case ".css":
   refs = append(refs, refcheck.CollectCSS(p, "css", content)...)
  }
 }

 findings := refcheck.CheckStatic(refs)
 findings = append(findings, refcheck.ResolveInternal(refs, outputPaths)...)

 if engine.linkCache == nil {
  engine.linkCache = refcheck.NewCache(engine.Allow, engine.LinkCheckTTL, engine.LinkCheckConcurrency)
 }
 findings = append(findings, refcheck.CheckRemote(refs, engine.linkCache)...)

 findings = engine.Allow.Filter(findings)
 refcheck.SortFindings(findings)

 for _, f := range findings {
  engine.Logger.Warn("Reference finding",
   "severity", f.Severity.String(),
   "file", f.Ref.File,
   "url", f.Ref.URL,
   "role", f.Ref.Role,
   "category", string(f.Category),
   "reason", f.Reason,
  )
 }

 if engine.Strict && len(findings) > 0 {
  return fmt.Errorf("%d reference findings, and strict mode is enabled", len(findings))
 }

 return nil
}
```

- [ ] **Step 5: Replace the `warnHTTPLinks` block in `pkg/temingo/Render.go`**

Replace these lines (currently at `Render.go:167-172`):

```go
 // Warn on insecure http:// links in rendered output
 for renderedTemplatePath, content := range renderedTemplates {
  if path.Ext(renderedTemplatePath) == ".html" {
   engine.warnHTTPLinks(renderedTemplatePath, content)
  }
 }
```

with:

```go
 // Check every reference in the rendered output. Findings never block the
 // write below; under Strict this returns an error after reporting them.
 if err = engine.checkReferences(renderedTemplates, staticPaths); err != nil {
  return err
 }
```

Note this returns before writing output under strict. That is a deliberate deviation from "findings never prevent output from being written": the spec's contract is that *findings* do not block, and strict is an explicit opt-in to failing. If you prefer output to be written even under strict, move the `checkReferences` call to just before `return nil` at the end of `Render()` and keep `staticPaths` in scope. Pick one and note it in the spec's post-implementation section.

- [ ] **Step 6: Delete the superseded warner**

```bash
cd ~/code/thetillhoff/temingo && git rm pkg/temingo/warnHTTPLinks.go pkg/temingo/warnHTTPLinks_test.go
```

The `http://` warning it provided is not carried forward: an insecure scheme is now visible as a remote reference like any other, and its string scan false-positived on URLs inside `<code>` and comments, which the spec forbids. If you want the insecure-scheme warning preserved, add a category for it in `CheckStatic` - a reference whose URL begins with `http://` and is not a loopback host - and a test mirroring the deleted one.

- [ ] **Step 7: Run the full suite**

Run: `cd ~/code/thetillhoff/temingo && go build ./... && go test ./... -v 2>&1 | tail -30`
Expected: PASS. `TestCheckReferences` passes and no test references the deleted warner.

- [ ] **Step 8: Commit**

```bash
cd ~/code/thetillhoff/temingo && git add -A pkg/temingo/ && git commit -m "feat: check every reference in rendered output during render"
```

---

### Task 10: `sri` template function

**Files:**

- Create: `pkg/temingo/tmpl_sri.go`
- Modify: `pkg/temingo/tmpl_funcmap.go` - the FuncMap needs access to the Engine
- Modify: `pkg/temingo/renderTemplate.go:24` and `pkg/temingo/verifyPartials.go:24` - pass the Engine to `templateFuncMap`
- Test: `pkg/temingo/tmpl_sri_test.go`

**Interfaces:**

- Consumes: `Cache` from Task 7, Engine fields from Task 9.
- Produces: `func (engine *Engine) tmplSRI(rawURL string, algorithm ...string) (string, error)`, registered as `sri`.

- [ ] **Step 1: Write the failing test**

```go
package temingo

import (
 "net/http"
 "net/http/httptest"
 "strings"
 "testing"
)

func TestTmplSRI(t *testing.T) {
 srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
  if r.URL.Path == "/gone" {
   w.WriteHeader(http.StatusNotFound)
   return
  }
  _, _ = w.Write([]byte("hello"))
 }))
 defer srv.Close()

 t.Run("default algorithm is sha384", func(t *testing.T) {
  engine := DefaultEngine()
  got, err := engine.tmplSRI(srv.URL + "/x.js")
  if err != nil {
   t.Fatalf("tmplSRI() error = %v", err)
  }
  if !strings.HasPrefix(got, "sha384-") {
   t.Errorf("tmplSRI() = %q, want a sha384- prefix", got)
  }
 })

 t.Run("explicit algorithm is honoured", func(t *testing.T) {
  engine := DefaultEngine()
  got, err := engine.tmplSRI(srv.URL+"/x.js", "sha512")
  if err != nil {
   t.Fatalf("tmplSRI() error = %v", err)
  }
  if !strings.HasPrefix(got, "sha512-") {
   t.Errorf("tmplSRI() = %q, want a sha512- prefix", got)
  }
 })

 t.Run("unknown algorithm is an error", func(t *testing.T) {
  engine := DefaultEngine()
  if _, err := engine.tmplSRI(srv.URL+"/x.js", "md5"); err == nil {
   t.Errorf("tmplSRI() error = nil, want an error for an unsupported algorithm")
  }
 })

 t.Run("local path is an error", func(t *testing.T) {
  engine := DefaultEngine()
  _, err := engine.tmplSRI("/style.css")
  if err == nil {
   t.Fatalf("tmplSRI() error = nil, want an error - the function is remote-only")
  }
  if !strings.Contains(err.Error(), "remote") {
   t.Errorf("error = %q, want it to explain that only remote URLs are supported", err)
  }
 })

 t.Run("unreachable target is a hard error", func(t *testing.T) {
  engine := DefaultEngine()
  if _, err := engine.tmplSRI("https://does-not-exist.invalid/x.js"); err == nil {
   t.Errorf("tmplSRI() error = nil, want an error - there is no correct output without a hash")
  }
 })

 t.Run("error status is a hard error", func(t *testing.T) {
  engine := DefaultEngine()
  if _, err := engine.tmplSRI(srv.URL + "/gone"); err == nil {
   t.Errorf("tmplSRI() error = nil, want an error for a 404")
  }
 })
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd ~/code/thetillhoff/temingo && go test ./pkg/temingo/ -run TestTmplSRI -v`
Expected: FAIL to build - `engine.tmplSRI undefined`.

- [ ] **Step 3: Write `pkg/temingo/tmpl_sri.go`**

```go
package temingo

import (
 "fmt"
 "net/http"

 "github.com/thetillhoff/temingo/pkg/refcheck"
)

// defaultSRIAlgorithm favours the algorithm in common recommendation over the
// weakest browsers permit.
const defaultSRIAlgorithm = "sha384"

// tmplSRI returns the integrity attribute value for a remote subresource.
//
// It is remote-only by design: hashing a file temingo itself produced protects
// nothing, because whoever can alter a same-origin file can alter the document
// carrying its hash.
//
// Failure is a hard error rather than a finding. There is no correct output
// without the hash: omitting the attribute would silently drop the protection
// the author asked for, and emitting a wrong one would break the page.
func (engine *Engine) tmplSRI(rawURL string, algorithm ...string) (string, error) {
 algo := defaultSRIAlgorithm
 if len(algorithm) > 0 && algorithm[0] != "" {
  algo = algorithm[0]
 }

 if refcheck.Classify(rawURL) != refcheck.OriginRemote {
  return "", fmt.Errorf("sri %q: only remote URLs are supported; a same-origin hash protects nothing", rawURL)
 }

 if engine.linkCache == nil {
  engine.linkCache = refcheck.NewCache(engine.Allow, engine.LinkCheckTTL, engine.LinkCheckConcurrency)
 }

 result := engine.linkCache.Fetch(rawURL, algo)
 if result.Err != nil {
  return "", fmt.Errorf("sri %q: %w", rawURL, result.Err)
 }
 if result.Status != http.StatusOK {
  return "", fmt.Errorf("sri %q: responded %d, cannot hash", rawURL, result.Status)
 }
 if result.Hash == "" {
  return "", fmt.Errorf("sri %q: no hash produced", rawURL)
 }

 return result.Hash, nil
}
```

- [ ] **Step 4: Give the FuncMap access to the Engine**

Replace `pkg/temingo/tmpl_funcmap.go` entirely:

```go
package temingo

import "text/template"

// templateFuncMap returns the functions available to templates. It takes the
// engine because some functions - sri - need engine-scoped state.
func templateFuncMap(engine *Engine) template.FuncMap {
 return template.FuncMap{
  "concat":                 tmpl_concat,
  "includeWithIndentation": tmpl_indent,
  "capitalize":             tmpl_capitalize,
  "reverse":                tmpl_reverse,
  "sortBy":                 tmpl_sortBy,
  "filterBy":               tmpl_filterBy,
  "sri":                    engine.tmplSRI,
 }
}
```

Then update both call sites to pass the engine:

- `pkg/temingo/renderTemplate.go:24`: `templateEngine = templateEngine.Funcs(templateFuncMap(engine))`
- `pkg/temingo/verifyPartials.go:24`: `temporaryTemplateEngine = temporaryTemplateEngine.Funcs(templateFuncMap(engine))`

Check both are methods on `*Engine`; if either is a free function, pass the engine in as a parameter.

- [ ] **Step 5: Run the full suite**

Run: `cd ~/code/thetillhoff/temingo && go build ./... && go test ./... 2>&1 | tail -20`
Expected: PASS. `pkg/temingo/tmpl_functions_test.go` may call `templateFuncMap()` with no argument; update it to pass a `DefaultEngine()` pointer.

- [ ] **Step 6: Commit**

```bash
cd ~/code/thetillhoff/temingo && git add pkg/temingo/ && git commit -m "feat: add sri template function for remote subresources"
```

---

### Task 11: CLI and config wiring

**Files:**

- Modify: `cmd/root.go` - add `--strict` flag and pass the new Engine fields
- Modify: `cmd/config.go` - read `strict` and `allow` from the config file
- Test: `cmd/config_test.go`

**Interfaces:**

- Consumes: `refcheck.Allowlist`, `refcheck.AllowEntry`, `refcheck.Category` from Task 6.
- Produces: `func allowlistFromConfig(config map[string]interface{}) refcheck.Allowlist`.

- [ ] **Step 1: Write the failing test**

```go
package cmd

import (
 "reflect"
 "testing"

 "github.com/thetillhoff/temingo/pkg/refcheck"
)

func TestAllowlistFromConfig(t *testing.T) {
 tests := []struct {
  name     string
  config   map[string]interface{}
  expected refcheck.Allowlist
 }{
  {
   name:     "absent key yields nothing",
   config:   map[string]interface{}{},
   expected: nil,
  },
  {
   name: "entry without checks",
   config: map[string]interface{}{
    "allow": []interface{}{
     map[string]interface{}{"url": "https://paywalled.example/*"},
    },
   },
   expected: refcheck.Allowlist{{URL: "https://paywalled.example/*"}},
  },
  {
   name: "entry with checks",
   config: map[string]interface{}{
    "allow": []interface{}{
     map[string]interface{}{
      "url":    "https://redirecting.example/*",
      "checks": []interface{}{"redirect", "status"},
     },
    },
   },
   expected: refcheck.Allowlist{{
    URL:    "https://redirecting.example/*",
    Checks: []refcheck.Category{refcheck.CategoryRedirect, refcheck.CategoryStatus},
   }},
  },
  {
   name: "malformed entries are skipped",
   config: map[string]interface{}{
    "allow": []interface{}{"not-a-map", map[string]interface{}{"nourl": "x"}},
   },
   expected: nil,
  },
 }

 for _, test := range tests {
  t.Run(test.name, func(t *testing.T) {
   got := allowlistFromConfig(test.config)
   if !reflect.DeepEqual(got, test.expected) {
    t.Errorf("allowlistFromConfig() = %+v, want %+v", got, test.expected)
   }
  })
 }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd ~/code/thetillhoff/temingo && go test ./cmd/ -run TestAllowlistFromConfig -v`
Expected: FAIL to build - `undefined: allowlistFromConfig`.

- [ ] **Step 3: Add `allowlistFromConfig` to `cmd/config.go`**

Append to the file, and add `"github.com/thetillhoff/temingo/pkg/refcheck"` to its imports:

```go
// allowlistFromConfig reads the allow list. Entries without a url are skipped;
// an entry with no checks accepts every category for its url.
func allowlistFromConfig(config map[string]interface{}) refcheck.Allowlist {
 raw, ok := config["allow"].([]interface{})
 if !ok {
  return nil
 }

 var list refcheck.Allowlist
 for _, item := range raw {
  entryMap, ok := item.(map[string]interface{})
  if !ok {
   continue
  }
  url, ok := entryMap["url"].(string)
  if !ok || url == "" {
   continue
  }

  entry := refcheck.AllowEntry{URL: url}
  if checks, ok := entryMap["checks"].([]interface{}); ok {
   for _, c := range checks {
    if s, ok := c.(string); ok {
     entry.Checks = append(entry.Checks, refcheck.Category(s))
    }
   }
  }
  list = append(list, entry)
 }

 return list
}
```

- [ ] **Step 4: Wire `strict` into `applyConfigToFlags`**

Add a `strictFlag *bool` parameter to `applyConfigToFlags` after `noDeleteOutputDirFlag`, and add this alongside the other `applyBoolFlag` calls:

```go
 applyBoolFlag("strict", "strict", strictFlag)
```

- [ ] **Step 5: Add the CLI flag and pass the fields in `cmd/root.go`**

Add to the `Flags` slice, after the `dry-run` flag:

```go
   &cli.BoolFlag{
    Name:    "strict",
    Usage:   "exit non-zero if any reference finding is reported",
    Sources: cli.EnvVars("TEMINGO_STRICT"),
   },
```

Read it alongside the other flags:

```go
   strictFlag := cmd.Bool("strict")
```

Pass it to `applyConfigToFlags` as the new final argument, then add to the `temingo.Engine` literal after `Minify: false,`:

```go
    Strict:               strictFlag,
    Allow:                allowlistFromConfig(config),
    LinkCheckTTL:         10 * time.Minute,
    LinkCheckConcurrency: 8,
```

Add `"time"` to `cmd/root.go` imports.

- [ ] **Step 6: Run the suite and check the CLI**

Run: `cd ~/code/thetillhoff/temingo && go build ./... && go test ./... 2>&1 | tail -20`
Expected: PASS.

Run: `cd ~/code/thetillhoff/temingo && go run . --help | grep -A1 strict`
Expected: the `--strict` flag is listed.

- [ ] **Step 7: Commit**

```bash
cd ~/code/thetillhoff/temingo && git add cmd/ && git commit -m "feat(cmd): add strict flag and allow list configuration"
```

---

### Task 12: End-to-end verification and documentation

**Files:**

- Create: `pkg/temingo/Render_refcheck_test.go`
- Modify: `README.md` - document `sri`, `strict`, `allow`
- Modify: `CHANGELOG.md`
- Modify: `ROADMAP.md` - remove the two delivered items
- Modify: `docs/specs/2026-07-30-link-and-sri-validation-design.md` - fill in the post-implementation section
- Modify: `TODO.md` - delete the delivered topic (only if it is tracked by git; otherwise leave it in the working tree)

**Interfaces:**

- Consumes: everything.
- Produces: no new API.

- [ ] **Step 1: Write the end-to-end test**

```go
package temingo

import (
 "bytes"
 "log/slog"
 "os"
 "path/filepath"
 "strings"
 "testing"
)

// TestRenderReportsMistypedAssetPath is the regression test for the failure this
// feature exists to catch: a stylesheet reference that points at nothing, which
// previously built cleanly and produced an unstyled page.
func TestRenderReportsMistypedAssetPath(t *testing.T) {
 dir := t.TempDir()
 src := filepath.Join(dir, "src")
 if err := os.MkdirAll(src, 0o755); err != nil {
  t.Fatal(err)
 }
 if err := os.WriteFile(filepath.Join(src, "style.css"), []byte(".a{color:red}"), 0o644); err != nil {
  t.Fatal(err)
 }
 page := `<html><head><link rel="stylesheet" href="/styl.css"></head><body><p>x</p></body></html>`
 if err := os.WriteFile(filepath.Join(src, "index.template.html"), []byte(page), 0o644); err != nil {
  t.Fatal(err)
 }

 var buf bytes.Buffer
 engine := DefaultEngine()
 engine.InputDir = src + string(filepath.Separator)
 engine.OutputDir = filepath.Join(dir, "output") + string(filepath.Separator)
 engine.TemingoignorePath = filepath.Join(dir, ".temingoignore")
 engine.Logger = slog.New(slog.NewTextHandler(&buf, nil))

 if err := engine.Render(); err != nil {
  t.Fatalf("Render() = %v, want nil - findings must not fail a non-strict build", err)
 }

 out := buf.String()
 if !strings.Contains(out, "missing-target") || !strings.Contains(out, "/styl.css") {
  t.Errorf("expected a missing-target finding for /styl.css, got:\n%s", out)
 }

 // Output must still have been written.
 if _, err := os.Stat(filepath.Join(dir, "output", "index.html")); err != nil {
  t.Errorf("output was not written: %v", err)
 }
}

// TestRenderStrictFailsOnFinding proves the CI gate works.
func TestRenderStrictFailsOnFinding(t *testing.T) {
 dir := t.TempDir()
 src := filepath.Join(dir, "src")
 if err := os.MkdirAll(src, 0o755); err != nil {
  t.Fatal(err)
 }
 page := `<html><body><a href="/nope/">x</a></body></html>`
 if err := os.WriteFile(filepath.Join(src, "index.template.html"), []byte(page), 0o644); err != nil {
  t.Fatal(err)
 }

 engine := DefaultEngine()
 engine.InputDir = src + string(filepath.Separator)
 engine.OutputDir = filepath.Join(dir, "output") + string(filepath.Separator)
 engine.TemingoignorePath = filepath.Join(dir, ".temingoignore")
 engine.Logger = slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
 engine.Strict = true

 if err := engine.Render(); err == nil {
  t.Errorf("Render() = nil, want an error under strict with a finding")
 }
}
```

- [ ] **Step 2: Run the end-to-end tests**

Run: `cd ~/code/thetillhoff/temingo && go test ./pkg/temingo/ -run 'TestRender.*' -v`
Expected: PASS. Neither test reaches the network - both references are internal.

- [ ] **Step 3: Verify against a real site**

Build and run against thetillhoff.de, which has both remote references with integrity hashes and a commented-out nav link:

```bash
cd ~/code/thetillhoff/temingo && go build -o /tmp/temingo-refcheck .
cd ~/code/thetillhoff/thetillhoff.de && /tmp/temingo-refcheck 2>&1 | grep -i 'finding' | head -20
```

Expected: no `missing-target` finding, and specifically no finding mentioning `reading-notes` - that link is inside an HTML comment, and reporting it would mean the collector is not walking elements. Remote findings depend on network reachability.

- [ ] **Step 4: Document in `README.md`**

Add a section after the existing template-functions documentation:

````markdown
### Reference checking

Every build reports references in the rendered output that are broken,
unverifiable, or point at nothing the build produced:

- external URLs that respond with an error, redirect, or require authorisation
- external URLs that respond but send no `Access-Control-Allow-Origin`, which
  makes an `integrity` hash unverifiable and causes the browser to block the
  subresource
- cross-origin scripts and stylesheets with no `integrity` hash
- an `integrity` hash with no `crossorigin` attribute, which the browser blocks
- a cross-origin `@import` from a stylesheet that carries an `integrity` hash,
  since imported sheets are not covered by it
- internal paths that no output file answers

URLs written as visible text - in a code sample, or inside an HTML comment -
are never reported.

Findings do not fail the build. Pass `--strict` (or set `strict: true`) to exit
non-zero when any finding is reported, which is the intended CI configuration.
Strict mode is fatal on unreachable and unresolvable hosts too, so a transient
network fault fails the build and the remedy is to run it again.

Accept expected findings with an allow list:

```yaml
strict: true
allow:
  - url: https://paywalled.example/*      # accept every finding for these
  - url: https://redirecting.example/*
    checks: [redirect]                     # accept only the named categories
```

A trailing `*` covers everything under it, so `https://example.com/*` matches
the whole host. A `*` in the middle of a pattern matches within one path
segment, so `https://cdn.example/lib/*/x.js` pins the filename while accepting
any version. Categories are `status`, `gated`, `redirect`, `unreachable`,
`missing-target`, `integrity`, `crossorigin`, `unverified-import` and `cors`.
A URL whose entry names no categories is never requested at all.

A redirect is best fixed by replacing the reference with its target rather than
allowlisting it.

### `sri`

Emits the integrity hash of a remote subresource:

```html
<script src="https://cdn.example/lib/5.2.1/x.js"
        integrity="{{ sri "https://cdn.example/lib/5.2.1/x.js" }}"
        crossorigin="anonymous"></script>
```

The default algorithm is `sha384`; pass another as a second argument:
`{{ sri "https://cdn.example/x.js" "sha512" }}`. Supported values are
`sha256`, `sha384` and `sha512`.

`sri` accepts remote URLs only. A hash of a file temingo produced would protect
nothing, because whoever can alter a same-origin file can alter the document
carrying its hash.

Because the hash is fetched at build time, a build using `sri` fails when the
target is unreachable - there is no correct output without the hash.

````

- [ ] **Step 5: Update `CHANGELOG.md`**

Add at the top, under a new unreleased heading:

```markdown
## Unreleased

- Check every reference in rendered output at build time: external URLs over the
  network, internal paths against the build's own output. Reports broken,
  redirecting and gated targets, missing `integrity` hashes, missing
  `crossorigin` opt-ins, unverifiable CORS, and cross-origin `@import` from a
  hash-verified stylesheet.
- Add `sri` template function, emitting the integrity hash of a remote
  subresource. `sha384` by default.
- Add `--strict` / `strict:` to exit non-zero on any finding, and `allow:` to
  accept expected ones per URL pattern and category.
- Replace the `http://` link warning: an insecure scheme is now visible as a
  remote reference, and the old string scan false-positived on URLs inside
  `<code>` and comments.
```

- [ ] **Step 6: Update `ROADMAP.md`**

Delete `- Validate internal links (warn on broken file references)` from **Soon → Output quality**, and delete `- Subresource integrity hashes (SHA256/384/512) for JS/CSS files, for use in CSP config (#92)` from **Later**.

Add to **Later**, since #92's CSP use case is not delivered:

```markdown
- Content hashes for inline `<style>` / `<script>`, so `unsafe-inline` can be dropped from a CSP (#92) — needs a delivery mechanism, since temingo does not own response headers
```

- [ ] **Step 7: Fill in the spec's post-implementation section**

Replace `To be filled in during implementation.` in `docs/specs/2026-07-30-link-and-sri-validation-design.md` with the concrete mechanisms: the `pkg/refcheck` package boundary, the in-memory output-path set rather than filesystem stats, the glob semantics of `path.Match`, the 10-minute cache window, the concurrency cap of 8, the regex-based CSS extraction and its known ceiling, and whether strict returns before or after the write phase (see Task 9 Step 5).

- [ ] **Step 8: Lint the markdown**

Run:

```bash
cd ~/code/thetillhoff/temingo && npx markdownlint-cli --fix --disable MD013 --ignore node_modules -- README.md CHANGELOG.md ROADMAP.md TODO.md docs/**/*.md
cd ~/code/thetillhoff/temingo && npx markdownlint-cli --disable MD013 --ignore node_modules -- README.md CHANGELOG.md ROADMAP.md TODO.md docs/**/*.md
```

Expected: exit 0.

- [ ] **Step 9: Full verification**

Run: `cd ~/code/thetillhoff/temingo && go build ./... && go vet ./... && go test ./... 2>&1 | tail -20`
Expected: PASS, no vet findings.

- [ ] **Step 10: Commit**

```bash
cd ~/code/thetillhoff/temingo && git add -A && git commit -m "docs: document reference checking and the sri function"
```

---

## Self-Review

**Spec coverage.** Every contract maps to a task:

| Spec contract | Task |
| --- | --- |
| Reference discovery, text and comments never references | 1, 2 |
| Reference classification | 1 |
| Internal reference resolution | 5, 9 |
| Integrity applicability | 1 (the flag), 4 (the enforcement) |
| Integrity does not transit | 4 |
| CORS opt-in equivalence | 1 (presence, not value), 4 |
| One request per URL, allowlist elision | 7 |
| Findings advisory by default, strict fatal on everything | 9 |
| Severity presentational only | 3 |
| Redirects reported for repair | 8 |
| Remote hashes load-bearing, hard error | 10 |
| Hash algorithm default and override | 10 |
| Rendering not otherwise network-dependent | 10 (remote-only guard), 9 (static checks need no network) |
| Reference / Finding / Result / Configuration data shapes | 1, 3, 6, 7, 11 |
| Boundaries | package layout across 1-8 vs 9-11 |
| Non-goals | no task builds them; local `sri` is actively rejected in 10 |

**Two deviations to decide during implementation**, both flagged inline rather than hidden:

1. Task 9 Step 5 - whether strict returns before or after output is written. The spec says findings never block writing; strict is an opt-in to failing, so either reading is defensible. Pick one, record it in the spec.
2. Task 9 Step 6 - the `http://` warning is dropped rather than ported. If you want it kept, the step says exactly how to restore it as a category.

**Type consistency.** `Reference`, `Finding`, `Category`, `Severity`, `Allowlist`, `AllowEntry`, `Result`, `Cache` are defined once and used with the same names and fields throughout. `CollectCSS` takes `(file, role string, css []byte)` in both Task 2's definition and Task 9's call. `Classify` returns `Origin` in Tasks 1 and 10. `NewCache(allow, ttl, concurrency)` is called identically in Tasks 7, 8, 9 and 10.

**One known rough edge.** Task 2 Step 4 splits an existing function in place, which is the most error-prone step in the plan - the instruction to drop the marker `defer func() {}()` line is easy to miss. Read that step twice.
