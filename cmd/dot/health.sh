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
SPECIFIC_CHECK=""

while [[ $# -gt 0 ]]; do
    case $1 in
        --fix) AUTO_FIX=true; shift ;;
        --check) SPECIFIC_CHECK="$2"; shift 2 ;;
        --help|-h)
            echo "Usage: ./health.sh [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --fix          Auto-fix issues where possible"
            echo "  --check NAME   Run specific check only"
            echo ""
            echo "Available checks: tools, brew, local, nvim, claude, signing, managed, setup"
            exit 0
            ;;
        *) echo "Unknown option: $1"; exit 1 ;;
    esac
done

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DOT_DIR="$SCRIPT_DIR"
DOTFILES_DIR="$(cd "$DOT_DIR/../.." && pwd)"

# Source libraries
# shellcheck source=lib/colors.sh
source "$DOT_DIR/lib/colors.sh"
# shellcheck source=lib/logging.sh
source "$DOT_DIR/lib/logging.sh"
# shellcheck source=lib/config.sh
source "$DOT_DIR/lib/config.sh"
# shellcheck source=lib/prompt.sh
source "$DOT_DIR/lib/prompt.sh"
# shellcheck source=lib/private.sh
source "$DOT_DIR/lib/private.sh"

MACHINE_TYPE="$(read_machine_type)"
MACHINE_TYPE="${MACHINE_TYPE:-unknown}"

ISSUES_FOUND=0

# Override warn/err to count issues
warn() { echo -e "  ${YELLOW}${SYM_WARN}${NC} $1"; ((ISSUES_FOUND++)) || true; }
err()  { echo -e "  ${RED}${SYM_ERR}${NC} $1"; ((ISSUES_FOUND++)) || true; }

check_tools() {
    echo "Required tools:"

    local required=("brew" "git" "nvim" "starship" "mise")
    local missing=()

    for tool in "${required[@]}"; do
        if command -v "$tool" &>/dev/null; then
            local version
            version=$("$tool" --version 2>/dev/null | head -1 || echo "installed")
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

check_dendrik_tools() {
    echo "Dendrik CLI tools:"

    local tools=(folio jf dendrik)
    for tool in "${tools[@]}"; do
        local dest="$HOME/.local/bin/$tool"
        local vfile="$DOTFILES_DIR/cmd/$tool/VERSION"
        local want=""
        [[ -f "$vfile" ]] && want="$(tr -d '[:space:]' < "$vfile")"

        if [[ ! -e "$dest" ]]; then
            if [[ -L "$dest" ]]; then
                err "$tool: dangling symlink — run 'dot sync' to install from releases"
            else
                err "$tool: not installed (expected $dest) — run 'dot sync'"
            fi
            continue
        fi
        # Post-migration these are real downloaded binaries, not repo symlinks.
        if [[ -L "$dest" ]]; then
            warn "$tool: still a symlink ($(readlink "$dest")) — pre-migration layout; run 'dot sync'"
            continue
        fi

        local have
        have="$("$dest" --version 2>/dev/null | awk '{print $NF}')"
        if [[ -z "$have" ]]; then
            err "$tool: installed but '--version' failed"
        elif [[ "$have" == "dev" ]]; then
            warn "$tool: version 'dev' (unstamped local build, not a release)"
        elif [[ -n "$want" && "$have" != "$want" ]]; then
            warn "$tool: $have installed, repo VERSION is $want — run 'dot sync' to update"
        else
            ok "$tool $have (matches release)"
        fi
    done
}

check_brew() {
    echo "Homebrew health:"

    if ! command -v brew &>/dev/null; then
        err "Homebrew not installed"
        return
    fi

    # Official taps now built-in or merged into homebrew/cask. The cask-*
    # taps were merged upstream, so untapping is safe even with installed
    # casks (they resolve from the main cask) — hence --force below.
    local deprecated_taps=(
        "homebrew/bundle"
        "homebrew/core"
        "homebrew/cask"
        "homebrew/cask-fonts"
        "homebrew/cask-versions"
        "homebrew/cask-drivers"
    )
    local found_deprecated=()

    for tap in "${deprecated_taps[@]}"; do
        brew tap 2>/dev/null | grep -q "^${tap}$" && found_deprecated+=("$tap")
    done

    if [[ ${#found_deprecated[@]} -gt 0 ]]; then
        warn "Deprecated taps: ${found_deprecated[*]}"
        if [[ "$AUTO_FIX" == true ]]; then
            for tap in "${found_deprecated[@]}"; do
                info "Untapping $tap..."
                brew untap --force "$tap" 2>/dev/null || true
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

    local outdated
    outdated=$(brew outdated --quiet 2>/dev/null | wc -l | tr -d ' ')
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
        local name email
        name=$(git config --file "$HOME/.gitconfig.local" user.name 2>/dev/null || echo "")
        email=$(git config --file "$HOME/.gitconfig.local" user.email 2>/dev/null || echo "")
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

    if [[ -n "$MACHINE_TYPE" && "$MACHINE_TYPE" != "unknown" ]]; then
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
        local plugin_count
        plugin_count=$(ls -1 "$HOME/.local/share/nvim/lazy" 2>/dev/null | wc -l | tr -d ' ')
        ok "lazy.nvim installed ($plugin_count plugins)"
    else
        warn "lazy.nvim not installed (open nvim to auto-install)"
    fi

}

check_managed() {
    echo "Managed files:"

    if ! command -v jq &>/dev/null; then
        warn "jq not found — cannot check managed file drift"
        return
    fi

    local map_file="$DOTFILES_DIR/managed_map.txt"
    if [[ ! -f "$map_file" ]]; then
        ok "No managed_map.txt"
        return
    fi

    if check_managed_drift "$map_file"; then
        ok "No drift detected"
    fi
}

check_claude() {
    echo "Claude Code:"

    if ! command -v claude &>/dev/null; then
        err "Claude Code not found"
        return
    fi

    local claude_path
    claude_path=$(command -v claude)
    local version
    version=$(claude --version 2>/dev/null || echo "unknown")

    # Check for legacy npm installation via mise
    if [[ -n "$(mise ls "npm:@anthropic-ai/claude-code" 2>/dev/null)" ]]; then
        warn "Legacy npm install found in mise (run 'dot sync' to migrate)"
    fi

    # Check if native binary
    if [[ -x "$HOME/.local/bin/claude" ]] && file "$HOME/.local/bin/claude" 2>/dev/null | grep -q "Mach-O\|ELF"; then
        ok "Native binary ($version) at $claude_path"
    elif echo "$claude_path" | grep -q "node"; then
        warn "Using npm-based claude ($version) — run 'dot sync' to migrate to native"
    else
        ok "Claude Code ($version) at $claude_path"
    fi

    # Statusline (starship-claude) — silently no-ops if starship is missing.
    if [[ -x "$HOME/.local/bin/starship-claude" ]]; then
        if command -v starship &>/dev/null; then
            ok "Statusline: starship-claude installed"
        else
            warn "Statusline: starship-claude present but starship is missing — statusline will not render"
        fi
    else
        warn "Statusline: starship-claude not installed — run 'dot sync'"
    fi
}

check_signing() {
    echo "Commit signing:"

    # Detect SSH-based commit signing (jj preferred, then git)
    local signing_on=false
    local sign_key=""

    if command -v jj &>/dev/null; then
        local jj_backend jj_behavior
        jj_backend=$(jj config get signing.backend 2>/dev/null || echo "")
        jj_behavior=$(jj config get signing.behavior 2>/dev/null || echo "")
        if [[ "$jj_backend" == "ssh" && -n "$jj_behavior" && "$jj_behavior" != "drop" ]]; then
            signing_on=true
            sign_key=$(jj config get signing.key 2>/dev/null || echo "")
        fi
    fi

    if [[ "$signing_on" == false ]] && command -v git &>/dev/null; then
        if [[ "$(git config --global commit.gpgsign 2>/dev/null)" == "true" \
            && "$(git config --global gpg.format 2>/dev/null)" == "ssh" ]]; then
            signing_on=true
            sign_key=$(git config --global user.signingkey 2>/dev/null || echo "")
        fi
    fi

    if [[ "$signing_on" == false ]]; then
        ok "SSH commit signing not enabled (nothing to check)"
        return
    fi

    # Signing is on — a key must be loaded in the agent or every commit fails to sign
    if ssh-add -l &>/dev/null; then
        ok "SSH signing enabled and ssh-agent has keys loaded"
        return
    fi

    warn "SSH signing is ON but ssh-agent has no keys — commits will fail to sign (often after a reboot)"

    # Resolve a private key to load (signing.key usually points at the .pub)
    local priv_key="${sign_key%.pub}"
    [[ -z "$priv_key" || ! -f "$priv_key" ]] && priv_key="$HOME/.ssh/id_ed25519"

    if [[ "$AUTO_FIX" == true && "$(uname)" == "Darwin" ]]; then
        info "Loading key into agent + keychain (you'll be prompted for the passphrase once)..."
        ssh-add --apple-use-keychain "$priv_key" || warn "ssh-add failed — verify the passphrase for $priv_key"
    elif [[ "$(uname)" == "Darwin" ]]; then
        info "Fix: ssh-add --apple-use-keychain $priv_key  (seeds keychain so it auto-loads after reboot)"
    else
        info "Fix: ssh-add $priv_key"
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
        local skill_count
        skill_count=$(ls -1 "$HOME/.claude/skills" 2>/dev/null | wc -l | tr -d ' ')
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
# Main
# ============================================================================

VERSION="$(get_dotfiles_version)"
echo "============================================"
echo "  Dotfiles Health Check ($VERSION)"
echo "============================================"
echo ""
echo "Repo: $DOTFILES_DIR"
echo "Machine type: $MACHINE_TYPE"
[[ "$AUTO_FIX" == true ]] && echo -e "Auto-fix: ${GREEN}enabled${NC}"
echo ""

if [[ -n "$SPECIFIC_CHECK" ]]; then
    case "$SPECIFIC_CHECK" in
        tools) check_tools ;;
        dendrik-tools) check_dendrik_tools ;;
        brew) check_brew ;;
        local) check_local ;;
        nvim) check_nvim ;;
        claude) check_claude ;;
        signing) check_signing ;;
        managed) check_managed ;;
        setup) check_setup_status >/dev/null ;;
        *) echo "Unknown check: $SPECIFIC_CHECK"; exit 1 ;;
    esac
else
    check_tools
    echo ""
    check_dendrik_tools
    echo ""
    check_brew
    echo ""
    check_local
    echo ""
    check_nvim
    echo ""
    check_claude
    echo ""
    check_signing
    echo ""
    check_managed
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
fi
echo "============================================"

exit $ISSUES_FOUND
