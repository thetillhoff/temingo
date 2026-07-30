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
	result Result
	// algorithm the body was hashed with, empty if it was not hashed.
	algorithm string
}

// Cache requests each distinct URL at most once and keeps the outcome for the
// life of the process, so repeated builds - watch mode - do not re-request
// unchanged references.
//
// ponytail: sequential, no concurrency limiting, no expiry. Only the first
// build in a session pays. If that becomes too slow, fan out in CheckRemote
// rather than adding a limiter here.
type Cache struct {
	allow   Allowlist
	client  *http.Client
	mu      sync.Mutex
	entries map[string]cacheEntry
}

// NewCache returns a cache that elides requests the allowlist fully covers.
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
	// A URL the allowlist fully covers is never requested at all.
	if c.allow.AllowsEverything(rawURL) {
		return Result{}
	}

	c.mu.Lock()
	entry, ok := c.entries[rawURL]
	// A held result answers the request unless a hash is wanted that it does
	// not carry.
	if ok && (algorithm == "" || entry.algorithm == algorithm) {
		c.mu.Unlock()
		return entry.result
	}
	c.mu.Unlock()

	result := c.request(rawURL, algorithm)

	c.mu.Lock()
	c.entries[rawURL] = cacheEntry{result: result, algorithm: algorithm}
	c.mu.Unlock()

	return result
}

func (c *Cache) request(rawURL, algorithm string) Result {
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
