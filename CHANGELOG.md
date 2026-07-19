# Changelog

## 2.3.1 - 2026-07-19

The config editor now uses the same frame as the dashboard — same header, footer,
and bordered side panel in the same spot — so opening it no longer shifts the
layout. Refreshed the README screenshots.

## 2.3.0 - 2026-07-19

Moved the TUI to Charm v2 (bubbletea/bubbles/lipgloss v2) and reworked the look.

New warm amber theme, with emerald for hits and tomato for errors. The config
editor's save step is a floating dialog over the deck instead of a full-screen
swap. Shortcut bars all share one chip style now. New intro: the wordmark burns
in over a scrolling starfield and reads "MAKING THE HITS".

Fixed the field guide wrapping text into single-word lines (it was wrapping
twice, at mismatched widths).

## 2.2.0 - 2026-07-19

Added entropy: links now behave a bit differently from each other instead of all
looking the same. Each one gets its own device mix, traffic volume, and idle
rhythm, and a few can go "viral" — near-max traffic, rarely idle.

Set the level (off/calm/chaos/mayhem, default chaos) in the config editor, with
`config set entropy`, or via `ENTROPY_*` env vars. `off` behaves like before.

Also bumped the Charm libraries within the v1 line and cobra to 1.10.2.

## 2.1.3 - 2026-07-15

### Security

- Built against Go 1.25.12, clearing **10 reachable standard-library
  vulnerabilities** in `crypto/tls`, `crypto/x509`, `net/http`, `net/url`, and
  `net/textproto` that affected every earlier binary. `go.mod` previously
  requested `go 1.25`, and releases build against whatever that resolves to.
  Also updates `golang.org/x/net` to v0.57.0 (GO-2026-4918).
  `govulncheck ./...` now reports no vulnerabilities.

### Added

- `govulncheck` runs in CI on every push and PR, plus weekly on a schedule —
  a CVE can appear with no code change on our side.
- Dependabot opens weekly PRs for Go modules and GitHub Actions.

## 2.1.2 - 2026-07-15

### Fixed

- `go install` now installs an actual release. The module path lacked the `/v2`
  suffix that Go requires once a project tags `v2.x`, so `@v2.1.1` failed
  outright and `@latest` silently ignored every release tag and installed an
  arbitrary commit instead. The module is now
  `github.com/zeb-link/hitmaker/v2` — install with:

  ```bash
  go install github.com/zeb-link/hitmaker/v2/cmd/hitmaker@latest
  ```

  npm users are unaffected; the Go module path is invisible to them.

## 2.1.1 - 2026-07-15

First release published from CI. Carries a provenance attestation.

### Fixed

- The publish script no longer aborts in CI. Trusted publishing authenticates
  via OIDC at publish time, so there is no logged-in npm user and the script's
  `npm whoami` guard treated that as a failure.

## 2.1.0 - 2026-07-15

### Changed

- **Distribution**: the native binary now ships inside the npm package instead
  of being downloaded from a GitHub release during `postinstall`. `npm i -g
  hitmaker` is unchanged, but installs no longer run a script, no longer depend
  on GitHub being reachable, and now work under `--ignore-scripts`. The binary
  is covered by the lockfile's integrity hash; the old downloader verified
  nothing. Platform binaries live in `@zeb-link/hitmaker-<platform>-<arch>`
  packages, and npm installs only the one matching the host.

### Added

- Releases publish automatically from GitHub Actions on a `v*` tag, using npm
  trusted publishing (OIDC). Published packages carry provenance attestations.
- CI runs gofmt, vet, tests, and a release-build dry run on every push and PR.

### Added (from the earlier 2.1.0 work)

- Added `auto` origin mode. Public domains with valid TLDs route through the configured paid proxy provider; localhost, `.local`, IP literals, and internal/reserved names stay direct with Vercel geo headers.
- Added public-suffix based domain classification for auto proxy routing.
- Added a contextual Field Guide panel in the config editor with per-field explanations, including detailed origin mode guidance.
- Added a brief animated Hitmaker intro banner.
- Added cross-platform release asset generation for the npm installer contract.

### Changed

- Reworked config select controls into inline radio-style controls: left/right changes values immediately, up/down navigates normally, and Enter advances to the next row.
- Aligned dashboard and config layouts so the Field Guide occupies the same right-side pane used by Recent hits.
- Embedded the npm package version into `hitmaker version` for normal builds.

### Documentation

- Documented `auto` origin mode in CLI help and README mode tables.
- Added release instructions for GitHub release assets and npm publishing.
