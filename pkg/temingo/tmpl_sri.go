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
		engine.linkCache = refcheck.NewCache(engine.Allow)
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
