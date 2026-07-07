# AGENTS.md

This directory is the Go rewrite of Hitmaker, a standalone synthetic-traffic
generator. It is not coupled to any product API. Do not add auth, product clients,
or workspace-specific behavior; targets are opaque URLs.

## Stack

- Go 1.25
- Cobra for commands
- Bubble Tea, Bubbles, and Lip Gloss for the TUI
- Stdlib `net/http` and `encoding/json`

Use `pnpm` nowhere in this project. This is a standalone Go module.

## Commands

```bash
make fmt
make test
make vet
make build
```

`make install-local` symlinks `bin/hitmaker` into `~/.local/bin`.

## Architecture

- `internal/config`: one typed config source, env/local/global/default precedence,
  presence-aware merges so configured `0` values survive.
- `internal/identity`: user-agent, referer, language, location, and fake-IP
  generation. Keep this independent from origin mode. `bots.go` holds the
  categorized bot / AI-crawler catalog (`Bots`), the `BotCategory` values, and
  `ParseBotSpec` / `BotPool` for filtering. Add new bot identities there with the
  agent's real, publicly documented User-Agent string and its category.
- `internal/urlparams`: probabilistic URL params and weighted payload variants.
- `internal/proxy`: paid proxy provider adapters. Never log credentials.
- `internal/simulator`: worker goroutines, phase scheduling, shared transport,
  structured stats.
- `internal/tui`: dashboard and config editor models.

## Guardrails

- Keep proxy credentials redacted in output and docs.
- Preserve the total worker cap unless the user explicitly changes it.
- Do not parse logs for stats.
- Redirects are **not** followed by default — this is a redirect tester, so the
  redirect's own status is the signal and third-party destinations shouldn't be
  hit. Keep that default; `--follow` opts in.
- Shutdown is bounded (`Runner.StopAndWait`): the process must never hang past a
  `--for` deadline or Ctrl-C, even if a proxied connection ignores cancellation.
- Do not reintroduce unbounded IP/subnet memory.
- Do not add free-proxy scraping.
- Keep config files mode `0600` and global config dir mode `0700`.
