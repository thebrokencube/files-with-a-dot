#!/bin/bash
# lib/config.sh - Machine configuration helpers
#
# Usage: source "$(dirname "$0")/lib/config.sh"
# Requires: DOTFILES_DIR to be set

# Read machine type from .machine file
# Returns: "aggressive", "conservative", or "" if not set
read_machine_type() {
    local machine_file="${DOTFILES_DIR}/.machine"
    if [[ -f "$machine_file" ]]; then
        cat "$machine_file"
    fi
}

# Read machine profile from .profile file
# Returns: "work", "personal", or "" if not set
read_profile() {
    local profile_file="${DOTFILES_DIR}/.profile"
    if [[ -f "$profile_file" ]]; then
        cat "$profile_file"
    fi
}

# Get dotfiles version from git
get_dotfiles_version() {
    if command -v git &>/dev/null && [[ -d "${DOTFILES_DIR}/.git" ]]; then
        git -C "$DOTFILES_DIR" describe --tags --always 2>/dev/null || echo "unknown"
    else
        echo "unknown"
    fi
}

# Initialize standard dotfiles variables
# Sets: MACHINE_TYPE, MACHINE_PROFILE, DOTFILES_VERSION
# shellcheck disable=SC2034  # Variables are used by scripts that source this file
init_dotfiles_vars() {
    MACHINE_TYPE="$(read_machine_type)"
    MACHINE_PROFILE="$(read_profile)"
    DOTFILES_VERSION="$(get_dotfiles_version)"
}
