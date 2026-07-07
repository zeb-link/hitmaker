#!/usr/bin/env node
// Thin launcher: exec the prebuilt native binary the installer downloaded,
// passing through args and the real TTY (needed for the TUI).
"use strict";

const fs = require("fs");
const path = require("path");
const { spawnSync } = require("child_process");

const ext = process.platform === "win32" ? ".exe" : "";
const bin = path.join(__dirname, "..", "vendor", `hitmaker${ext}`);

if (!fs.existsSync(bin)) {
  console.error(
    "hitmaker: native binary not found. Reinstall without --ignore-scripts, " +
      "or download it from https://github.com/zeb-link/hitmaker/releases"
  );
  process.exit(1);
}

const res = spawnSync(bin, process.argv.slice(2), { stdio: "inherit" });
if (res.error) {
  console.error(res.error.message);
  process.exit(1);
}
process.exit(res.status === null ? 1 : res.status);
