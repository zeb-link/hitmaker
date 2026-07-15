# Changelog

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
