package refcheck

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

// Each test here pins a bug found in review. They are grouped so the reason a
// behaviour exists stays attached to it.

func TestSrcsetURLsMayContainCommas(t *testing.T) {
	tests := []struct {
		name     string
		srcset   string
		expected []string
	}{
		{
			name:     "descriptor separated candidates",
			srcset:   "a.jpg 1x, b.jpg 2x",
			expected: []string{"a.jpg", "b.jpg"},
		},
		{
			name: "comma inside a transform path is part of the url",
			// Splitting on bare commas produced "/img/w_100" and "h_100/a.jpg",
			// two references that do not exist, and lost the one that does.
			srcset:   "/img/w_100,h_100/a.jpg 1x",
			expected: []string{"/img/w_100,h_100/a.jpg"},
		},
		{
			name:     "candidates without descriptors",
			srcset:   "a.jpg, b.jpg",
			expected: []string{"a.jpg", "b.jpg"},
		},
		{
			name:     "single candidate without a descriptor",
			srcset:   "a.jpg",
			expected: []string{"a.jpg"},
		},
		{
			name:     "query string containing a comma",
			srcset:   "/i/x.png?w=1,2 1x",
			expected: []string{"/i/x.png?w=1,2"},
		},
		{
			name:     "extra whitespace",
			srcset:   "  a.jpg   1x  ,   b.jpg   2x  ",
			expected: []string{"a.jpg", "b.jpg"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := splitURLs("srcset", test.srcset)
			if len(got) != len(test.expected) {
				t.Fatalf("splitURLs() = %q, want %q", got, test.expected)
			}
			for i, want := range test.expected {
				if got[i] != want {
					t.Errorf("splitURLs()[%d] = %q, want %q", i, got[i], want)
				}
			}
		})
	}
}

func TestResolveTargetDecodesOnlyOnce(t *testing.T) {
	// url.Parse already decodes. Decoding again turned a literal %20 in a
	// filename into a space, and could fail outright on a literal %, silently
	// skipping a reference instead of checking it.
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{name: "encoded percent stays encoded once", url: "/a%2520b.jpg", expected: "a%20b.jpg"},
		{name: "encoded space becomes a space", url: "/a%20b.jpg", expected: "a b.jpg"},
		{name: "literal percent survives", url: "/100%25.html", expected: "100%.html"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, ok := resolveTarget(Reference{File: "index.html", URL: test.url, Origin: OriginInternal})
			if !ok {
				t.Fatalf("resolveTarget(%q) reported no target", test.url)
			}
			if got != test.expected {
				t.Errorf("resolveTarget(%q) = %q, want %q", test.url, got, test.expected)
			}
		})
	}
}

func TestBaseElementSuppressesRelativeResolution(t *testing.T) {
	// A base element changes what every relative reference points at. The build
	// cannot resolve those, so it must not claim they are missing.
	page := `<html><head><base href="/blog/"></head><body>` +
		`<a href="post/">relative</a><img src="/img/a.jpg"></body></html>`

	refs := CollectHTML("blog/index.html", []byte(page))

	var relative, rootRelative *Reference
	for i := range refs {
		switch refs[i].URL {
		case "post/":
			relative = &refs[i]
		case "/img/a.jpg":
			rootRelative = &refs[i]
		}
	}
	if relative == nil || rootRelative == nil {
		t.Fatalf("expected both references to be collected, got %+v", refs)
	}
	if !relative.ResolutionUnknown {
		t.Errorf("relative reference should be unresolvable when a base element is declared")
	}
	if rootRelative.ResolutionUnknown {
		t.Errorf("a root-relative reference ignores the base element and stays resolvable")
	}

	// No base element declared: relative references resolve as normal.
	plain := CollectHTML("blog/index.html", []byte(`<a href="post/">x</a>`))
	if len(plain) != 1 || plain[0].ResolutionUnknown {
		t.Errorf("without a base element a relative reference must stay resolvable, got %+v", plain)
	}

	if got := ResolveInternal(refs, map[string]bool{"img/a.jpg": true}); len(got) != 0 {
		t.Errorf("ResolveInternal() = %+v, want no findings", got)
	}
}

func TestAllowlistTrailingStarStopsAtPathBoundary(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		url     string
		want    bool
	}{
		{
			name: "host prefix without a boundary must not match another host",
			// "https://cdn.example.com*" previously accepted
			// "https://cdn.example.com.attacker.test/x", suppressing findings for
			// and eliding requests to an unrelated host.
			pattern: "https://cdn.example.com*", url: "https://cdn.example.com.attacker.test/x", want: false,
		},
		{name: "boundary prefix matches below it", pattern: "https://cdn.example.com/*", url: "https://cdn.example.com/a/b", want: true},
		{name: "boundary prefix does not match another host", pattern: "https://cdn.example.com/*", url: "https://cdn.example.com.attacker.test/x", want: false},
		{
			name: "a question mark in a pasted url is literal",
			// path.Match reads ? as any-single-character, so this pattern used to
			// accept a URL that merely resembled it.
			pattern: "https://ex.example/p?id=5", url: "https://ex.example/pXid=5", want: false,
		},
		{name: "a pasted url still matches itself exactly", pattern: "https://ex.example/p?id=5", url: "https://ex.example/p?id=5", want: true},
		{name: "interior wildcard still works", pattern: "https://cdn.example/lib/*/x.js", url: "https://cdn.example/lib/5.2.1/x.js", want: true},
		{name: "empty pattern matches nothing", pattern: "", url: "https://ex.example/a", want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := matchURL(test.pattern, test.url); got != test.want {
				t.Errorf("matchURL(%q, %q) = %v, want %v", test.pattern, test.url, got, test.want)
			}
		})
	}
}

func TestCacheDoesNotRetainFailures(t *testing.T) {
	// A transient failure must not outlive itself: caching it kept a watch
	// session red until the process was restarted, and made sri unrecoverable.
	var hits int64
	var healthy atomic.Bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&hits, 1)
		if !healthy.Load() {
			panic("simulated outage")
		}
		w.Header().Set("Access-Control-Allow-Origin", "*")
		_, _ = w.Write([]byte("hello"))
	}))
	defer srv.Close()

	c := NewCache(nil)

	if got := c.Fetch(srv.URL, ""); got.Err == nil {
		t.Fatalf("expected an error during the outage, got %+v", got)
	}

	healthy.Store(true)

	got := c.Fetch(srv.URL, "")
	if got.Err != nil {
		t.Errorf("Err = %v, want nil - a recovered host must be retried", got.Err)
	}
	if got.Status != 200 {
		t.Errorf("Status = %d, want 200", got.Status)
	}
	if hits := atomic.LoadInt64(&hits); hits != 2 {
		t.Errorf("server saw %d requests, want 2 - the failure must not have been cached", hits)
	}
}

func TestCacheRejectsBadAlgorithmWithoutRequesting(t *testing.T) {
	// Validating after the request wasted it and recorded a failure against the
	// URL, so a later plain check reported an unreachable host that answered 200.
	var hits int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&hits, 1)
		_, _ = w.Write([]byte("hello"))
	}))
	defer srv.Close()

	c := NewCache(nil)

	if got := c.Fetch(srv.URL, "md5"); got.Err == nil {
		t.Errorf("expected an error for an unsupported algorithm")
	}
	if hits := atomic.LoadInt64(&hits); hits != 0 {
		t.Errorf("server saw %d requests, want 0 - the algorithm is rejected before requesting", hits)
	}

	if got := c.Fetch(srv.URL, ""); got.Err != nil || got.Status != 200 {
		t.Errorf("Fetch() = %+v, want a clean 200 - the rejected algorithm must not poison the URL", got)
	}
}

func TestCacheKeepsOneResultPerAlgorithm(t *testing.T) {
	// The entry used to be overwritten by whichever algorithm asked last, so a
	// site alternating digests re-requested every time.
	var hits int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&hits, 1)
		_, _ = w.Write([]byte("hello"))
	}))
	defer srv.Close()

	c := NewCache(nil)
	c.Fetch(srv.URL, "sha384") // 1 request
	c.Fetch(srv.URL, "")       // answered by the hashed result
	c.Fetch(srv.URL, "sha512") // 1 request, a different digest needs the body
	c.Fetch(srv.URL, "sha384") // answered from cache
	c.Fetch(srv.URL, "sha512") // answered from cache

	if hits := atomic.LoadInt64(&hits); hits != 2 {
		t.Errorf("server saw %d requests, want 2 (one per distinct algorithm)", hits)
	}
}

func TestAllowlistDoesNotElideHashRequests(t *testing.T) {
	// Allowlisting a CDN is a findings decision. Eliding its hash request left
	// sri with nothing to emit, failing the render with "responded 0".
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("hello"))
	}))
	defer srv.Close()

	c := NewCache(Allowlist{{URL: srv.URL + "/"}, {URL: srv.URL + "/*"}})

	if got := c.Fetch(srv.URL+"/x.js", ""); got.Status != 0 || got.Err != nil {
		t.Errorf("a check against an allowlisted URL should be elided, got %+v", got)
	}

	got := c.Fetch(srv.URL+"/x.js", "sha384")
	if got.Hash == "" {
		t.Errorf("hash request against an allowlisted URL produced no hash: %+v", got)
	}
}

func TestCacheRequestsProtocolRelativeURLsOverHTTPS(t *testing.T) {
	// //host/path inherits the document scheme, which a build does not have.
	// Requesting the raw string failed with "unsupported protocol scheme".
	got := NewCache(nil).Fetch("//temingo-does-not-exist.invalid/x.js", "")
	if got.Err == nil {
		t.Skip("host unexpectedly resolved; the assertion below only holds offline")
	}
	if msg := got.Err.Error(); containsAny(msg, "unsupported protocol scheme") {
		t.Errorf("Err = %v, want a transport error rather than a scheme error", got.Err)
	}
}

func TestPermitsCORSRead(t *testing.T) {
	tests := []struct {
		name string
		acao string
		want bool
	}{
		{name: "absent", acao: "", want: false},
		{name: "wildcard", acao: "*", want: true},
		{name: "our probe origin", acao: checkOrigin, want: true},
		{
			name: "some other origin is not readable from ours",
			// The build cannot know the deploy origin, so a response restricted to
			// one specific other origin must not be called verifiable.
			acao: "https://someone-else.example", want: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := permitsCORSRead(test.acao); got != test.want {
				t.Errorf("permitsCORSRead(%q) = %v, want %v", test.acao, got, test.want)
			}
		})
	}
}

// TestCheckRemoteSkipsInternalReferences pins a guard that had no coverage:
// without it every internal path would be sent over HTTP and reported
// unreachable.
func TestCheckRemoteSkipsInternalReferences(t *testing.T) {
	var hits int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&hits, 1)
	}))
	defer srv.Close()

	refs := []Reference{{File: "i.html", URL: "/style.css", Role: "link stylesheet href", Origin: OriginInternal}}

	if got := CheckRemote(refs, NewCache(nil)); len(got) != 0 {
		t.Errorf("CheckRemote() = %+v, want no findings for an internal reference", got)
	}
	if hits := atomic.LoadInt64(&hits); hits != 0 {
		t.Errorf("an internal reference was requested over HTTP")
	}
}

// TestCheckRemoteOnlyChecksCORSWhereIntegrityIsVerified pins the other untested
// guard: an image carrying a stray integrity attribute must not be reported,
// because no browser verifies one there.
func TestCheckRemoteOnlyChecksCORSWhereIntegrityIsVerified(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()

	refs := []Reference{{
		File: "i.html", URL: srv.URL, Role: "img src", Origin: OriginRemote,
		CanCarryIntegrity: false, HasIntegrity: true, HasCrossOrigin: true,
	}}

	if got := CheckRemote(refs, NewCache(nil)); len(got) != 0 {
		t.Errorf("CheckRemote() = %+v, want no findings - an image cannot carry integrity", got)
	}
}

// TestCheckRemoteRequestsEachURLOnce pins requirement 5 at the level callers use.
func TestCheckRemoteRequestsEachURLOnce(t *testing.T) {
	var hits int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&hits, 1)
	}))
	defer srv.Close()

	refs := []Reference{
		{File: "a.html", URL: srv.URL, Role: "a href", Origin: OriginRemote},
		{File: "b.html", URL: srv.URL, Role: "a href", Origin: OriginRemote},
		{File: "c.html", URL: srv.URL, Role: "a href", Origin: OriginRemote},
	}

	CheckRemote(refs, NewCache(nil))

	if hits := atomic.LoadInt64(&hits); hits != 1 {
		t.Errorf("server saw %d requests, want 1 for three references to one URL", hits)
	}
}

func TestCollectHTMLHonoursMultipleRelValues(t *testing.T) {
	// rel is a space-separated list; only checking it whole would drop integrity
	// capability for a perfectly ordinary "preload stylesheet".
	refs := CollectHTML("i.html", []byte(`<link rel="preload stylesheet" href="https://cdn.example/a.css">`))
	if len(refs) != 1 {
		t.Fatalf("CollectHTML() = %+v, want 1 reference", refs)
	}
	if !refs[0].CanCarryIntegrity {
		t.Errorf("CanCarryIntegrity = false, want true for rel=%q", "preload stylesheet")
	}
}

func containsAny(s, substr string) bool {
	return len(substr) > 0 && len(s) >= len(substr) && indexOf(s, substr) >= 0
}

func indexOf(s, substr string) int {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
