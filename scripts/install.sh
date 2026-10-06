#!/usr/bin/env bash
set -euo pipefail
repo_dir="$(cd "$(dirname "$0")/.." && pwd)"
install_dir="$HOME/.scriptor/bin"
artifact="$repo_dir/dist/scriptor"
if [ ! -f "$artifact" ]; then
    echo 'Run scripts/build.sh first.' >&2
    exit 1
fi
mkdir -p "$install_dir"
cp "$artifact" "$install_dir/scriptor"
chmod +x "$install_dir/scriptor"
echo "Installed scriptor to $install_dir"
