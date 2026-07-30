# CHANGELOG

## Unreleased

- Fix HTML beautifier corrupting output: text nodes are re-escaped (`&lt;div&gt;` no longer becomes a real element), `<script>`/`<style>`/`<pre>`/`<textarea>` content is no longer escaped (`.a > .b` and `a && b` stay valid), and `<pre>`/`<textarea>` content keeps its exact whitespace
- Check every reference in rendered output at build time: external URLs over the network, internal paths against the build's own output. Reports broken, redirecting and gated targets, missing `integrity` hashes, missing `crossorigin` opt-ins, unverifiable CORS, and cross-origin `@import`. URLs in visible text or HTML comments are never reported
- Add `sri` template function, emitting the integrity hash of a remote subresource. `sha384` by default, `sha256` and `sha512` on request
- Add `--strict` / `strict:` to exit non-zero on any reference finding, and `allow:` to accept expected ones per URL pattern and category
- Report references fetched over plain `http` as an `insecure-scheme` finding, with a distinct reason for subresources, which a browser blocks as mixed content on an `https` page. Loopback targets are exempt. Disable with `--allow-insecure-scheme` / `allowInsecureScheme:`. This replaces the previous `http://` warning, which scanned raw text and so reported URLs inside `<code>` and comments
- Add `--no-remote-checks` / `noRemoteChecks:` to skip every check needing a request, keeping the static and internal ones. Makes a build hermetic; does not disable `sri`, which cannot hash without fetching

## v2.6.0

- Add `filterBy` template function — filters `.childMeta` by a field value; entries missing the field are kept by default (composable with `sortBy` and `reverse`)

## v2.5.0

- Add `sortBy` template function — sorts `childMeta` by any field, returning a slice of `{"key", "value"}` entries composable with `reverse`

## v2.4.0

- Extract shared `templateFuncMap()` helper — partial verification now uses the same functions as rendering
- Warn on unused partials at build time (does not fail the build)

## v2.3.0

- Add `validateEngine()` — fail early on conflicting config (e.g. `Beautify` + `Minify`, duplicate extensions)
- Warn on insecure `http://` links in rendered HTML output (localhost/loopback excluded)
- Fail on invalid folder names containing URL-breaking characters
- Improve error messages with `%w` wrapping throughout the render pipeline

## v2.2.1

- Update dependencies

## v2.2.0

- Replace abandoned `serve` dependency with inline `net/http` file server
- Replace abandoned `gohtml` HTML beautifier with a new implementation using `golang.org/x/net/html`
- Replace abandoned `radovskyb/watcher` with `fsnotify` (via fileIO v1.1.0)
- Add ROADMAP.md and DEVELOPMENT.md
- Clean up README: replace scattered TODO sections with a pointer to ROADMAP.md

## v2.1.8

Updated dependencies.

## v2.1.7

Updated dependencies.

## v2.1.6

Updated dependencies.

## v2.1.5

Updated dependencies.

## v2.1.4

- Update GitHub Actions to use the latest action-golang-build action.
  It has CGO_ENABLED=0 set by default, so the binary is statically linked.

## v2.1.3

- Remove BUILDPLATFORM label from image references in Dockerfile.
- Remove static version reference in README.md.

## v2.1.2

- Fix Docker image build to chmod the binary to 755 and have it executable by default.
- Transition from scratch-based Docker image to alpine-based Docker image.
  Otherwise the binary has to be statically linked, and it's hard to use it as a builder image in multi-stage builds.

## v2.1.1

- Fix Homebrew tap update workflow trigger to use the correct token.

## v2.1.0

- Add from-scratch-based Docker image published to `ghcr.io/thetillhoff/temingo` for use in multi-stage Docker builds and containerized environments.
  Docker images are automatically built and published for Linux (amd64, arm64) on each release.
- Add GitHub Action (`thetillhoff/temingo`) for use in GitHub Actions workflows.
- Fix config file precedence: CLI/env flags now properly override config file values using `IsSet()` instead of comparing to defaults.

## v2.0.0

- Remove `version` subcommand, use `--version` flag instead to print the version of temingo.
- Add a basic test that the binary to be released executes correctly and prints the correct version.

## v1.1.0

- Add `Makefile` to sample project.
- Add `reverse` template function.
- Breadcrumb links now end with a slash if they don't point to a file (based on whether there's a file extension).

## v1.0.0

### Breaking Changes

- Now that the project has proper tests, it is time to release it's first stable version.
- Migrate from cobra/viper to urfave/cli v3.
- Removed the integration of a global ~/.temingo.yaml. Instead, `./.temingo.yaml` is read by default. That path can be adjusted with the `--config` flag.
- Add breadcrumbs with `Name` and `Path` fields. Breadcrumbs are now `[]Breadcrumb` structs instead of `[]string`. Each breadcrumb has both a name and a full path, enabling `{{ range .breadcrumbs }}<a href="{{ .Path }}">{{ .Name }}</a>{{ end }}` usage.

### Improvements

- Improve install.sh, with better error handling, tempdir, autodeletion of temporary files and improved log messages.
- Add `concat`, `includeWithIndentation`, and `capitalize` template functions.
- Add support for multiple `--valuesfile` flags. Multiple values files are merged in order, with later files overriding earlier ones. This allows separation of concerns (e.g., base values, environment-specific values).
- Add directory validation: checks ensure input/output directories are valid, create the output directory if missing, and automatically ignore the output directory at runtime if it's inside the input directory to prevent processing loops (with a warning shown). The ignore file itself remains unchanged.
- Added tests. Lots of tests.
- Update dependencies

## v0.6.0

- Add `--valuesfile` flag to load values from a YAML file, with CLI values taking precedence over file values. It is also added to the implicit ignore list, like the `meta.yaml` files.

## v0.5.0

- Add `--noDeleteOutputDir` flag to preserve existing output directory contents. This only overwrites the rendered template files and makes it possible to have inputDir==outputDir.
- Add `--value key=value` flag to pass custom values to templates, which are accessible in templates via `.<key>`

## v0.4.0

- `--serve` now only listens on `127.0.0.1`

## v0.3.0

- Update dependencies
- Improve install instructions
- Support ARM64 architecture

## v0.2.0

- Update dependencies
- Add `temingo version` command to print the current build version

## v0.1.1

- Don't fail on templating errors during watch mode.

## v0.1.0 on 2023-07-06

Reworked whole application for this release. Battle-tested it in the last month, and added a bunch of features, for example:

- New internal structure
- New docs
- Automatic tests
- New command syntax
- New component/partial integration
- New meta handling (i.e. lists in childfolders)
- New meta templates (i.e. same template file for all childfolders)
- Now includes webserver for development
- Can now create initial project files
- New markdown content integration (i.e. available as metadata in templates)

## v0.0.3 on 2021-09-17

- Fixed a bug, where temingo would fail if no `.temingoignore` file exists.
  From now on, it will assume nothing should be ignored in such a case.
- Restructured codebase (split from one file into multiple).

## v0.0.2 on 2021-05-17

- reworked exclusions from ground up and added support for a `.temingoignore` file
- improved debugging

## v0.0.1 on 2021-04-30

- initial release
- added github actions release workflow
