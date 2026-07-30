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
	// CategoryMissingIntegrity is a verifiable subresource carrying no
	// integrity hash.
	CategoryMissingIntegrity Category = "missing-integrity"
	// CategoryMissingCrossOrigin is an integrity hash the browser cannot verify
	// because the reference does not opt into CORS.
	CategoryMissingCrossOrigin Category = "missing-crossorigin"
	// CategoryNoCORSHeader is an integrity hash the browser cannot verify
	// because the target does not permit cross-origin reads.
	CategoryNoCORSHeader Category = "no-cors-header"
	// CategoryUnverifiedImport is a cross-origin stylesheet imported by a
	// stylesheet that carries an integrity hash.
	CategoryUnverifiedImport Category = "unverified-import"
)

// Finding is one problem with one reference.
//
// There is deliberately no severity: Category already says what kind of problem
// a finding is, and strict mode is fatal on all of them equally, so a severity
// would order output without informing anything.
type Finding struct {
	Ref      Reference
	Category Category
	Reason   string
}

func (f Finding) String() string {
	return fmt.Sprintf("%s: %s %s: %s: %s",
		f.Ref.File, f.Ref.Role, f.Ref.URL, f.Category, f.Reason)
}

// SortFindings orders findings by file then URL, so output is stable across
// builds even though the rendered-file map is iterated in random order.
func SortFindings(fs []Finding) {
	sort.SliceStable(fs, func(i, j int) bool {
		if fs[i].Ref.File != fs[j].Ref.File {
			return fs[i].Ref.File < fs[j].Ref.File
		}
		return fs[i].Ref.URL < fs[j].Ref.URL
	})
}
