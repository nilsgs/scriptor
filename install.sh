#!/usr/bin/env bash
set -euo pipefail
repo_dir="$(cd "$(dirname "$0")" && pwd)"
install_dir="$HOME/.scriptor/bin"
bash "$repo_dir/scripts/build.sh"
bash "$repo_dir/scripts/install.sh"
add_to_path() {
    profile="$1"
    if [ -f "$profile" ] && grep -q '\.scriptor/bin' "$profile"; then return; fi
    printf '\n# scriptor CLI\nexport PATH="$HOME/.scriptor/bin:$PATH"\n' >> "$profile"
    echo "Added to $profile"
}
if ! printf '%s\n' "$PATH" | tr ':' '\n' | grep -Fxq "$install_dir"; then
    case "$(basename "${SHELL:-/bin/bash}")" in
        zsh) add_to_path "$HOME/.zshrc" ;;
        bash)
            if [ -f "$HOME/.bash_profile" ]; then add_to_path "$HOME/.bash_profile";
            else add_to_path "$HOME/.bashrc"; fi
            ;;
        fish)
            config="$HOME/.config/fish/conf.d/scriptor.fish"
            mkdir -p "$(dirname "$config")"
            if [ ! -f "$config" ]; then printf 'fish_add_path $HOME/.scriptor/bin\n' > "$config"; fi
            ;;
        *) echo "Add $install_dir to your PATH manually." ;;
    esac
fi
echo 'Done. Restart your terminal, then run: scriptor'
