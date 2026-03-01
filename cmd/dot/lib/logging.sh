#!/bin/bash
# lib/logging.sh - Logging functions for consistent output
#
# Usage: source "$(dirname "$0")/lib/logging.sh"
# Requires: lib/colors.sh to be sourced first

# Logging functions
ok()      { echo -e "  ${GREEN}${SYM_OK}${NC} $1"; }
err()     { echo -e "  ${RED}${SYM_ERR}${NC} $1"; }
warn()    { echo -e "  ${YELLOW}${SYM_WARN}${NC} $1"; }
info()    { echo -e "  ${CYAN}${SYM_INFO}${NC} $1"; }
pending() { echo -e "  ${YELLOW}${SYM_PENDING}${NC} $1"; }

# Section headers
section() { echo -e "\n${CYAN}$1${NC}"; }

# Debug logging (only if DEBUG=1)
debug() {
    [[ "${DEBUG:-}" == "1" ]] && echo -e "  [DEBUG] $1"
}
