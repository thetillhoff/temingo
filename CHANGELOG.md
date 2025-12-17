# CHANGELOG

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
- Restructed codebase (split from one file into multiple).

## v0.0.2 on 2021-05-17

- reworked exlusions from ground up and added support for a `.temingoignore` file
- improved debugging

## v0.0.1 on 2021-04-30

- initial release
- added github actions release workflow
