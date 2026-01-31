#!/bin/bash
# health.sh - Diagnose dotfiles setup and system health
#
# Usage:
#   ./health.sh              # Run all checks
#   ./health.sh --fix        # Auto-fix issues where possible
#   ./health.sh --setup      # Interactive first-time setup guide
#   ./health.sh --check X    # Run specific check (tools, brew, local, nvim, setup)

set -e

AUTO_FIX=false
INTERACTIVE_SETUP=false
SPECIFIC_CHECK=""

while [[ $# -gt 0 ]]; do
    case $1 in
        --fix) AUTO_FIX=true; shift ;;
        --setup) INTERACTIVE_SETUP=true; shift ;;
        --check) SPECIFIC_CHECK="$2"; shift 2 ;;
        --help|-h)
            echo "Usage: ./health.sh [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --fix          Auto-fix issues where possible"
            echo "  --setup        Interactive first-time setup guide"
            echo "  --check NAME   Run specific check only"
            echo ""
            echo "Available checks: tools, brew, local, nvim, setup"
            exit 0
            ;;
        *) echo "Unknown option: $1"; exit 1 ;;
    esac
done

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DOTFILES_DIR="$SCRIPT_DIR"
MACHINE_FILE="$DOTFILES_DIR/.machine"

MACHINE_TYPE="unknown"
[[ -f "$MACHINE_FILE" ]] && MACHINE_TYPE=$(cat "$MACHINE_FILE")

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

ISSUES_FOUND=0

ok() { echo -e "  ${GREEN}✓${NC} $1"; }
warn() { echo -e "  ${YELLOW}⚠${NC} $1"; ((ISSUES_FOUND++)) || true; }
err() { echo -e "  ${RED}✗${NC} $1"; ((ISSUES_FOUND++)) || true; }
info() { echo -e "  ${CYAN}→${NC} $1"; }
pending() { echo -e "  ${YELLOW}○${NC} $1"; }

check_tools() {
    echo "Required tools:"

    local required=("brew" "git" "nvim" "starship" "mise")
    local missing=()

    for tool in "${required[@]}"; do
        if command -v "$tool" &>/dev/null; then
            local version=$("$tool" --version 2>/dev/null | head -1 || echo "installed")
            ok "$tool ($version)"
        else
            err "$tool not found"
            missing+=("$tool")
        fi
    done

    if [[ ${#missing[@]} -gt 0 && "$AUTO_FIX" == true ]]; then
        echo ""
        info "Installing missing tools..."
        brew install "${missing[@]}" 2>/dev/null || true
    fi
}

check_brew() {
    echo "Homebrew health:"

    if ! command -v brew &>/dev/null; then
        err "Homebrew not installed"
        return
    fi

    local deprecated_taps=("homebrew/bundle" "homebrew/core" "homebrew/cask")
    local found_deprecated=()

    for tap in "${deprecated_taps[@]}"; do
        brew tap 2>/dev/null | grep -q "^${tap}$" && found_deprecated+=("$tap")
    done

    if [[ ${#found_deprecated[@]} -gt 0 ]]; then
        warn "Deprecated taps: ${found_deprecated[*]}"
        if [[ "$AUTO_FIX" == true ]]; then
            for tap in "${found_deprecated[@]}"; do
                info "Untapping $tap..."
                brew untap "$tap" 2>/dev/null || true
            done
        fi
    else
        ok "No deprecated taps"
    fi

    if brew doctor 2>&1 | grep -q "Your system is ready to brew"; then
        ok "brew doctor: healthy"
    else
        warn "brew doctor has warnings (run 'brew doctor' for details)"
    fi

    local outdated=$(brew outdated --quiet 2>/dev/null | wc -l | tr -d ' ')
    if [[ "$outdated" -gt 0 ]]; then
        warn "$outdated outdated packages"
        [[ "$AUTO_FIX" == true ]] && info "Running brew upgrade..." && brew upgrade 2>/dev/null || true
    else
        ok "All packages up to date"
    fi
}

check_local() {
    echo "Local configuration:"

    # Check GitHub SSH access (needed for git operations)
    if ssh -T git@github.com 2>&1 | grep -q "successfully authenticated"; then
        ok "GitHub SSH access configured"
    else
        err "GitHub SSH not working (ssh -T git@github.com failed)"
        info "Set up SSH key: https://docs.github.com/en/authentication/connecting-to-github-with-ssh"
    fi

    if [[ -f "$HOME/.gitconfig.local" ]]; then
        local name=$(git config --file "$HOME/.gitconfig.local" user.name 2>/dev/null || echo "")
        local email=$(git config --file "$HOME/.gitconfig.local" user.email 2>/dev/null || echo "")
        if [[ -n "$name" && -n "$email" ]]; then
            ok "Git identity: $name <$email>"
        else
            warn "~/.gitconfig.local missing name or email"
        fi
    else
        warn "~/.gitconfig.local not found"
    fi

    if [[ -f "$HOME/.env.local" ]]; then
        ok "~/.env.local exists"
    else
        warn "~/.env.local not found"
    fi

    if [[ -f "$MACHINE_FILE" ]]; then
        ok "Machine type: $MACHINE_TYPE"
    else
        warn ".machine not set (run sync.sh)"
    fi
}

check_nvim() {
    echo "Neovim setup:"

    if ! command -v nvim &>/dev/null; then
        err "Neovim not installed"
        return
    fi

    if [[ -f "$HOME/.config/nvim/init.lua" ]]; then
        ok "Config: ~/.config/nvim/init.lua"
    else
        err "Config missing"
    fi

    local lazy_path="$HOME/.local/share/nvim/lazy/lazy.nvim"
    if [[ -d "$lazy_path" ]]; then
        local plugin_count=$(ls -1 "$HOME/.local/share/nvim/lazy" 2>/dev/null | wc -l | tr -d ' ')
        ok "lazy.nvim installed ($plugin_count plugins)"
    else
        warn "lazy.nvim not installed (open nvim to auto-install)"
    fi

}

# ============================================================================
# Setup Status - checks things that may need manual action
# ============================================================================

check_setup_status() {
    echo "Setup status:"

    local pending_items=()

    # Check nvim plugins
    if [[ -d "$HOME/.local/share/nvim/lazy" ]] && [[ $(ls "$HOME/.local/share/nvim/lazy" 2>/dev/null | wc -l) -gt 5 ]]; then
        ok "Neovim plugins installed"
    else
        pending "Neovim: Open nvim to trigger plugin install"
        pending_items+=("nvim")
    fi

    # Check Claude Code authentication
    if command -v claude &>/dev/null; then
        if claude auth status &>/dev/null; then
            ok "Claude Code authenticated"
        else
            pending "Claude Code: Run 'claude auth' to authenticate"
            pending_items+=("claude-auth")
        fi
    fi

    # Check iTerm2 profile (macOS only)
    if [[ "$(uname)" == "Darwin" ]]; then
        local iterm_profile="$HOME/Library/Application Support/iTerm2/DynamicProfiles/dotfiles-profile.json"
        if [[ -f "$iterm_profile" ]]; then
            ok "iTerm2 profile installed"
            # Can't detect if they're using it, so always remind
            info "iTerm2: Make sure 'Dotfiles Default' profile is selected"
        else
            pending "iTerm2: Profile not linked (run sync.sh)"
            pending_items+=("iterm")
        fi
    fi

    # Check Claude Code skills
    if [[ -d "$HOME/.claude/skills" ]] && [[ $(ls "$HOME/.claude/skills" 2>/dev/null | wc -l) -gt 0 ]]; then
        local skill_count=$(ls -1 "$HOME/.claude/skills" 2>/dev/null | wc -l | tr -d ' ')
        ok "Claude Code skills installed ($skill_count skills)"
    else
        pending "Claude Code: Skills not found (run sync.sh)"
        pending_items+=("claude")
    fi

    # Check shell is configured
    if [[ -f "$HOME/.zshrc" ]] && grep -q "zshrc.dotfiles" "$HOME/.zshrc" 2>/dev/null; then
        ok "Shell configured"
    else
        pending "Shell: Run 'exec \$SHELL -l' to reload"
        pending_items+=("shell")
    fi

    # Return pending items for interactive setup
    echo "${pending_items[*]}"
}

# ============================================================================
# Interactive Setup - walks through manual steps
# ============================================================================

run_interactive_setup() {
    echo "============================================"
    echo -e "  ${BOLD}Interactive Setup Guide${NC}"
    echo "============================================"
    echo ""
    echo "This will walk you through manual setup steps."
    echo "Press Enter to continue after each step, or 's' to skip."
    echo ""

    # Step 1: Shell
    echo -e "${BOLD}Step 1: Shell Configuration${NC}"
    echo "Your shell needs to be reloaded to pick up new config."
    echo ""
    echo "  Action: Run this command (or open a new terminal):"
    echo -e "  ${CYAN}exec \$SHELL -l${NC}"
    echo ""
    read -p "Press Enter when done (or 's' to skip): " response
    [[ "$response" != "s" ]] && ok "Shell reloaded"
    echo ""

    # Step 2: Neovim
    echo -e "${BOLD}Step 2: Neovim Plugins${NC}"
    if [[ -d "$HOME/.local/share/nvim/lazy" ]] && [[ $(ls "$HOME/.local/share/nvim/lazy" 2>/dev/null | wc -l) -gt 5 ]]; then
        ok "Already installed - skipping"
    else
        echo "Neovim plugins install automatically on first launch."
        echo ""
        echo "  Action: Open neovim and wait for plugins to install:"
        echo -e "  ${CYAN}nvim${NC}"
        echo ""
        echo "  You'll see Lazy.nvim installing plugins. Wait for it to finish."
        echo ""
        read -p "Press Enter when done (or 's' to skip): " response
        [[ "$response" != "s" ]] && ok "Neovim plugins installed"
    fi
    echo ""

    # Step 3: iTerm2 (macOS only)
    if [[ "$(uname)" == "Darwin" ]]; then
        echo -e "${BOLD}Step 3: iTerm2 Profile${NC}"
        echo "To get icons working in iTerm2, use the Dotfiles profile."
        echo ""
        echo "  Action:"
        echo "  1. Open iTerm2"
        echo "  2. Go to Profiles (⌘+O or Profiles menu)"
        echo "  3. Select 'Dotfiles Default'"
        echo "  4. Click 'Set as Default' (optional)"
        echo ""
        echo "  Test icons work: echo -e \"\\uf015 \\ue0a0 \\uf121\""
        echo ""
        read -p "Press Enter when done (or 's' to skip): " response
        [[ "$response" != "s" ]] && ok "iTerm2 configured"
        echo ""
    fi

    # Step 4: Claude Code Authentication
    if command -v claude &>/dev/null; then
        echo -e "${BOLD}Step 4: Claude Code Authentication${NC}"
        if claude auth status &>/dev/null; then
            ok "Already authenticated - skipping"
        else
            echo "Claude Code needs to be authenticated."
            echo ""
            echo "  Action: Run this command:"
            echo -e "  ${CYAN}claude auth${NC}"
            echo ""
            echo "  This will open your browser to authenticate."
            echo ""
            read -p "Press Enter when done (or 's' to skip): " response
            [[ "$response" != "s" ]] && ok "Claude Code authenticated"
        fi
        echo ""
    fi

    # Step 5: Git identity
    echo -e "${BOLD}Step 5: Git Identity${NC}"
    if [[ -f "$HOME/.gitconfig.local" ]]; then
        local name=$(git config --file "$HOME/.gitconfig.local" user.name 2>/dev/null || echo "")
        if [[ -n "$name" ]]; then
            ok "Already configured as: $name - skipping"
        fi
    else
        echo "Set your git name and email for commits."
        echo ""
        echo "  Action: The install script should have prompted for this."
        echo "  If not, edit ~/.gitconfig.local:"
        echo ""
        echo -e "  ${CYAN}[user]"
        echo "      name = Your Name"
        echo -e "      email = you@example.com${NC}"
        echo ""
        read -p "Press Enter when done (or 's' to skip): " response
        [[ "$response" != "s" ]] && ok "Git identity configured"
    fi
    echo ""

    # Done
    echo "============================================"
    echo -e "  ${GREEN}Setup Complete!${NC}"
    echo "============================================"
    echo ""
    echo "Run './health.sh' to verify everything is working."
    echo ""
    echo "Useful commands:"
    echo "  ./sync.sh        - Sync dotfiles (update if needed)"
    echo "  ./sync.sh --pull - Force pull and sync"
    echo "  ./health.sh      - Check system health"
    echo "  /dotfiles        - Claude Code skill for help"
    echo ""
}

# ============================================================================
# Main
# ============================================================================

# If --setup flag, run interactive setup and exit
if [[ "$INTERACTIVE_SETUP" == true ]]; then
    run_interactive_setup
    exit 0
fi

echo "============================================"
echo "  Dotfiles Health Check"
echo "============================================"
echo ""
echo "Repo: $DOTFILES_DIR"
echo "Machine: $MACHINE_TYPE"
[[ "$AUTO_FIX" == true ]] && echo -e "Auto-fix: ${GREEN}enabled${NC}"
echo ""

if [[ -n "$SPECIFIC_CHECK" ]]; then
    case "$SPECIFIC_CHECK" in
        tools) check_tools ;;
        brew) check_brew ;;
        local) check_local ;;
        nvim) check_nvim ;;
        setup) check_setup_status >/dev/null ;;
        *) echo "Unknown check: $SPECIFIC_CHECK"; exit 1 ;;
    esac
else
    check_tools
    echo ""
    check_brew
    echo ""
    check_local
    echo ""
    check_nvim
    echo ""
    check_setup_status >/dev/null
fi

echo ""
echo "============================================"
if [[ $ISSUES_FOUND -eq 0 ]]; then
    echo -e "  ${GREEN}All checks passed!${NC}"
else
    echo -e "  ${YELLOW}Found $ISSUES_FOUND issue(s)${NC}"
    echo ""
    echo "  Run with --fix to auto-fix where possible"
    echo "  Run with --setup for interactive first-time guide"
fi
echo "============================================"

exit $ISSUES_FOUND
