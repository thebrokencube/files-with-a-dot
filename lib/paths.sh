#!/bin/bash
# lib/paths.sh - Path resolution utilities
#
# Usage: source "$(dirname "$0")/lib/paths.sh"

# Check if a path is managed by dotfiles (points to DOTFILES_DIR)
is_ours() {
    local path="$1"
    [[ -e "$path" ]] || return 1
    local real_path
    real_path=$(realpath "$path" 2>/dev/null || echo "")
    [[ -n "$real_path" && "$real_path" == "$DOTFILES_DIR"* ]]
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
