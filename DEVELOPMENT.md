# Development

## Quick Start

```sh
go build          # build for current OS
make test         # run all tests
make format       # gofmt
make upgrade      # update + tidy dependencies
make list         # list all make targets
```

## Architecture

Temingo's core is the `Engine` struct (`pkg/temingo/Engine.go`) — all state, no globals. The `Render()` method (`pkg/temingo/Render.go`) orchestrates the full pipeline:

1. Parse `.temingoignore` (gitignore-style rules)
2. Validate and ensure input/output directories
3. Walk input dir, classify files by extension:
   - `.template*` → single output file
   - `.metatemplate*` → one output file per child metadata entry
   - `.partial*` → parsed and registered, no direct output
   - `meta.yaml` → loaded per directory for metadata inheritance
   - `content.md` → auto-converted to HTML, available as `.content`
   - everything else → copied as-is
4. Load and verify partials (names must be unique)
5. Execute templates with merged metadata + custom values
6. Beautify HTML output (default: enabled)

### Package Layout

| Package | Purpose |
| --------- | --------- |
| `cmd/` | CLI (urfave/cli v3), flags, config loading |
| `pkg/temingo/` | Core engine: render pipeline, template functions, init |
| `pkg/markdown2html/` | Goldmark-based Markdown → HTML |
| `pkg/mergeYaml/` | Deep-merge YAML maps and lists |
| `pkg/prettifyHTML/` | gohtml-based HTML beautification |

### Template Variables

| Variable | Type | Description |
| ---------- | ------ | ------------- |
| `.path` | string | Relative path from input dir |
| `.meta` | map | Merged metadata (child overrides parent) |
| `.childMeta` | map[string]map | Direct children's metadata, keyed by folder |
| `.breadcrumbs` | []struct{Name,Path} | Hierarchy from root to current |
| `.content` | string | HTML from `content.md`, if present |
| `.<key>` | string | Custom values from `--value` / `--valuesfile` |

### Template Functions

Beyond Go's built-ins: `concat`, `capitalize`, `includeWithIndentation`, `reverse` — all defined in `pkg/temingo/tmpl_*.go`.

### Metadata Inheritance

`meta.yaml` files are merged from the input root down to the template's directory. Child values win. The merged result is available as `.meta`; direct children's metadata is available as `.childMeta`.

## Development Workflow

```sh
# Build and run against a test project
temingo init test       # generates test project in current dir
temingo --watch --serve # rebuild on change + serve localhost:8080

# Inject version at build time
go build -ldflags="-X github.com/thetillhoff/temingo/cmd.version=v1.2.3"
```

### Adding a Template Function

1. Create `pkg/temingo/tmpl_<name>.go` with the function.
2. Register it in the template FuncMap in `pkg/temingo/Render.go`.
3. Add tests in `pkg/temingo/tmpl_functions_test.go`.
4. Document in README.md.

## Testing

```sh
make test         # go test -v ./...
go test ./...     # equivalent
```

Tests use `t.TempDir()` for isolation. The embedded `pkg/temingo/InitFiles/test/` project is used as a canonical rendering fixture. The main render tests are in `pkg/temingo/Render_test.go`.

Pre-commit hooks enforce formatting, vet, tests, tidy, and golangci-lint — run `pre-commit install` once after cloning. CI repeats `go vet ./...` and `go test -race ./...` in its `go-test` job, so a bypassed hook is caught anyway.

## Cross-Platform Builds

CI builds 6 targets: `linux`, `darwin`, `windows` × `amd64`, `arm64`. The version string is injected via `-ldflags`; when built locally it defaults to `"dev"`.

## Branch Protection

`main` is protected: `go-test` and `verify-version` are required checks, `enforce_admins` is on, and force-pushes and deletions are refused. Direct `git push origin main` is rejected for everyone, admins included - every change lands via `gh pr merge <n> --merge --delete-branch`. Renovate is unaffected: `automergeType: branch` fast-forwards `main` to a SHA whose checks already passed.

The corollary is that **no bot can commit to `main`** here, since `GITHUB_TOKEN` triggers no workflows and its commits can therefore never report the required checks. Any automation that needs a commit on `main` needs a PAT or GitHub App token - which is why the auto patch release carries its release notes as a `release_body` input instead of writing a CHANGELOG entry.

## Release Process

1. Renovate lands a dependency update on `main` → `tag-on-main.yaml` auto-bumps the patch version, tags it, and calls `release-golang-executable-on-tag.yaml` directly (a tag pushed with `GITHUB_TOKEN` does not fire `on: push: tags:`). Renovate opens no PR for these: `automergeType: branch` fast-forwards `main` once the `renovate/**` branch is green, which is why the build workflow also runs `on: push` for those branches. Major updates come as a PR and are merged by hand - and, being merged by a human, they auto-release nothing (`tag-on-main` only runs for `renovate[bot]`).
2. That workflow:
   - Builds 6 platform binaries
   - Builds and pushes multi-arch Docker image to `ghcr.io/thetillhoff/temingo`
   - Creates GitHub Release with artifacts
   - Updates Homebrew tap and GoReport card
3. Manual releases: create and push a `vX.Y.Z` tag.

The release body comes from the `CHANGELOG.md` section for the tag, and the release fails if there is none - so a manual release needs its entry written first. Automated dependency patches have no entry: `tag-on-main` passes `release_body: "Updated dependencies."` instead, and `CHANGELOG.md` skips those version numbers.
