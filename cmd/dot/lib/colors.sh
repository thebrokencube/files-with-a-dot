#!/bin/bash
# lib/colors.sh - Color definitions and symbols for terminal output
#
# Usage: source "$(dirname "$0")/lib/colors.sh"

# shellcheck disable=SC2034  # Variables are used by scripts that source this file

# Colors
RED=$'\033[0;31m'
GREEN=$'\033[0;32m'
YELLOW=$'\033[0;33m'
CYAN=$'\033[0;36m'
BOLD=$'\033[1m'
NC=$'\033[0m'  # No Color

# Symbols
SYM_OK="✓"
SYM_ERR="✗"
SYM_WARN="⚠"
SYM_INFO="→"
SYM_PENDING="○"
SYM_BACKUP="⟳"
