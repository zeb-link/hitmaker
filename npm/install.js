#!/usr/bin/env node
// Downloads the prebuilt hitmaker binary for this platform from the matching
// GitHub release, so `npm i hitmaker` ships the native binary — not the source.
"use strict";

const fs = require("fs");
const path = require("path");
const https = require("https");
const { version } = require("./package.json");

const REPO = "zeb-link/hitmaker";
const OS = { darwin: "darwin", linux: "linux", win32: "windows" }[process.platform];
const ARCH = { x64: "amd64", arm64: "arm64" }[process.arch];

if (!OS || !ARCH) {
  console.error(`hitmaker: no prebuilt binary for ${process.platform}/${process.arch}.`);
  console.error(`Build from source or grab a binary at https://github.com/${REPO}/releases`);
  process.exit(1);
}

const ext = process.platform === "win32" ? ".exe" : "";
const asset = `hitmaker-${OS}-${ARCH}${ext}`;
const url = `https://github.com/${REPO}/releases/download/v${version}/${asset}`;
const outDir = path.join(__dirname, "vendor");
const outFile = path.join(outDir, `hitmaker${ext}`);

function download(u, dest, redirects) {
  redirects = redirects || 0;
  return new Promise((resolve, reject) => {
    if (redirects > 10) return reject(new Error("too many redirects"));
    https
      .get(u, { headers: { "User-Agent": "hitmaker-npm-installer" } }, (res) => {
        if ([301, 302, 303, 307, 308].includes(res.statusCode)) {
          res.resume();
          return resolve(download(res.headers.location, dest, redirects + 1));
        }
        if (res.statusCode !== 200) {
          res.resume();
          return reject(new Error(`HTTP ${res.statusCode} for ${u}`));
        }
        const file = fs.createWriteStream(dest, { mode: 0o755 });
        res.pipe(file);
        file.on("finish", () => file.close(() => resolve()));
        file.on("error", reject);
      })
      .on("error", reject);
  });
}

fs.mkdirSync(outDir, { recursive: true });
download(url, outFile)
  .then(() => {
    fs.chmodSync(outFile, 0o755);
    console.log(`hitmaker ${version}: installed ${asset}`);
  })
  .catch((err) => {
    console.error(`hitmaker: could not download the binary — ${err.message}`);
    console.error(`Grab it manually from https://github.com/${REPO}/releases`);
    process.exit(1);
  });
