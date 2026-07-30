# Link and Subresource Integrity Validation

## Purpose

Rendered output accumulates references that can break without anyone noticing. Some point at things temingo does not control - CDN scripts, stylesheets, fonts, background images, third-party links - and break when the far end returns a 404, moves, or stops permitting cross-origin reads. Others point at the build's own output and break on a typo or a rename. In both cases nothing surfaces at build time, and the failure is found by a visitor, or never.

This subsystem checks every reference in the rendered output as part of the normal build, and provides a template function that pins remote subresources to a content hash so a compromised CDN cannot silently substitute different bytes.

## Featureset

- `temingo` reports broken, redirecting and access-denied external references after rendering, with the file and URL that produced each one.
- It reports references to the build's own output that address nothing the build produced - a mistyped asset path, a link to a page that was renamed or removed.
- It reports subresources that carry no integrity hash, and integrity hashes that will fail in the browser because of a missing CORS opt-in.
- References written as visible text - a URL inside a code sample, a comment - are never reported.
- A config allowlist accepts references that are expected to fail, per URL pattern, optionally narrowed to specific checks.
- Strict mode turns any finding into a non-zero exit, for use in CI.
- A template function emits the integrity hash of a remote subresource, so the attribute does not have to be maintained by hand.

## Contracts

**Reference discovery.** A reference is an attribute value on an element in the build's output, or a `url()` / `@import` target in stylesheet content. Output that is copied verbatim rather than rendered is examined too: a hand-written page or stylesheet is as capable of pointing at nothing as a generated one, and skipping it would leave most stylesheets unchecked. Text content is never a reference - a URL appearing as visible text produces no finding, regardless of which element contains it. Neither is anything inside a comment: commented-out markup does not address anything, and reporting it is noise the author cannot act on without un-commenting code they deliberately disabled. Stylesheet content is examined wherever it occurs: standalone stylesheets, embedded style blocks, and per-element style attributes.

**Reference classification.** Every reference is classified as addressing either a remote origin or the build's own output. The distinction is required regardless of what is checked, because only remote references can be requested and only cross-origin references can be subject to the integrity and CORS contracts below. References that address neither - fragment-only, and non-fetchable schemes such as mail and telephone - are not references and produce no findings.

**Internal reference resolution.** A reference addressing the build's own output resolves against the output tree, and a reference that resolves to nothing the build produced is a finding. Resolution accepts a reference that names an output file directly, and a reference that names a location served by an index document, and a reference that names a document by omitting its extension - because all three are addressable in a static tree and a generator cannot know which form a given server prefers. A reference that cannot be resolved by any of those forms is reported. Resolution is by construction hermetic: it requires no network and holds in an offline build.

Server-side rewriting is not modelled. Where a reference would only resolve through a rule the server applies and the build cannot see, resolution is inconclusive and no finding is raised - the check reports what it can prove absent, never what it merely failed to confirm.

The same restraint applies wherever the build cannot know what a reference addresses. A document may declare a base against which its relative references resolve; the build does not know what that base will mean once served, so relative references in such a document are left alone. Likewise, when the build is configured not to clear its output directory, that directory holds files this build did not write, and absence from the set of written paths proves nothing - so resolution is skipped for that build entirely rather than reporting targets that exist on disk.

**Integrity applicability.** Each reference carries whether it *can* carry an integrity hash, which is a property of the reference, not of its element type. Only references the browser will actually verify - scripts, and stylesheet-or-preload links - can. References reached from inside a stylesheet never can, because stylesheets have no integrity syntax; neither can images or embedded documents. Findings about missing or misplaced integrity are only ever raised against references that can carry it. This is a hard requirement: raising them elsewhere produces advice the author cannot act on.

**Integrity does not transit.** A stylesheet verified by hash may itself pull in unverified content, and stylesheets have no integrity syntax of their own. A cross-origin `@import` is therefore reported: it cannot be protected by a hash of its own, and the importing stylesheet's hash does not extend to it. Constraining it is a matter for the serving policy, or for self-hosting the sheet.

**CORS opt-in equivalence.** A cross-origin subresource carrying an integrity hash is fetched in a mode the browser can read only if the reference also opts into CORS. Absence of that opt-in means the browser blocks the subresource outright, so it is reported as breakage rather than as advice. An opt-in that is present but carries no value, or an unrecognised value, is equivalent to the anonymous form and is accepted - a minifier may legitimately strip the value, and must not be flagged for it. Only a fully absent opt-in is a finding.

**One request per URL.** Each distinct URL is requested at most once per build, however many pages reference it, and at most once per distinct digest where a hash is wanted. Determinate results are retained for the life of the process, so watch-mode rebuilds do not re-request unchanged references. Indeterminate outcomes are never retained: a transient failure must not outlive itself, or a single dropped packet would keep a watch session failing until the process is restarted.

A URL whose allowlist entry accepts every category is not requested at all, because nothing could be reported about it. An entry naming specific categories is requested and its findings filtered, since the categories it does not name must still be reported. A hash request is never elided, whatever the allowlist says: it is not a check, and there is no output without it.

**Checks that need a request can be declined.** Rendering is otherwise offline-capable, and a build may run somewhere with no egress, so the checks that require one can be switched off as a group. The checks that need no network are unaffected, and remain the more valuable half: they prove things about the build's own output. Declining remote checks does not decline hashing, which cannot be satisfied without fetching.

**Findings are advisory by default.** A finding is reported and the build succeeds. Under strict mode, any finding causes a non-zero exit. Strict mode makes no distinction between categories: a definite failure and an indeterminate one - a timeout, an unresolvable host - are equally fatal. The consequence is accepted: a transient network fault fails a strict build, and the remedy is to run it again.

**Findings carry no severity.** A category already states what kind of problem a finding is, and strict mode is fatal on every category equally, so a severity would order output without informing any decision. Findings are ordered by file then URL, so output is stable across builds and can be diffed.

**Redirects are reported for repair, not suppression.** A reference that redirects reports the target it was sent to, so the author can replace the reference with it. Redirects are not followed, so a chain reports its first hop rather than its eventual destination - the first hop is what the author edits. Allowlisting a redirect is possible but is not the intended remedy.

**An insecure scheme is a finding.** A reference fetched over plain `http` can be read and altered in transit, and as a subresource on an `https` page a browser blocks it outright. It is reported without any request, since the scheme is visible in the reference. Loopback targets are exempt: a local development server legitimately serves plain `http`, and there is no network for anyone to sit on. Because a site may deliberately link to a host that offers no TLS, the check can be declined as a group.

**Remote hashes are load-bearing.** The hash template function accepts a remote URL and yields an integrity attribute value. It resolves against the same per-build request budget as the checks. Failure to obtain a hash is a hard render error, not a finding: there is no correct output without it, and emitting an absent or wrong hash would either break the page or silently remove the protection the author asked for. This makes rendering depend on network reachability for any output that uses the function, in strict mode and otherwise alike. The consequence is accepted.

**Hash algorithm.** The function pins a documented default and accepts an explicit alternative. The default favours the algorithm in common recommendation over the weakest the browser permits.

**Rendering is not otherwise network-dependent.** Output that does not use the hash function renders without network access. Check findings never prevent output from being written.

## Data shapes

**Reference.** The source file it was found in; the URL as written; whether it addresses a remote origin or the build's own output; the syntactic role it was found in, sufficient to describe it back to the author; whether it can carry an integrity hash; whether an integrity hash is present; whether a CORS opt-in is present.

**Finding.** The reference it concerns; a category; and a human-readable reason. Categories are stable identifiers, because allowlist entries name them.

**Request result.** The status as answered, which for a redirect is the redirect itself; the target a redirect names; whether the response permits any caller to read it, which is what an integrity hash depends on and which cannot be answered for a response restricted to one specific other origin, since the build does not know the origin the site will be served from; the content hash, when one was requested; or an indeterminate outcome.

**Configuration.** Additions to existing configuration, each also available as a command-line flag:

```yaml
strict: true                  # any finding exits non-zero
noRemoteChecks: true          # skip the checks that need a request
allowInsecureScheme: true     # stop reporting plain http references

allow:
  - url: https://paywalled.example/*     # all findings allowed for matches
  - url: https://redirecting.example/*
    checks: [redirect]                   # only the named categories allowed
```

`url` matches by pattern, not exact string, so a versioned URL does not need a new entry each time its version changes. A pattern ending in a wildcard covers everything beneath it, but only from a path boundary: a prefix that stops mid-host must not be read as covering every host sharing that prefix. `checks` is optional and defaults to all categories.

## Boundaries

Four responsibilities, separable and independently testable:

- **Reference collection** turns rendered output into classified references. Owns the contracts that text content and comments are never references, the table of which syntactic positions carry URLs and which of those can carry integrity, and the remote-versus-own-output classification.
- **Static checking** produces findings from references alone, without network access. Owns integrity applicability, CORS opt-in equivalence, and internal reference resolution against the output tree.
- **Request and cache** turns remote URLs into request results. Owns one-request-per-URL, the freshness window, concurrency limits, and allowlist elision.
- **Hash function** exposes remote hashing to templates. Consumes request-and-cache; owns the hard-error-on-failure contract.

Static checking must not depend on request-and-cache; a build with no network reachability still produces the full static finding set, which includes every internal reference finding.

## Non-goals

- **Local subresource integrity.** Hashing output files temingo itself produced protects nothing: whoever can alter a same-origin file can alter the document that carries its hash. It becomes worthwhile only if documents and assets are served from different trust domains, which is a change of deployment architecture, not of this subsystem.
- **Content hashes for CSP or cache-busting.** Both want a related primitive - the hash of a local output file's final bytes - and both have real value, but each needs its own delivery design, and neither is required here.
- **Per-category strictness.** One gate, deliberately.
- **Cross-run hash persistence.** Nothing is written to disk to survive a cold process; a fresh build re-requests.
- **Exhaustive stylesheet parsing.** Reference extraction from stylesheet content is approximate at the edges. False positives are silenced by the allowlist.
- **Modelling the server.** Rewrites, redirects and try-file rules the deployment applies are invisible to the build, and no attempt is made to infer them. Internal resolution proves absence or stays silent.
- **Repairing anything.** No reference is rewritten, no attribute is inserted. Findings are reported; the author fixes them.

## Open questions

None outstanding.

## Post-implementation notes

*How the current build satisfies the above. Not required by the contracts.*

The four responsibilities live in `internal/refcheck`, which imports nothing from `pkg/temingo`. The engine calls into it from `Render`, after beautify/minify and before the write phase.

**Output-path set, not the filesystem.** Internal resolution is handed the keys of the rendered-template map plus the static-file paths, so it works before anything is written and under a dry run. Resolution tries the path itself, the path joined with `index.html`, and the path with `.html` appended.

**Allowlist matching** uses `path.Match`, plus a prefix comparison when the pattern ends in `*`. The prefix case is not optional: `path.Match` never lets `*` cross a `/`, so a host-wide pattern would otherwise miss every nested path.

**Cache** holds results in a map for the life of the engine, with no expiry, and requests sequentially with the mutex held for the whole call - which makes deduplication correct if callers ever stop being sequential. An entry keeps one digest per algorithm alongside the algorithm-independent status, so an unhashed request is answered by a hashed result while a second algorithm still costs a request. Results carrying an error are not stored. The requested algorithm is validated before any request is made, so a typo cannot spend a request or leave a failure recorded against the URL.

**Bodies are only read when a hash is wanted.** A check closes the response without reading it, so a link to a large file costs no transfer and cannot exhaust the client timeout mid-download.

**Requests** send an `Origin` header, which is what makes CORS posture observable, and a descriptive `User-Agent`. The latter is load-bearing rather than cosmetic: Go's default agent is answered with 403 by some hosts, which would surface as a gated-link finding manufactured by the check itself.

**Redirects are not followed** - the client returns the first response so its `Location` can be reported as the target to move to.

**Stylesheet extraction** is by regular expression rather than a tokenizer, and can match inside a comment or string literal. The allowlist is the escape hatch. A tokenizer is expected to arrive with CSS minification and beautification.

**Strict returns before the write phase.** `Render` propagates the error from the check, so under strict no output is written. The contract that findings do not block writing still holds, because strict is an explicit opt-in to failing; if writing-then-failing is preferred, moving the call to the end of `Render` satisfies the same contracts.
