#!/usr/bin/env bash
set -euo pipefail

version="${VERSION:-}"
if [[ -z "$version" ]]; then
  version="$(node -p "require('./npm/package.json').version")"
fi

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
dist="$root/dist"

rm -rf "$dist"
mkdir -p "$dist"

build_one() {
  local goos="$1"
  local goarch="$2"
  local ext=""
  if [[ "$goos" == "windows" ]]; then
    ext=".exe"
  fi
  local out="$dist/hitmaker-$goos-$goarch$ext"
  echo "building $out"
  CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" \
    go build -trimpath -ldflags "-s -w -X main.version=$version" -o "$out" ./cmd/hitmaker
}

cd "$root"
build_one darwin amd64
build_one darwin arm64
build_one linux amd64
build_one linux arm64
build_one windows amd64
build_one windows arm64

(
  cd "$dist"
  shasum -a 256 hitmaker-* > checksums.txt
)

echo "release assets for v$version written to $dist"
