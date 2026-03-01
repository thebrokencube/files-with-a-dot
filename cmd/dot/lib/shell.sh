#!/bin/bash
# lib/shell.sh - Shell config integration (source lines in .zshrc, .bashrc, etc.)
#
# Usage: source "$(dirname "$0")/lib/shell.sh"
# Requires: lib/colors.sh, lib/logging.sh, lib/backup.sh

# Managed shell config pairs: "target_config:source_dotfile"
# Each pair means: target_config should source the source_dotfile
# shellcheck disable=SC2034  # Array used by scripts that source this file
SHELL_CONFIG_PAIRS=(
    "zprofile:zprofile.dotfiles"
    "zshrc:zshrc.dotfiles"
    "bash_profile:bash_profile.dotfiles"
    "bashrc:bashrc.dotfiles"
)

# Check if a shell config file contains a source line for the given pattern
# Returns 0 if found, 1 if not
check_source_line() {
    local config="$1"
    local pattern="$2"
    [[ -f "$config" ]] && grep -qF "$pattern" "$config" 2>/dev/null
}

# Add a source line to a shell config file
# Creates the file if it doesn't exist, backs up before modifying
add_source_line() {
    local target="$1"
    local source_file="$2"
    local source_line="[ -f \"$source_file\" ] && source \"$source_file\""

    [[ ! -f "$target" ]] && echo "# Shell configuration" > "$target"

    if ! grep -qF "$source_file" "$target" 2>/dev/null; then
        backup_file "$target" "shell_config"
        echo "" >> "$target"
        echo "# Added by dotfiles" >> "$target"
        echo "$source_line" >> "$target"
        echo "  Added to $target"
    fi
    return 0
}

# Remove a source line (and its comment) from a shell config file
remove_source_line() {
    local target_file="$1"
    local pattern="$2"

    if [[ -f "$target_file" ]] && grep -qF "$pattern" "$target_file" 2>/dev/null; then
        grep -v "$pattern" "$target_file" | grep -v "# Added by dotfiles" > "$target_file.tmp"
        mv "$target_file.tmp" "$target_file"
        echo "  Cleaned: $target_file"
    fi
}

# Add source lines to all managed shell configs
integrate_shell_configs() {
    echo "Integrating shell configs..."

    local pair
    for pair in "${SHELL_CONFIG_PAIRS[@]}"; do
        local target="$HOME/.${pair%%:*}"
        local source_file="$HOME/.${pair##*:}"
        add_source_line "$target" "$source_file"
    done
    echo ""
}

# Remove source lines from all managed shell configs
remove_shell_configs() {
    local pair
    for pair in "${SHELL_CONFIG_PAIRS[@]}"; do
        local target="$HOME/.${pair%%:*}"
        local pattern=".${pair##*:}"
        remove_source_line "$target" "$pattern"
    done
}
