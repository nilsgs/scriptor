#!/usr/bin/env bash
set -euo pipefail
repo_dir="$(cd "$(dirname "$0")/.." && pwd)"
version="$(tr -d '\r\n' < "$repo_dir/VERSION")"
commit="$(git -C "$repo_dir" rev-parse --short HEAD 2>/dev/null || printf 'unknown')"
mkdir -p "$repo_dir/dist"
cd "$repo_dir/src"
for os in linux darwin windows; do
    for arch in amd64 arm64; do
        name="scriptor-$os-$arch"
        if [ "$os" = windows ]; then name="$name.exe"; fi
        echo "Building $name..."
        CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" go build \
            -ldflags "-s -w -X scriptor/cmd.version=$version -X scriptor/cmd.commit=$commit" \
            -o "$repo_dir/dist/$name" .
    done
done
