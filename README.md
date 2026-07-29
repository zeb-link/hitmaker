# Hitmaker

[![ci](https://github.com/zeb-link/hitmaker/actions/workflows/ci.yml/badge.svg)](https://github.com/zeb-link/hitmaker/actions/workflows/ci.yml)
[![npm](https://img.shields.io/npm/v/hitmaker)](https://www.npmjs.com/package/hitmaker)
[![license](https://img.shields.io/badge/license-MIT-blue)](LICENSE)

![Hitmaker](docs/intro.png)

Synthetic traffic for testing analytics, redirect services, and click-tracking
systems. Hitmaker sends realistic-looking HTTP requests to plain target URLs,
with rotating browser identity, referers, language, geo headers, and human-like
active/idle scheduling.

It is a standalone tool written in Go. Targets are opaque `http` or `https`
URLs — no product API, no auth, no coupling to anything.

It was built to exercise the analytics pipeline behind
[Zebra](https://zeblink.io), the short link operating system, and kept
standalone so it works against anything that answers HTTP.

![Hitmaker live dashboard](docs/dashboard.png)

## Install

```bash
npm i -g hitmaker
hitmaker --help
```

The npm package ships a prebuilt native binary for your platform (macOS, Linux,
and Windows on x64 or arm64). Node is used to install it, not to run it. There
is no install script, so it works under `--ignore-scripts`.

If you have Go 1.25 and would rather not go through npm:

```bash
go install github.com/zeb-link/hitmaker/v2/cmd/hitmaker@latest
```

The `/v2` is required — Go encodes the major version in the module path.

To build from a clone:

```bash
make build          # builds ./bin/hitmaker
make install-local  # symlinks it into ~/.local/bin
```

## Quick Start

Interactive TUI:

```bash
hitmaker https://example.com/a https://example.com/b
hitmaker tui links.txt
hitmaker links.txt
```

Headless CLI:

```bash
hitmaker run --for 10m --rate 5-20 --mode spoof --edge cloudflare https://example.com/a
hitmaker run --json --targets links.txt --for 30s --rate 60
```

Fast diagnostics:

```bash
hitmaker probe --factory --mode off https://example.com/a
hitmaker run --factory --for 10s --rate 60 --mode off https://example.com/a
```

Use `--factory` when you want to ignore saved `~/.hitmaker/config.json` and
`./.hitmaker.json` settings. This is useful when an old local config is set to
`proxy` mode and you just want to prove direct traffic works.

Text files accept one URL per line. Blank lines and `#` comments are ignored.

## How it works

Each target gets goroutine workers that share one keep-alive `http.Transport`,
under a total worker cap. Workers alternate between active and idle phases to
approximate human browsing rather than a flat request rate.

Every hit is assembled from a rotating identity: a User-Agent, a referer, an
`Accept-Language`, and — depending on the origin mode — a location and a fake
source IP. Returning visitors are simulated by reusing `/24` subnets from a
bounded LRU ring.

Stats flow through a snapshot struct that both the TUI and the headless output
read directly, so `--json` and the dashboard report the same numbers.

`run` exits when `--for` elapses or on Ctrl-C. Shutdown is bounded, so it won't
hang on a proxied connection that ignores cancellation.

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
| `--mode off\|auto\|spoof\|proxy` | Origin mode (see below). |
| `--edge vercel\|cloudflare` | Spoofed origin — which edge a direct (spoof) request mimics. Applies in `spoof` and `auto`. |
| `--bot-ratio 0-100` | Percent of traffic that is bots. **Same knob** as config `unknownRatio`. |
| `--bots <spec>` | Which bots (see [Bots & AI crawlers](#bots--ai-crawlers)). |
| `--device-ratio 0-100` | Of the human hits, percent desktop vs mobile. |
| `--follow` | Follow redirects (default off). |
| `--seed N` | Reproducible identities and schedule (0 = random). |
| `--json` | Stream one compact JSON snapshot per line (NDJSON). |
| `--summary` | Suppress per-interval output; print only the final totals on exit. |
| `--factory` | Ignore saved config; use built-in defaults. |

### Output for scripts and agents

```bash
hitmaker run --json --for 30s --rate 60 URL        # NDJSON: one snapshot per line
hitmaker run --json --summary --for 30s URL        # a single final totals object
hitmaker run --summary --for 30s URL               # human one-line final summary
```

`--json` is line-delimited (NDJSON) — parse it one line at a time. Pair it with
`--summary` when you only care about the totals at the end of a run. `--seed`
makes identities and the schedule reproducible, which is what you want for
fixtures.

## Modes

Origin mode controls only network origin and geo behavior. User-Agent, referer,
language, and method still rotate in every mode. Two settings work together:
**mode** decides routing (direct vs paid proxy) and **edge** decides which edge
provider's geo/IP headers a direct request carries.

| Mode | Behavior |
| --- | --- |
| `off` | Direct requests, no geo/IP spoofing headers. |
| `auto` | Public domains with valid TLDs route through the configured paid proxy provider; localhost, `.local`, IP literals, and internal/reserved names stay direct with the configured edge's geo headers. |
| `spoof` | Direct requests carrying the configured edge's geo/IP headers. |
| `proxy` | Routes through a paid proxy provider; geo spoofing headers are disabled. |

The **edge** setting (`--edge`, config `origin.edge`) is consulted by `spoof`
and by `auto`'s local/internal branch:

| Edge | Headers | Notes |
| --- | --- | --- |
| `vercel` | `x-forwarded-for`, `x-real-ip`, `x-vercel-ip-*` | Vercel passes geo as request headers, so the origin reads them directly. |
| `cloudflare` | `cf-connecting-ip`, `cf-ipcountry`, `x-hitmaker-*` | Cloudflare derives geo from `request.cf` (the real connecting IP) — a client can't set it. Your service reads the `x-hitmaker-*` geo override only in dev/non-prod, so spoofed geo registers against the dev/preview edge, never production (see below). |

`cloudflare` is the default edge (the live provider). Legacy `--mode vercel` /
`--mode cloudflare` still work — they fold to `spoof` with the matching edge.

### Testing Cloudflare spoofing locally

Vercel passes geolocation as ordinary request headers, so a direct request that
carries `x-vercel-ip-*` is enough — the origin reads what it's given. **Cloudflare
does not work this way.** Behind Cloudflare, geo lives on `request.cf`, which the
edge fills in from the *real* connecting IP, and inbound `cf-*` headers are
stripped at the edge. That means geolocation **cannot be spoofed with headers
against production** — no matter what a client sends, `request.cf` wins.

To make Cloudflare-style traffic testable, Hitmaker's `cloudflare` edge sends a
side-channel that your own service can choose to honor **in non-production only**:

| Header | Carries |
| --- | --- |
| `cf-connecting-ip` | Fake client IP (the header a Worker normally reads for IP). |
| `x-forwarded-for` | Same IP, as a fallback. |
| `cf-ipcountry` | Country code. |
| `x-hitmaker-country` | Country code. |
| `x-hitmaker-city` | City, URL-encoded (e.g. `S%C3%A3o%20Paulo`). |
| `x-hitmaker-region` | Region / subdivision. |
| `x-hitmaker-latitude` | Latitude. |
| `x-hitmaker-longitude` | Longitude. |

To wire it up on your side:

1. **Add a dev-gated override in your edge handler.** Where you currently read geo
   from `request.cf`, first check a gate that is *only ever true off production* —
   e.g. `NODE_ENV !== "production"`, or a dedicated flag/binding you set on your dev
   worker. If the gate is open **and** the spoof headers are present, read geo from
   the `x-hitmaker-*` headers instead; otherwise fall back to `request.cf`. Gate it
   so a real visitor can never forge their own location:

   ```ts
   // Real geo comes from request.cf; the override exists only for local/dev testing.
   const allowSpoof = env.NODE_ENV !== "production"; // or a dedicated dev flag/binding
   const cf = request.cf;

   let country   = cf?.country;
   let city      = cf?.city;
   let region    = cf?.region;
   let latitude  = cf?.latitude;
   let longitude = cf?.longitude;

   if (allowSpoof && request.headers.get("x-hitmaker-country")) {
     country   = request.headers.get("x-hitmaker-country")   ?? country;
     city      = decodeURIComponent(request.headers.get("x-hitmaker-city") ?? "") || city;
     region    = request.headers.get("x-hitmaker-region")    ?? region;
     latitude  = request.headers.get("x-hitmaker-latitude")  ?? latitude;
     longitude = request.headers.get("x-hitmaker-longitude") ?? longitude;
   }
   ```

   The IP needs no gate: your Worker already reads `cf-connecting-ip`, which Hitmaker
   sets — and which Cloudflare overwrites in production anyway.

2. **Run your service locally or on a dev deployment.** `wrangler dev`, or a worker
   deployed to a non-production environment, is where the gate is open. Never point
   this at production — the override is deliberately dead there.

3. **Send spoofed traffic at it** with the `cloudflare` edge:

   ```bash
   hitmaker probe --mode spoof --edge cloudflare http://127.0.0.1:8787/your-path
   hitmaker run --for 2m --rate 10 --mode spoof --edge cloudflare http://127.0.0.1:8787/your-path
   ```

   The probe header line reads `mode=spoof/cloudflare`. Confirm the faked country/city
   show up wherever your service records geo (logs, analytics, a debug endpoint) rather
   than your own real location.

Proxy support is an adapter interface; the one adapter so far is `iproyal`:

```bash
IPROYAL_URL='http://user:pass@geo.iproyal.com:12321' hitmaker run --mode proxy https://example.com
IPROYAL_URL='http://user:pass@geo.iproyal.com:12321' hitmaker run --mode auto https://example.com http://localhost:3000
```

Proxy credentials are never printed by `config print`.

## Bots & AI crawlers

Hitmaker can impersonate 78 well-known bot and AI-crawler identities (GPTBot,
ClaudeBot, PerplexityBot, Googlebot, Slackbot, curl, python-requests, and dozens
more). Each identity is sent as that agent's real, publicly documented
User-Agent string, so analytics that classify traffic by bot type will see and
categorize the hits.

List the full catalog (add `--json` for a machine-readable version):

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

Bot traffic is controlled by two separate settings:

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

A configured `0` (rate, ratio, odds) is honored rather than treated as unset, so
you can pin a value to zero.

Config files are written with mode `0600`; the global config directory uses
`0700`. Saves use the typed nested shape:

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
    "mode": "off",
    "edge": "cloudflare"
  }
}
```

Useful commands:

```bash
hitmaker config print
hitmaker config edit
hitmaker --config
hitmaker config set min_per_min 5
hitmaker config set mode spoof
hitmaker config set edge cloudflare
hitmaker config reset
```

If traffic looks dead, start here:

```bash
hitmaker config print
hitmaker probe --factory --mode off https://your-target.example/path
hitmaker run --factory --for 10s --rate 60 --mode off https://your-target.example/path
```

If those work but a normal run does not, the problem is probably saved config
or the target/proxy path rather than the traffic engine.

## TUI

```bash
hitmaker links.txt
```

![Hitmaker config editor](docs/config.png)

The dashboard is full-screen and adapts to the terminal size: a pinned header
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

The config editor uses the same full-screen frame as the dashboard and
highlights the focused field. Pressing `A` shows a preview; `Enter` there saves
the config locally and closes the editor (from the live dashboard it also
hot-reloads the running traffic). `Esc` discards edits.

The **IDENTITY** group is where bot traffic is shaped: **Bot traffic %** sets how
much of the traffic is bots, **Bot pool** sets which bots, and **Desktop share %**
splits the remaining human traffic between desktop and mobile — see
[Setting the ratio](#setting-the-ratio-two-independent-knobs).

The URL params editor supports weighted payload variants for simulating QR, UTM,
campaign, or other attribution payloads.

## Safety

Hitmaker caps total worker goroutines at `128` by default, uses a shared
keep-alive transport, drains response bodies, and bounds returning-visitor subnet
memory.

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
make release-check  # test, cross-compile, assemble and dry-run the npm packages
```

Releases are automatic: push a `v*` tag and CI publishes to npm. See
`RELEASE.md` for the details.

## Feedback

Bug reports, missing bot identities, and ideas are all welcome — open an issue
or email <support@zeblink.io>.
