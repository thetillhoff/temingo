package refcheck

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"hash"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// userAgent identifies temingo to the hosts it checks, so operators can see who
// is asking and so hosts that reject Go's default agent do not turn into
// findings.
const userAgent = "temingo (+https://github.com/thetillhoff/temingo)"

// Result is the outcome of one URL request. A non-nil Err means the outcome is
// indeterminate - unreachable or unresolvable - rather than known-bad.
type Result struct {
	Status     int
	FinalURL   string
	AllowsCORS bool
	Hash       string
	Err        error
}

// cacheEntry holds what is known about one URL. The status fields do not depend
// on which digest was asked for, so they are shared; hashes are per algorithm,
// because computing one requires reading the body.
type cacheEntry struct {
	result Result
	hashes map[string]string
}

// Cache requests each distinct URL at most once and keeps the outcome for the
// life of the process, so repeated builds - watch mode - do not re-request
// unchanged references.
//
// Indeterminate outcomes are never cached: a transient failure must not outlive
// itself, or one dropped packet would keep a watch session red until the process
// is restarted.
//
// ponytail: the mutex is held across the request, which serialises all fetches.
// That is free today because callers are sequential, and it makes deduplication
// correct if they ever stop being. Replace it with per-URL single-flight if
// concurrent fetching is ever wanted.
type Cache struct {
	allow   Allowlist
	client  *http.Client
	mu      sync.Mutex
	entries map[string]cacheEntry
}

// NewCache returns a cache that elides checks the allowlist fully covers.
func NewCache(allow Allowlist) *Cache {
	return &Cache{
		allow:   allow,
		client:  &http.Client{Timeout: 15 * time.Second, CheckRedirect: stopAtFirstRedirect},
		entries: map[string]cacheEntry{},
	}
}

// stopAtFirstRedirect keeps the first redirect visible so its target can be
// reported back to the author for repair.
func stopAtFirstRedirect(req *http.Request, via []*http.Request) error {
	return http.ErrUseLastResponse
}

// Fetch returns the outcome for rawURL, requesting it only if no usable result
// is held. algorithm is empty when no content hash is wanted.
func (c *Cache) Fetch(rawURL, algorithm string) Result {
	// Reject an unusable algorithm before spending a request, so a typo cannot
	// consume network traffic or leave a failure recorded against the URL.
	if algorithm != "" {
		if _, err := hasherFor(algorithm); err != nil {
			return Result{Err: err}
		}
	}

	// A URL the allowlist fully covers needs no checking. A hash request is not
	// a check, though - there is no output without it - so the allowlist must
	// not elide one.
	if algorithm == "" && c.allow.AllowsEverything(rawURL) {
		return Result{}
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.entries[rawURL]
	if ok {
		if algorithm == "" {
			return entry.result
		}
		if h, held := entry.hashes[algorithm]; held {
			result := entry.result
			result.Hash = h
			return result
		}
		// The status is known but this digest is not, and computing it needs the
		// body. Fall through to a request.
	}

	result := c.request(rawURL, algorithm)

	// An indeterminate outcome says nothing durable about the URL, so it is not
	// recorded - the next build asks again.
	if result.Err == nil {
		if !ok {
			entry = cacheEntry{hashes: map[string]string{}}
		}
		entry.result = Result{
			Status:     result.Status,
			FinalURL:   result.FinalURL,
			AllowsCORS: result.AllowsCORS,
		}
		if algorithm != "" && result.Hash != "" {
			entry.hashes[algorithm] = result.Hash
		}
		c.entries[rawURL] = entry
	}

	return result
}

func (c *Cache) request(rawURL, algorithm string) Result {
	// A protocol-relative URL inherits the document's scheme, which a build does
	// not have. https is the only defensible assumption, and requesting the raw
	// string would fail with "unsupported protocol scheme".
	requestURL := rawURL
	if strings.HasPrefix(requestURL, "//") {
		requestURL = "https:" + requestURL
	}

	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		return Result{Err: err}
	}
	// An Origin header makes the response's CORS posture observable, which is
	// what an integrity hash on a cross-origin subresource depends on.
	req.Header.Set("Origin", checkOrigin)
	// Go's default User-Agent is blocked outright by some hosts - Wikipedia
	// answers it with 403 - which would be reported as a gated link even though
	// the reference is fine in a browser. A descriptive agent avoids inventing
	// findings out of our own request.
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.client.Do(req)
	if err != nil {
		return Result{Err: err}
	}
	defer func() { _ = resp.Body.Close() }()

	result := Result{
		Status:     resp.StatusCode,
		AllowsCORS: permitsCORSRead(resp.Header.Get("Access-Control-Allow-Origin")),
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

// checkOrigin is the origin declared when probing CORS posture. A build does not
// know the origin the site will be served from, so responses restricted to one
// specific origin cannot be judged - see permitsCORSRead.
const checkOrigin = "https://temingo.invalid"

// permitsCORSRead reports whether an Access-Control-Allow-Origin value lets any
// caller read the response. A value naming one specific origin is not treated as
// permissive, because the build cannot know the origin the site is served from
// and guessing would report a subresource as verifiable when the browser will
// block it.
func permitsCORSRead(acao string) bool {
	switch strings.TrimSpace(acao) {
	case "":
		return false
	case "*", checkOrigin:
		return true
	default:
		return false
	}
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
