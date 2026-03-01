#!/bin/bash
# lib/prompt.sh - User interaction with non-interactive support
#
# Usage: source "$(dirname "$0")/lib/prompt.sh"
#
# Non-interactive protocol:
#   DOTFILES_NONINTERACTIVE=1  - env var, auto-set by dot when stdin isn't a TTY
#   --force / -f flag to confirm() - always return yes
#
# See cmd/dot/ARCHITECTURE.md for full protocol details.

# Check if we're running interactively
_is_interactive() {
    # Explicit env var overrides everything
    [[ "${DOTFILES_NONINTERACTIVE:-}" == "1" ]] && return 1
    # Check if stdin is a TTY
    [[ -t 0 ]]
}

# confirm [-f] PROMPT [DEFAULT]
#
# Ask user for yes/no confirmation.
#   -f          Always return 0 (yes) — used with ${FORCE:+-f}
#   PROMPT      Question text (auto-appends [Y/n] or [y/N])
#   DEFAULT     "yes" or "no" (default: "yes")
#
# Returns: 0 for yes, 1 for no
confirm() {
    local force=false
    if [[ "${1:-}" == "-f" ]]; then
        force=true
        shift
    fi

    local prompt="${1:?confirm requires a prompt}"
    local default="${2:-yes}"

    # Force flag: always yes
    if [[ "$force" == true ]]; then
        return 0
    fi

    # Build prompt suffix
    local suffix
    if [[ "$default" == "yes" ]]; then
        suffix="[Y/n]"
    else
        suffix="[y/N]"
    fi

    # Non-interactive: use default
    if [[ "${DOTFILES_NONINTERACTIVE:-}" == "1" ]]; then
        [[ "$default" == "yes" ]] && return 0 || return 1
    fi

    # Not a TTY and no explicit env var: use default + warn
    if [[ ! -t 0 ]]; then
        echo "  (non-interactive: defaulting to $default for: $prompt)" >&2
        [[ "$default" == "yes" ]] && return 0 || return 1
    fi

    # Interactive: prompt user
    local answer
    read -p "$prompt $suffix " answer

    # Empty answer uses default
    if [[ -z "$answer" ]]; then
        [[ "$default" == "yes" ]] && return 0 || return 1
    fi

    # Check answer
    case "$answer" in
        [yY]|[yY][eE][sS]) return 0 ;;
        [nN]|[nN][oO]) return 1 ;;
        *) [[ "$default" == "yes" ]] && return 0 || return 1 ;;
    esac
}

# choose PROMPT OPT1 OPT2 ... [DEFAULT_INDEX]
#
# Present numbered choices. Returns the chosen option on stdout.
# DEFAULT_INDEX is 1-based (default: 1).
# In non-interactive mode, returns the default choice.
choose() {
    local prompt="${1:?choose requires a prompt}"
    shift

    local options=()
    local default_index=1

    # Collect options; last numeric-only arg is default_index
    while [[ $# -gt 0 ]]; do
        if [[ $# -eq 1 && "$1" =~ ^[0-9]+$ ]]; then
            default_index="$1"
        else
            options+=("$1")
        fi
        shift
    done

    if [[ ${#options[@]} -eq 0 ]]; then
        echo "choose: no options provided" >&2
        return 1
    fi

    # Non-interactive or not a TTY: return default
    if ! _is_interactive; then
        echo "${options[$((default_index - 1))]}"
        return 0
    fi

    # Interactive: show options
    echo ""
    echo "$prompt"
    local i=1
    for opt in "${options[@]}"; do
        echo "  $i) $opt"
        ((i++))
    done

    local choice
    read -p "Enter choice [1-${#options[@]}]: " choice

    # Validate choice
    if [[ "$choice" =~ ^[0-9]+$ ]] && [[ "$choice" -ge 1 ]] && [[ "$choice" -le ${#options[@]} ]]; then
        echo "${options[$((choice - 1))]}"
    else
        echo "${options[$((default_index - 1))]}"
    fi
}

# require_interactive MSG
#
# Exit with error if not interactive. Use to guard purely interactive features.
require_interactive() {
    local msg="${1:-This operation}"
    if ! _is_interactive; then
        echo -e "${RED:-}Error:${NC:-} $msg requires an interactive terminal." >&2
        echo "  Run from an interactive shell, or remove DOTFILES_NONINTERACTIVE." >&2
        exit 1
    fi
}
