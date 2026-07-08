# Changelog

## 2.1.0 - 2026-07-08

### Added

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
