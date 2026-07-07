# Hitmaker Redux

Synthetic traffic for testing analytics, redirect services, and click-tracking
systems. Hitmaker sends realistic-looking HTTP requests to plain target URLs,
with rotating browser identity, referers, language, geo headers, and human-like
active/idle scheduling.

It is a standalone tool. Targets are opaque `http` or `https` URLs — no product
API, no auth, no coupling to anything.

![Hitmaker live dashboard](docs/dashboard.png)

**Redux** is a from-scratch rewrite in Go (Charm/Bubble Tea TUI stack) that
replaces the original Node.js implementation. If you're looking for the old
version, it lives in this repo's git history prior to the Redux migration.

## What's new in Redux

The rewrite is not just a port — it fixes the old architecture's real defects and
adds a stack of new capabilities:

**Engine & correctness**
- **Goroutine workers** replace the fork-a-Node-process-per-URL model; one
  process, a shared keep-alive `http.Transport`, bounded worker cap.
- **Structured stats bus** — the TUI and headless output read a real snapshot
  struct instead of scraping the human-readable log (the old load-bearing smell).
- **Bounded, guaranteed shutdown** — `run` exits promptly when `--for` elapses or
  on Ctrl-C, even if a proxied connection is slow to close. No more lingering.
- **Presence-aware config merge** — a configured `0` (rate, ratio, odds) is
  honored instead of being silently replaced by a default.
- **Bounded returning-visitor memory** — the `/24`-subnet reuse set is an LRU
  ring, not an unbounded map.

**Bots & AI traffic (new)**
- A catalog of **78 bot / AI-crawler identities** you can impersonate, each sent
  as that agent's real, publicly documented User-Agent string.
- `hitmaker bots` (+ `--json`) to discover the whole catalog.
- Two clean knobs: **`--bot-ratio`** (how much traffic is bots) and **`--bots`**
  (which bots — `ai`, `crawler`, exact names, …).

**Redirect-aware (new)**
- **Does not follow 3xx by default** — reports the redirect's own status (the
  thing under test) and doesn't hammer real destination sites. `--follow` opts in.

**Origin modes**
- `none` / `vercel` geo-header spoofing / `proxy` via a **pluggable paid-provider
  adapter** system (IPRoyal first; credentials never logged). The old
  free-proxy-list feature was dropped.

**Agent- & script-friendly output (new)**
- `--json` streams **NDJSON** (one compact snapshot per line).
- `--summary` prints only the final totals object on exit.
- `--seed` makes identities and schedule **reproducible** for fixtures.
- `--device-ratio` flag (no local config file required).

**TUI**
- Full-screen, responsive dashboard that adapts to any terminal size and never
  pushes the header/footer off screen; full-width selection highlight.
- A proper config editor (Bubble Tea sub-models) replacing the old four-boolean
  nested-modal state machine, with a nested URL-params → payloads editor.

## Install

```bash
make build
make install-local
```

The binary is `bin/hitmaker`; `install-local` symlinks it to `~/.local/bin`.

## Quick Start

Interactive TUI:

```bash
hitmaker https://example.com/a https://example.com/b
hitmaker tui links.txt
hitmaker links.txt
```

Headless CLI:

```bash
hitmaker run --for 10m --rate 5-20 --mode vercel https://example.com/a
hitmaker run --json --targets links.txt --for 30s --rate 60
```

Fast diagnostics:

```bash
hitmaker probe --factory --mode none https://example.com/a
hitmaker run --factory --for 10s --rate 60 --mode none https://example.com/a
```

Use `--factory` when you want to ignore saved `~/.hitmaker/config.json` and
`./.hitmaker.json` settings. This is useful when an old local config is set to
`proxy` mode and you just want to prove direct traffic works.

Text files accept one URL per line. Blank lines and `#` comments are ignored.

### Redirects

Hitmaker does **not** follow 3xx redirects by default. It is built to test
redirect services, so the redirect's own status (e.g. `302`) is the signal — and
not following avoids sending synthetic bot traffic on to the real destination.
Pass `--follow` if you want the client to chase redirects and report the final
status instead. `probe` prints the status of the hop it stops on.

### Run flags

| Flag | Meaning |
| --- | --- |
| `--for 10m` | Run for a duration, then exit. Omit to run until Ctrl-C. |
| `--rate N` / `--rate MIN-MAX` | Hits per minute **per worker** (equals per-target at the default 1 worker). |
| `--concurrent N` | Workers per target. |
| `--mode none\|vercel\|proxy` | Origin mode (see below). |
| `--bot-ratio 0-100` | Percent of traffic that is bots. **Same knob** as config `unknownRatio`. |
| `--bots <spec>` | Which bots (see [Bots & AI crawlers](#bots--ai-crawlers)). |
| `--device-ratio 0-100` | Of the human hits, percent desktop vs mobile. |
| `--follow` | Follow redirects (default off). |
| `--seed N` | Reproducible identities and schedule (0 = random). |
| `--json` | Stream one compact JSON snapshot per line (NDJSON). |
| `--summary` | Suppress per-interval output; print only the final totals on exit. |
| `--factory` | Ignore saved config; use built-in defaults. |

`run` always exits promptly when `--for` elapses or on Ctrl-C; shutdown is
bounded, so it never lingers even if a proxied connection is slow to close.

### Output for scripts and agents

```bash
hitmaker run --json --for 30s --rate 60 URL        # NDJSON: one snapshot per line
hitmaker run --json --summary --for 30s URL        # a single final totals object
hitmaker run --summary --for 30s URL               # human one-line final summary
```

`--json` is line-delimited (NDJSON) — parse it one line at a time. Pair it with
`--summary` when you only care about the totals at the end of a run.

## Modes

Origin mode controls only network origin and geo behavior. User-Agent, referer,
language, and method still rotate in every mode.

| Mode | Behavior |
| --- | --- |
| `none` | Direct requests, no geo/IP spoofing headers. |
| `vercel` | Direct requests plus `x-forwarded-for`, `x-real-ip`, and `x-vercel-ip-*` headers. |
| `proxy` | Routes through a paid proxy provider; geo spoofing headers are disabled. |

The first proxy adapter is `iproyal`:

```bash
IPROYAL_URL='http://user:pass@geo.iproyal.com:12321' hitmaker run --mode proxy https://example.com
```

Proxy credentials are never printed by `config print`.

## Bots & AI crawlers

Hitmaker can impersonate a wide range of well-known bot and AI-crawler
identities (GPTBot, ClaudeBot, PerplexityBot, Googlebot, Slackbot, curl,
python-requests, and dozens more). Each identity is sent as that agent's real,
publicly documented User-Agent string, so any analytics that classify traffic by
bot type will see and categorize the hits — handy for exercising bot dashboards
and AI-traffic reporting on demand.

List the full catalog (add `--json` for a machine-readable version an agent can
consume):

```bash
hitmaker bots
hitmaker bots --json
```

Shape traffic with `--bots` (categories, the `ai` alias, `all`, or exact names)
and `--bot-ratio` (percent of traffic that is bots):

```bash
hitmaker run --bots ai --bot-ratio 100 https://example.com/a     # only AI bots
hitmaker run --bots GPTBot,ClaudeBot https://example.com/a       # specific names
hitmaker run --bots crawler --bot-ratio 40 https://example.com/a # 40% search spiders
hitmaker probe --bots PerplexityBot https://example.com/a        # one diagnostic hit
```

Passing `--bots` without `--bot-ratio` implies 100% bot traffic. The categories
group the catalog by kind:

| Category | Alias | What it is |
| --- | --- | --- |
| `ai_crawler` | `ai` | AI training/index crawlers (GPTBot, ClaudeBot, Bytespider, CCBot) |
| `ai_assistant` | `ai`, `assistant` | AI assistants fetching live for a user (ChatGPT-User, Claude-User) |
| `crawler` | `search` | Search & SEO spiders (Googlebot, Bingbot, AhrefsBot) |
| `fetcher` | `social` | Link-preview unfurlers (Slackbot, Twitterbot, Discordbot) |
| `cli` | | Command-line clients (curl, wget, HTTPie) |
| `library` | `lib` | HTTP client libraries (python-requests, axios, Go-http-client) |

### Setting the ratio (two independent knobs)

Bot traffic is controlled by two separate settings — this is the thing people
trip over, so it's worth stating plainly:

1. **How much** of the traffic is bots — the **`Bot traffic %`** slider in the
   config editor, the `--bot-ratio` flag, or `unknownRatio` in config. `0` = all
   human browsers, `100` = every hit is a bot. The rest of the traffic is human
   (desktop/mobile per the `Desktop share %` slider).
2. **Which** bots — the **`Bot pool`** selector in the config editor, the
   `--bots` flag, or `bots` in config. Pick a category (`ai`, `crawler`, …) or
   exact names. Within the pool, individual bots appear weighted by how common
   they are in the wild.

So "50% AI-agent traffic" = `Bot pool: AI (crawlers + assistants)` with
`Bot traffic %: 50`. "Pure GPTBot" = `Bot pool: GPTBot`, `Bot traffic %: 100`.
On the CLI, `--bots` without `--bot-ratio` implies 100%, so
`--bots ai` is shorthand for "send only AI bots".

```bash
hitmaker run --bots ai --bot-ratio 50 URL   # half AI bots, half human browsers
hitmaker config set bots ai_assistant       # default pool = live AI assistants
hitmaker config set bots GPTBot,ClaudeBot   # default pool = just these two
hitmaker config set bots all                # clear the restriction (whole catalog)
```

There is no per-category ratio (e.g. "30% crawlers + 20% assistants") — the pool
is one selection at a time, and the catalog weights decide the split inside it.

## Configuration

Precedence is:

1. Environment variables
2. Local `./.hitmaker.json`
3. Global `~/.hitmaker/config.json`
4. Defaults

Config files are written with mode `0600`; the global config directory uses
`0700`. Old flat config keys such as `MIN_PER_MIN` and `PROXY_SERVICE_URL` are
still read, but new saves use the typed nested shape:

```json
{
  "traffic": {
    "minPerMin": 1,
    "maxPerMin": 25,
    "concurrent": 1,
    "timeoutMs": 5000
  },
  "requests": {
    "method": "GET",
    "deviceRatio": 60,
    "unknownRatio": 5,
    "uniqueIpProb": 0.95,
    "urlParams": [
      { "key": "qr", "value": "1", "probability": 1 }
    ]
  },
  "schedule": {
    "minActive": 5,
    "maxActive": 15,
    "idleOdds": 0.75,
    "minIdle": 1,
    "maxIdle": 15
  },
  "origin": {
    "mode": "none"
  }
}
```

Useful commands:

```bash
hitmaker config print
hitmaker config edit
hitmaker --config
hitmaker config set min_per_min 5
hitmaker config set mode vercel
hitmaker config reset
```

If traffic looks dead, start here:

```bash
hitmaker config print
hitmaker probe --factory --mode none https://your-target.example/path
hitmaker run --factory --for 10s --rate 60 --mode none https://your-target.example/path
```

If those work but a normal run does not, the problem is probably saved config
or the target/proxy path rather than the traffic engine.

## TUI

```bash
hitmaker links.txt
```

![Hitmaker config editor](docs/config.png)

The dashboard is full-screen and adapts to any terminal size: a pinned header
(totals) on top, the target table and a live recent-hits panel in the middle,
and a pinned shortcut bar at the bottom. The recent-hits list is clamped to the
window, so it never pushes the header or footer off screen. The selected row is
shown with a full-width highlight. On narrow terminals the recent-hits panel is
hidden and the header collapses to essentials.

Keys:

| Key | Action |
| --- | --- |
| `↑` / `↓` | Select target (highlighted row) |
| `K` | Pause/resume selected target |
| `C` | Open config screen |
| `Q` / `Ctrl-C` | Quit |

Config screen:

| Key | Action |
| --- | --- |
| `↑` / `↓` | Move |
| Type numbers | Replace/edit the highlighted numeric field immediately |
| `Backspace` | Edit the highlighted numeric field |
| `←` / `→` | Nudge numbers, sliders, and selectors (incl. **Bot pool**) |
| `Tab` / `Shift-Tab` | Move between fields |
| `Enter` | Edit text fields or open nested URL params |
| `A` | Preview, then save to `./.hitmaker.json` and close |
| `G` | Save global config (stays open) |
| `L` | Save local config (stays open) |
| `D` | Restore defaults in the editor |
| `Esc` | Back / discard and close |

The config editor uses the same full-screen frame as the dashboard (title on
top, shortcut bar pinned to the bottom) and highlights the focused field so you
can see where you are as you move. Pressing `A` shows a preview; `Enter` there
saves the config locally and closes the editor (from the live dashboard it also
hot-reloads the running traffic). `Esc` discards edits.

The **IDENTITY** group is where bot traffic is shaped: **Bot traffic %** sets how
much of the traffic is bots, **Bot pool** sets which bots, and **Desktop share %**
splits the remaining human traffic between desktop and mobile — see
[Setting the ratio](#setting-the-ratio-two-independent-knobs).

The URL params editor supports weighted payload variants for simulating QR,
UTM, campaign, or other tracking payloads.

## Safety

Hitmaker caps total worker goroutines at `128` by default. It uses a shared
keep-alive `http.Transport`, drains response bodies, and bounds returning-visitor
subnet memory. Stats flow through structured snapshots, not log scraping.

Avoid sustained high traffic against local build-on-demand dev servers. Use low
rates for correctness checks and deployed/production-like targets for volume
tests.

## Development

```bash
make build          # build ./bin/hitmaker
make test           # go test ./...
make fmt            # go fmt ./...
make vet            # go vet ./...
make install-local  # symlink bin/hitmaker into ~/.local/bin
```

See `AGENTS.md` for the source-of-truth conventions for working in this repo.
