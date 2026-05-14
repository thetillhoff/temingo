# Roadmap

## Now

Reliability and code quality — fix before adding features.

- Move FuncMap registration into a shared helper (currently duplicated between `templateEngine` and `temporaryTemplateEngine`)
- Move `getMetaForDir` call into the main `Render()` pipeline (currently buried inside `renderTemplate()`)
- Increase unit test coverage — rendering is currently only tested manually (#27)

## Soon

High-value improvements to the daily dev loop.

**Watch mode**

- Partial rerender: only rebuild files affected by a changed source; reflect only those changes in logs
- Hash-based change detection: skip re-writing rendered files and static files whose contents haven't changed; initialise hash table from existing output on first run if output dir is non-empty
- WebSocket push to browser on rebuild — auto-inject the client library into output so no manual setup is needed

**Authoring**

- Auto-indent multiline partials to match their `{{ template }}` call-site indentation (configurable, default on)
- Hierarchical `values.yaml` auto-discovery: walk from root to template dir, merge in order, nesting/chaining supported (#24)
- Detect unused partials and unused values, warn at build time (#12)

**Output quality**

- Save source-file + line-number mapping during rendering so validation errors point back to the original template, not the generated output
- Validate internal links (warn on broken file references)
- Auto-generate `sitemap.xml` (#26)
- Auto-generate `feed.xml` / RSS feed

## Later

Useful but not urgent.

- Subresource integrity hashes (SHA256/384/512) for JS/CSS files, for use in CSP config (#92)
- Use template variables inside `content.md` (uncertain value, explore first)
- Global template variables: `renderTime` etc.
- File extension autodiscover: make explicit extension config optional; minimum coverage is `.html`, `.css`, `.js`; stretch goal `.svg` with auto-inline or color-variant pregeneration
- CSS and JS beautification (currently only HTML)
- Image optimization and WebP conversion with thumbnail support (#11, #13)

## Someday

Large scope, speculative, or better as separate tools.

- Minification of HTML/CSS/JS (#93, #6) — likely its own binary; HTML minification should warn on undefined CSS classes; CSS minification should warn on unused classes; `div`-merging on HTML minify (note: may conflict with CSS rules)
- Content validation: check generated output is valid HTML/CSS/JS per file extension (#10) — likely its own binary; prettify/minify/validate packages should live in dedicated repos
- Component system with argument passing (spec commented out in README)
- Component library / dependency management (#16, #29): `component.yaml` referencing git repos + tags; global registry (helm/godocs/apt style); local overrides remain possible; per-library `values.yaml` for default values; print CSS dependency and override tree per component
- Use HTML `<meta>` tags as listview attributes
