# CHANGELOG

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
