# Link and Subresource Integrity Validation

## Purpose

Rendered output accumulates references to things temingo does not control: CDN scripts, stylesheets, fonts, background images, third-party links. When one of them breaks - a 404, a moved page, a missing CORS header - nothing surfaces at build time and the failure is discovered by a visitor, or never.

This subsystem checks every reference in the rendered output as part of the normal build, and provides a template function that pins remote subresources to a content hash so a compromised CDN cannot silently substitute different bytes.

## Featureset

- `temingo` reports broken, redirecting and access-denied references after rendering, with the file and URL that produced each one.
- It reports subresources that carry no integrity hash, and integrity hashes that will fail in the browser because of a missing CORS opt-in.
- References written as visible text - a URL inside a code sample, a comment - are never reported.
- A config allowlist accepts references that are expected to fail, per URL pattern, optionally narrowed to specific checks.
- Strict mode turns any finding into a non-zero exit, for use in CI.
- A template function emits the integrity hash of a remote subresource, so the attribute does not have to be maintained by hand.

## Contracts

**Reference discovery.** A reference is an attribute value in the rendered markup, or a `url()` / `@import` target in rendered stylesheet content. Text content is never a reference - a URL appearing as visible text produces no finding, regardless of which element contains it. Stylesheet content is examined wherever it occurs: standalone stylesheets, embedded style blocks, and per-element style attributes.

**Integrity applicability.** Each reference carries whether it *can* carry an integrity hash, which is a property of the reference, not of its element type. Only references the browser will actually verify - scripts, and stylesheet-or-preload links - can. References reached from inside a stylesheet never can, because stylesheets have no integrity syntax; neither can images or embedded documents. Findings about missing or misplaced integrity are only ever raised against references that can carry it. This is a hard requirement: raising them elsewhere produces advice the author cannot act on.

**Integrity does not transit.** A stylesheet verified by hash may itself pull in unverified content. An `@import` of a cross-origin stylesheet from a stylesheet that carries an integrity hash is therefore reported: the hash implies a protection the imported sheet does not receive.

**CORS opt-in equivalence.** A cross-origin subresource carrying an integrity hash is fetched in a mode the browser can read only if the reference also opts into CORS. Absence of that opt-in means the browser blocks the subresource outright, so it is reported as breakage rather than as advice. An opt-in that is present but carries no value, or an unrecognised value, is equivalent to the anonymous form and is accepted - a minifier may legitimately strip the value, and must not be flagged for it. Only a fully absent opt-in is a finding.

**One request per URL.** Each distinct URL is requested at most once per build, however many pages reference it. Results are retained across rebuilds within the same process for a bounded freshness window, so that watch-mode rebuilds do not re-request unchanged references. A URL matched by an allowlist entry is never requested at all - the allowlist elides the request, it does not filter the result.

**Findings are advisory by default.** A finding is reported and the build succeeds. Under strict mode, any finding causes a non-zero exit. Strict mode makes no distinction between categories: a definite failure and an indeterminate one - a timeout, an unresolvable host - are equally fatal. The consequence is accepted: a transient network fault fails a strict build, and the remedy is to run it again.

**Findings carry a severity for presentation only.** Severity orders and groups output. It does not gate strictness.

**Redirects are reported for repair, not suppression.** A reference that redirects reports its final target, so the author can replace the reference with that target. Allowlisting a redirect is possible but is not the intended remedy.

**Remote hashes are load-bearing.** The hash template function accepts a remote URL and yields an integrity attribute value. It resolves against the same per-build request budget as the checks. Failure to obtain a hash is a hard render error, not a finding: there is no correct output without it, and emitting an absent or wrong hash would either break the page or silently remove the protection the author asked for. This makes rendering depend on network reachability for any output that uses the function, in strict mode and otherwise alike. The consequence is accepted.

**Hash algorithm.** The function pins a documented default and accepts an explicit alternative. The default favours the algorithm in common recommendation over the weakest the browser permits.

**Rendering is not otherwise network-dependent.** Output that does not use the hash function renders without network access. Check findings never prevent output from being written.

## Data shapes

**Reference.** The source file it was found in; the URL as written; whether the URL is same-origin relative to the document; the syntactic role it was found in, sufficient to describe it back to the author; whether it can carry an integrity hash; whether an integrity hash is present; whether a CORS opt-in is present.

**Finding.** The reference it concerns; a category; a severity; and a human-readable reason. Categories are stable identifiers, because allowlist entries name them.

**Request result.** Final status after redirects; the final URL, when it differs from the requested one; whether the response permits cross-origin reads; the content hash, when one was requested; or an indeterminate outcome distinguishing unreachable from unresolvable.

**Configuration.** Two additions to existing configuration:

```yaml
strict: true            # any finding exits non-zero

allow:
  - url: https://paywalled.example/*     # all findings allowed for matches
  - url: https://redirecting.example/*
    checks: [status]                     # only the named categories allowed
```

`url` matches by pattern, not exact string, so a versioned URL does not need a new entry each time its version changes. `checks` is optional and defaults to all categories.

## Boundaries

Four responsibilities, separable and independently testable:

- **Reference collection** turns rendered output into references. Owns the contract that text content is never a reference, and the table of which syntactic positions carry URLs and which of those can carry integrity.
- **Static checking** produces findings from references alone, without network access. Owns integrity applicability and CORS opt-in equivalence.
- **Request and cache** turns URLs into request results. Owns one-request-per-URL, the freshness window, concurrency limits, and allowlist elision.
- **Hash function** exposes remote hashing to templates. Consumes request-and-cache; owns the hard-error-on-failure contract.

Static checking must not depend on request-and-cache; a build with no network reachability still produces the full static finding set.

## Non-goals

- **Local subresource integrity.** Hashing output files temingo itself produced protects nothing: whoever can alter a same-origin file can alter the document that carries its hash. It becomes worthwhile only if documents and assets are served from different trust domains, which is a change of deployment architecture, not of this subsystem.
- **Content hashes for CSP or cache-busting.** Both want a related primitive - the hash of a local output file's final bytes - and both have real value, but each needs its own delivery design, and neither is required here.
- **Per-category strictness.** One gate, deliberately.
- **Cross-run hash persistence.** Nothing is written to disk to survive a cold process; a fresh build re-requests.
- **Exhaustive stylesheet parsing.** Reference extraction from stylesheet content is approximate at the edges. False positives are silenced by the allowlist.
- **Repairing anything.** No reference is rewritten, no attribute is inserted. Findings are reported; the author fixes them.

## Open questions

**Should relative references be checked against the output tree?** A reference to a path temingo itself produced can be verified without network access, by asking whether the build wrote that file.

- **A:** In scope. The reference set already exists, and checking it is hermetic and cheap. Catches a mistyped path, which is otherwise invisible until someone loads the page.
- **B:** Out of scope. It is a separate roadmap item with its own edge cases - directory indexes, extensionless routes, server-side rewrites - and folding it in widens this subsystem before the external case is proven.

Recommendation: **A**, restricted to references that resolve to a literal output path, with no attempt to model server-side rewriting. Anything ambiguous is not a finding.

## Post-implementation notes

*How the current build satisfies the above. Not required by the contracts.*

To be filled in during implementation.
