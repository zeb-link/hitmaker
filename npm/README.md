# hitmaker

Synthetic traffic for testing analytics, redirect services, and click-tracking
systems — a standalone Go tool with a Bubble Tea terminal UI and a headless CLI.

We built it to exercise the analytics pipeline behind
[Zebra](https://zeblink.io), the short link operating system, and kept it
standalone so it works against anything that answers HTTP.

This package ships a **prebuilt native binary** (Go) for your platform. Node is
used to install it, not to run it. The binary lives in a per-platform package
that npm resolves automatically; the Go source is not included.

There is no install script — nothing is downloaded at install time, so it works
under `--ignore-scripts`.

## Install

```bash
npm i -g hitmaker      # or: pnpm add -g hitmaker
hitmaker --help
```

## Supported platforms

macOS, Linux, and Windows on x64 or arm64.

## Quick start

```bash
hitmaker https://example.com/a          # interactive dashboard
hitmaker run --for 30s --rate 60 URL    # headless
hitmaker bots                           # list bot / AI-crawler identities
```

Full documentation, source, and release binaries:
**https://github.com/zeb-link/hitmaker**

Bug reports and ideas are welcome — open an issue or email <support@zeblink.io>.
