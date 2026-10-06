#!/usr/bin/env bash
set -euo pipefail
repo_dir="$(cd "$(dirname "$0")/.." && pwd)"
version="$(tr -d '\r\n' < "$repo_dir/VERSION")"
commit="$(git -C "$repo_dir" rev-parse --short HEAD 2>/dev/null || printf 'unknown')"
mkdir -p "$repo_dir/dist"
echo "Building scriptor v${version}+${commit}..."
cd "$repo_dir/src"
go build -ldflags "-s -w -X scriptor/cmd.version=$version -X scriptor/cmd.commit=$commit" -o "$repo_dir/dist/scriptor" .
