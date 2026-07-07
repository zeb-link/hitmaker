# hitmaker

Synthetic traffic for testing analytics, redirect services, and click-tracking
systems — a standalone Go tool with a Bubble Tea terminal UI and a headless CLI.

This npm package ships the **prebuilt native binary** for your platform (it is
downloaded from the GitHub release on install; the Go source is not included).

## Install

```bash
npm i -g hitmaker      # or: pnpm add -g hitmaker
hitmaker --help
```

## Quick start

```bash
hitmaker https://example.com/a          # interactive dashboard
hitmaker run --for 30s --rate 60 URL    # headless
hitmaker bots                           # list bot / AI-crawler identities
```

Full documentation, source, and release binaries:
**https://github.com/zeb-link/hitmaker**
