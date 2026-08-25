#!/bin/bash
# lib/paths.sh - Path resolution utilities
#
# Usage: source "$(dirname "$0")/lib/paths.sh"

# Check if a path is managed by dotfiles (points to DOTFILES_DIR)
# Resolves both paths to handle DOTFILES_DIR being a symlink
is_ours() {
    local path="$1"
    local real_path real_dotfiles target
    real_dotfiles=$(realpath "$DOTFILES_DIR" 2>/dev/null || echo "$DOTFILES_DIR")
    if [[ -L "$path" ]]; then
        target=$(readlink "$path")
        [[ "$target" != /* ]] && target="$(dirname "$path")/$target"
        real_path=$(realpath "$target" 2>/dev/null || echo "$target")
    elif [[ -e "$path" ]]; then
        real_path=$(realpath "$path" 2>/dev/null || echo "")
    else
        return 1
    fi
    [[ -n "$real_path" ]] && { [[ "$real_path" == "$real_dotfiles" ]] || [[ "$real_path" == "$real_dotfiles/"* ]]; }
}

# Check if path is a symlink NOT managed by dotfiles
is_foreign_symlink() {
    local path="$1"
    [[ -L "$path" ]] && ! is_ours "$path"
}

# Extract source path from symlink_map line
get_source() {
    local line="$1"
    echo "$line" | cut -d':' -f1
}

# Extract destination path from symlink_map line and expand $HOME
get_dest() {
    local line="$1"
    local dest
    dest=$(echo "$line" | cut -d':' -f2-)
    echo "${dest/\$HOME/$HOME}"
}

# Safely resolve real path
resolve_path() {
    local path="$1"
    realpath "$path" 2>/dev/null || echo ""
}
