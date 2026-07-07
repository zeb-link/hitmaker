#!/usr/bin/env sh
set -eu

install_dir="${HITMAKER_INSTALL_DIR:-$HOME/.local/bin}"
repo_dir="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
binary="$repo_dir/bin/hitmaker"

if [ ! -x "$binary" ]; then
  (cd "$repo_dir" && make build)
fi

mkdir -p "$install_dir"
ln -sf "$binary" "$install_dir/hitmaker"
# `hm` is a short alias for the same binary.
ln -sf "$binary" "$install_dir/hm"
printf 'hitmaker installed at %s (alias: hm)\n' "$install_dir/hitmaker"
