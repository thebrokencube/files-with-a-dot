#!/bin/bash
# ~/ in user-facing display strings is literal shorthand, not a path to expand;
# real path operations in this file use $HOME.
# shellcheck disable=SC2088
#
# sync.sh - Synchronize dotfiles to current machine state
#
# This is the main command for dotfiles management. It intelligently detects
# whether this is a first-time setup or an update, and applies the state from
# symlink_map.txt to your system.
#
# Usage:
#   ./sync.sh                     # Auto-detect mode and sync
#   ./sync.sh --pull              # Force git pull before sync
#   ./sync.sh --dry-run           # Preview changes without applying
#   ./sync.sh --links-only        # Only re-create symlinks
#
# First-time setup (non-interactive):
#   ./sync.sh --machine aggressive --git-name "Name" --git-email "email"

set -e

# Error handler - show detailed error info (only on actual errors).
# exitcode is assigned at the head of this same trap command (SC2154 false positive).
# shellcheck disable=SC2154
trap 'exitcode=$?; if [[ $exitcode -ne 0 ]]; then echo ""; echo "ERROR: Command failed at line $LINENO with exit code $exitcode"; echo "Function: ${FUNCNAME[0]:-main}"; exit $exitcode; fi' ERR

# ============================================================================
# Arguments & Configuration
# ============================================================================

SKIP_BREW=false
SKIP_PULL=false
DRY_RUN=false
NO_BACKUP=false
FORCE=false
FORCE_PULL=false

MACHINE_ARG=""
GIT_NAME_ARG=""
GIT_EMAIL_ARG=""

while [[ $# -gt 0 ]]; do
    case $1 in
        --dry-run) DRY_RUN=true; shift ;;
        --skip-brew) SKIP_BREW=true; shift ;;
        --skip-pull) SKIP_PULL=true; shift ;;
        --pull) FORCE_PULL=true; shift ;;
        --links-only) SKIP_BREW=true; SKIP_PULL=true; shift ;;
        --no-backup) NO_BACKUP=true; shift ;;
        --force) FORCE=true; shift ;;
        --machine) MACHINE_ARG="$2"; shift 2 ;;
        --git-name) GIT_NAME_ARG="$2"; shift 2 ;;
        --git-email) GIT_EMAIL_ARG="$2"; shift 2 ;;
        --help|-h)
            echo "Usage: ./sync.sh [OPTIONS]"
            echo ""
            echo "Reconcile the machine to the LOCAL dotfiles repo (symlinks, managed"
            echo "files, CLI tools). Local-only by default — does NOT contact origin."
            echo "Use --pull (or 'dot pull') to fetch from origin first."
            echo ""
            echo "Options:"
            echo "  --dry-run            Preview changes without applying"
            echo "  --pull               Fetch + rebase from origin before applying"
            echo "  --skip-brew          Skip Homebrew package operations"
            echo "  --skip-pull          Hard override: never pull (even with --pull)"
            echo "  --links-only         Only re-create symlinks (implies --skip-brew --skip-pull)"
            echo "  --no-backup          Skip backing up existing files"
            echo "  --force              Auto-confirm all prompts"
            echo ""
            echo "First-time setup (non-interactive):"
            echo "  --machine TYPE       Set machine type (aggressive|conservative)"
            echo "  --git-name NAME      Set git user name"
            echo "  --git-email EMAIL    Set git user email"
            echo ""
            echo "Private overlay:"
            echo "  Place private configs in ~/.dotfiles.private/"
            echo "  Run: dot private init"
            echo ""
            echo "Debug:"
            echo "  DEBUG=1 ./sync.sh    Run with debug output"
            exit 0
            ;;
        *) echo "Unknown option: $1"; exit 1 ;;
    esac
done

MACHINE_TYPE="${MACHINE_ARG:-${DOTFILES_MACHINE:-}}"
GIT_NAME="${GIT_NAME_ARG:-${DOTFILES_GIT_NAME:-}}"
GIT_EMAIL="${GIT_EMAIL_ARG:-${DOTFILES_GIT_EMAIL:-}}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DOT_DIR="$SCRIPT_DIR"
DOTFILES_DIR="$(cd "$DOT_DIR/../.." && pwd)"
DOTFILES_LINK="$HOME/.dotfiles"
BACKUP_DIR="$DOTFILES_DIR/.backup"
BACKUP_MANIFEST="$BACKUP_DIR/manifest"
SYMLINK_MAP="$DOTFILES_DIR/symlink_map.txt"
MACHINE_FILE="$DOTFILES_DIR/.machine"
PRIVATE_DIR="$HOME/.dotfiles.private"

# Source libraries
# shellcheck source=lib/colors.sh
source "$DOT_DIR/lib/colors.sh"
# shellcheck source=lib/logging.sh
source "$DOT_DIR/lib/logging.sh"
# shellcheck source=lib/config.sh
source "$DOT_DIR/lib/config.sh"
# shellcheck source=lib/prompt.sh
source "$DOT_DIR/lib/prompt.sh"
# shellcheck source=lib/paths.sh
source "$DOT_DIR/lib/paths.sh"
# shellcheck source=lib/backup.sh
source "$DOT_DIR/lib/backup.sh"
# shellcheck source=lib/symlinks.sh
source "$DOT_DIR/lib/symlinks.sh"
# shellcheck source=lib/tools.sh
source "$DOT_DIR/lib/tools.sh"
# shellcheck source=lib/shell.sh
source "$DOT_DIR/lib/shell.sh"
# shellcheck source=lib/private.sh
source "$DOT_DIR/lib/private.sh"
# shellcheck source=lib/brew.sh
source "$DOT_DIR/lib/brew.sh"
# shellcheck source=lib/git.sh
source "$DOT_DIR/lib/git.sh"

# State arrays
ACTIONS=()
DONE_INFRA=()      # git status, ~/.dotfiles link
DONE_SHELL=()      # shell config integrations
DONE_SYMLINKS=()   # public symlinks (from symlink_map.txt)
DONE_PRIVATE=()    # private overlay symlinks
FRICTIONS=()
WILL_BACKUP=()


# ============================================================================
# State Detection
# ============================================================================

detect_first_time() {
    [[ ! -f "$MACHINE_FILE" ]]
}

# ============================================================================
# Analysis & Planning
# ============================================================================

analyze_state() {
    local is_first_time="$1"

    cd "$DOTFILES_DIR" || {
        echo "Error: Cannot access dotfiles directory: $DOTFILES_DIR"
        exit 1
    }

    # Check git/jj status. Only when pulling (--pull) — plain sync is local-only and
    # must not contact origin (the fetch below would).
    [[ "${DEBUG:-}" == "1" ]] && echo "  Checking git status..."
    if [[ "$FORCE_PULL" == true ]]; then
        if [[ -d "$DOTFILES_DIR/.jj" ]]; then
            # jj repo: check if working copy (@) has uncommitted changes
            local jj_wc
            jj_wc=$(jj log -r "@" --no-graph -T 'if(empty, "empty", "changed")' -R "$DOTFILES_DIR" 2>/dev/null || echo "unknown")
            if [[ "$jj_wc" == "changed" ]]; then
                FRICTIONS+=("Uncommitted changes in repo - commit or abandon before pulling")
            else
                DONE_INFRA+=("Git repo up to date")
            fi
        elif ! git rev-parse --is-inside-work-tree &>/dev/null; then
            FRICTIONS+=("Not a git repository")
        elif ! git diff --quiet 2>/dev/null || ! git diff --cached --quiet 2>/dev/null; then
            FRICTIONS+=("Uncommitted changes in repo - commit or stash before pulling")
        else
            git fetch --quiet 2>/dev/null || true
            LOCAL=$(git rev-parse HEAD 2>/dev/null || echo "")
            REMOTE=$(git rev-parse '@{u}' 2>/dev/null || echo "")
            if [[ -n "$LOCAL" && -n "$REMOTE" && "$LOCAL" != "$REMOTE" ]]; then
                BEHIND=$(git rev-list --count 'HEAD..@{u}' 2>/dev/null || echo "0")
                if [[ "$BEHIND" -gt 0 ]]; then
                    ACTIONS+=("Pull $BEHIND commit(s) from remote")
                else
                    DONE_INFRA+=("Git repo up to date")
                fi
            else
                DONE_INFRA+=("Git repo up to date")
            fi
        fi
    fi

    # Check ~/.dotfiles
    [[ "${DEBUG:-}" == "1" ]] && echo "  Checking ~/.dotfiles..."
    if [[ "$DOTFILES_DIR" == "$DOTFILES_LINK" ]]; then
        DONE_INFRA+=("~/.dotfiles (repo location)")
    elif [[ -L "$DOTFILES_LINK" ]]; then
        if is_ours "$DOTFILES_LINK"; then
            DONE_INFRA+=("~/.dotfiles symlink")
        else
            FRICTIONS+=("~/.dotfiles is a symlink to $(readlink "$DOTFILES_LINK"), not this repo")
        fi
    elif [[ -d "$DOTFILES_LINK" ]]; then
        FRICTIONS+=("~/.dotfiles is a directory (not this repo or symlink to it) - remove or rename it")
    elif [[ -e "$DOTFILES_LINK" ]]; then
        FRICTIONS+=("~/.dotfiles exists but is not a symlink or directory")
    else
        ACTIONS+=("Create ~/.dotfiles symlink")
    fi

    # Check required tools
    command -v git &>/dev/null || FRICTIONS+=("Required tool not found: git")
    command -v brew &>/dev/null || FRICTIONS+=("Homebrew not installed - run bootstrap.sh first")

    # Check symlink map exists
    [[ ! -f "$SYMLINK_MAP" ]] && FRICTIONS+=("symlink_map.txt not found")

    # Check shell configs
    [[ "${DEBUG:-}" == "1" ]] && echo "  Checking shell configs..."
    local pair
    for pair in "${SHELL_CONFIG_PAIRS[@]}"; do
        check_shell_config "$HOME/.${pair%%:*}" ".${pair##*:}"
    done

    # Check all symlink map entries
    [[ "${DEBUG:-}" == "1" ]] && echo "  Checking symlink map entries..."
    if [[ -f "$SYMLINK_MAP" ]]; then
        while IFS= read -r line || [[ -n "$line" ]]; do
            [[ -z "$line" || "$line" =~ ^[[:space:]]*# ]] && continue
            local source dest
            source=$(get_source "$line")
            dest=$(get_dest "$line")
            check_symlink "$source" "$dest" || true
        done < "$SYMLINK_MAP"
    fi

    # Check private symlinks
    if has_private_overlay; then
        [[ "${DEBUG:-}" == "1" ]] && echo "  Checking private symlinks..."
        local private_map
        private_map="$(get_private_symlink_map)"
        if [[ -f "$private_map" ]]; then
            while IFS= read -r line || [[ -n "$line" ]]; do
                [[ -z "$line" || "$line" =~ ^[[:space:]]*# ]] && continue
                local source dest
                source=$(get_source "$line")
                dest=$(get_dest "$line")
                check_private_symlink "$source" "$dest" "$PRIVATE_DIR" || true
            done < "$private_map"
        fi

        # Private skills
        if [[ -d "$PRIVATE_DIR/skills" ]]; then
            for skill in "$PRIVATE_DIR/skills"/*/; do
                [[ -d "$skill" ]] || continue
                local skill_name
                skill_name=$(basename "$skill")
                [[ "$skill_name" == ".gitkeep" ]] && continue
                local dest="$HOME/.claude/skills/$skill_name"
                check_private_symlink "skills/$skill_name" "$dest" "$PRIVATE_DIR" || true
            done
        fi
    fi

    # Check brew
    if [[ "$SKIP_BREW" == false ]]; then
        if ! command -v brew &>/dev/null; then
            FRICTIONS+=("Homebrew not installed")
        else
            ACTIONS+=("Install/update Homebrew packages")
        fi
    fi

    # Check backups
    [[ -f "$BACKUP_MANIFEST" ]] && ACTIONS+=("Sync backups (update if files changed)")

    # Ensure function always returns success
    return 0
}

check_shell_config() {
    local config="$1"
    local source_file="$2"
    local display="~/${config#"$HOME"/}"

    if [[ ! -e "$config" ]]; then
        ACTIONS+=("Create $display with source line")
    elif is_foreign_symlink "$config"; then
        FRICTIONS+=("$display is a symlink to $(readlink "$config") - can't modify")
    elif grep -qF "$source_file" "$config" 2>/dev/null; then
        DONE_SHELL+=("$display sources $source_file")
    else
        ACTIONS+=("Append source line to $display")
        [[ "$NO_BACKUP" != true ]] && WILL_BACKUP+=("$config (shell config)")
    fi
    return 0
}

# ============================================================================
# Reporting
# ============================================================================

report_state() {
    local total_done=$(( ${#DONE_INFRA[@]} + ${#DONE_SHELL[@]} + ${#DONE_SYMLINKS[@]} + ${#DONE_PRIVATE[@]} ))

    echo "--- Already in place ---"
    if [[ $total_done -eq 0 ]]; then
        echo "  (nothing)"
    else
        for item in "${DONE_INFRA[@]}"; do echo -e "  ${GREEN}✓${NC} $item"; done
        for item in "${DONE_SHELL[@]}"; do echo -e "  ${GREEN}✓${NC} $item"; done
        for item in "${DONE_SYMLINKS[@]}"; do echo -e "  ${GREEN}✓${NC} $item"; done
        for item in "${DONE_PRIVATE[@]}"; do echo -e "  ${GREEN}✓${NC} $item (private)"; done
    fi
    echo ""

    echo "--- Will do ---"
    if [[ ${#ACTIONS[@]} -eq 0 ]]; then
        echo "  (nothing)"
    else
        for item in "${ACTIONS[@]}"; do echo -e "  ${CYAN}→${NC} $item"; done
    fi
    echo ""

    echo "--- Will backup ---"
    if [[ ${#WILL_BACKUP[@]} -eq 0 ]]; then
        echo "  (nothing)"
    else
        for item in "${WILL_BACKUP[@]}"; do echo -e "  ${YELLOW}⟳${NC} $item"; done
    fi
    echo ""

    echo "--- Frictions (cannot proceed) ---"
    if [[ ${#FRICTIONS[@]} -eq 0 ]]; then
        echo -e "  ${GREEN}None!${NC}"
    else
        for item in "${FRICTIONS[@]}"; do echo -e "  ${RED}✗${NC} $item"; done
    fi
    echo ""
}

# ============================================================================
# Git Operations
# ============================================================================

# Pulling from origin is OPT-IN: only `dot pull` (--pull) contacts the remote.
# Plain `dot sync` reconciles the machine to the LOCAL repo and never touches
# origin — so you can sync without syncing-with-origin (e.g. apply local edits, or
# a deliberately-behind checkout, without fetching). --skip-pull is a hard override.
handle_git_pull() {
    local is_first_time="$1"

    [[ "$is_first_time" == true ]] && return 0
    [[ "$SKIP_PULL" == true ]] && return 0
    [[ "$FORCE_PULL" != true ]] && return 0

    # jj-managed repos: HEAD is detached, git pull doesn't work
    if [[ -d "$DOTFILES_DIR/.jj" ]]; then
        echo "Pulling latest dotfiles..."
        jj git fetch -R "$DOTFILES_DIR"
        jj bookmark set main -r "main@origin" -R "$DOTFILES_DIR" 2>/dev/null || true
        # fetch + bookmark move the bookmark but leave the working copy (@) on the OLD
        # main, so on-disk files would stay stale. Rebase @ onto the new main: this
        # fast-forwards an empty @, carries local edits forward, and no-ops when current.
        jj rebase -d main -R "$DOTFILES_DIR" || true
        echo ""
        return 0
    fi

    echo "Pulling latest dotfiles..."
    git pull --rebase
    echo ""
}

# ============================================================================
# Homebrew Management
# ============================================================================

install_brew_packages() {
    local is_first_time="$1"

    [[ "$SKIP_BREW" == true ]] && return 0

    echo "Installing Homebrew packages..."
    brew bundle --file="$DOTFILES_DIR/Brewfile.shared" --verbose || echo "  (some packages may have failed)"

    if [[ "$MACHINE_TYPE" == "aggressive" && -f "$DOTFILES_DIR/configs/aggressive/Brewfile" ]]; then
        echo ""
        echo "Installing aggressive-mode packages..."
        brew bundle --file="$DOTFILES_DIR/configs/aggressive/Brewfile" --verbose || echo "  (some packages may have failed)"
    fi

    # Private overlay Brewfile
    if has_private_overlay; then
        local private_brewfile="$PRIVATE_DIR/Brewfile"
        if [[ -f "$private_brewfile" ]]; then
            echo ""
            echo "Installing private packages..."
            brew bundle --file="$private_brewfile" --verbose || echo "  (some packages may have failed)"
        fi
    fi

    # Update mode: upgrade existing packages
    if [[ "$is_first_time" == false ]]; then
        echo ""
        echo "Upgrading installed packages..."
        brew upgrade || true
        brew upgrade --cask || true
    fi

    if command -v mise &>/dev/null; then
        echo ""
        echo "Updating mise tools..."
        mise upgrade || echo "  (no updates available)"
    fi

    if [[ "$MACHINE_TYPE" == "aggressive" ]]; then
        echo ""
        echo "Running brew cleanup..."
        brew cleanup || echo "  (no cleanup needed)"
    fi
    echo ""
}

# ============================================================================
# Local Config Setup
# ============================================================================

setup_local_configs() {
    local is_first_time="$1"

    # Migrate: move local/ files to $HOME (local/ removed in v0.7.0)
    if [[ -d "$DOTFILES_DIR/local" ]]; then
        local local_dir="$DOTFILES_DIR/local"
        local migrated=false
        if [[ -f "$local_dir/gitconfig.local" && ! -f "$HOME/.gitconfig.local" ]]; then
            mv "$local_dir/gitconfig.local" "$HOME/.gitconfig.local"
            echo "  Migrated local/gitconfig.local to ~/.gitconfig.local"
            migrated=true
        fi
        if [[ -f "$local_dir/env.local" && ! -f "$HOME/.env.local" ]]; then
            mv "$local_dir/env.local" "$HOME/.env.local"
            echo "  Migrated local/env.local to ~/.env.local"
            migrated=true
        fi
        if [[ -f "$local_dir/shell.local" && ! -f "$HOME/.shell.local" ]]; then
            mv "$local_dir/shell.local" "$HOME/.shell.local"
            echo "  Migrated local/shell.local to ~/.shell.local"
            migrated=true
        fi
        # Remove local/ dir if only shell.managed (or empty) remains
        rm -f "$local_dir/shell.managed" 2>/dev/null  # now lives in aggressive/
        rmdir "$local_dir" 2>/dev/null || true
        [[ "$migrated" == true ]] && echo ""
    fi

    # Git config - only prompt on first-time if no gitconfig.local exists yet
    if [[ ! -f "$HOME/.gitconfig.local" ]]; then
        if [[ -n "$GIT_NAME" && -n "$GIT_EMAIL" ]]; then
            cat > "$HOME/.gitconfig.local" << EOF
[user]
    name = $GIT_NAME
    email = $GIT_EMAIL
EOF
            echo "  Created ~/.gitconfig.local"
        elif [[ "$is_first_time" == true ]] && _is_interactive; then
            echo ""
            read -rp "Git name: " GIT_NAME
            read -rp "Git email: " GIT_EMAIL
            if [[ -n "$GIT_NAME" && -n "$GIT_EMAIL" ]]; then
                cat > "$HOME/.gitconfig.local" << EOF
[user]
    name = $GIT_NAME
    email = $GIT_EMAIL
EOF
                echo "  Created ~/.gitconfig.local"
            fi
        fi
    fi

    echo ""
}

# ============================================================================
# First-Time Setup
# ============================================================================

prompt_first_time_config() {
    if [[ -n "$MACHINE_TYPE" ]]; then
        if [[ "$MACHINE_TYPE" != "aggressive" && "$MACHINE_TYPE" != "conservative" ]]; then
            echo "Error: Invalid machine type '$MACHINE_TYPE'. Must be 'aggressive' or 'conservative'."
            exit 1
        fi
    else
        MACHINE_TYPE=$(choose "Machine type:" "aggressive" "conservative")
    fi
    echo "$MACHINE_TYPE" > "$MACHINE_FILE"
}

# ============================================================================
# Main Execution
# ============================================================================

VERSION="$(get_dotfiles_version)"
echo "============================================"
echo "  Dotfiles Sync ($VERSION)"
echo "============================================"
echo ""
echo "Repo: $DOTFILES_DIR"
[[ "$DRY_RUN" == true ]] && echo -e "${CYAN}Mode: DRY RUN${NC}"
echo ""

# Detect mode
IS_FIRST_TIME=$(detect_first_time && echo true || echo false)

if [[ "$IS_FIRST_TIME" == true ]]; then
    echo "Mode: First-time setup"
    [[ -f "$MACHINE_FILE" ]] && MACHINE_TYPE=$(cat "$MACHINE_FILE")
else
    echo "Mode: Update"
    MACHINE_TYPE=$(cat "$MACHINE_FILE")
fi
echo "Machine type: ${MACHINE_TYPE:-unknown}"
if has_private_overlay; then
    echo "Private: $PRIVATE_DIR"
fi
echo ""

echo "Detecting current state..."
echo ""

# Analyze state
analyze_state "$IS_FIRST_TIME"

# Report
report_state

# Gates
if [[ ${#FRICTIONS[@]} -gt 0 ]]; then
    echo -e "${RED}Cannot proceed due to frictions above.${NC}"
    exit 1
fi

if [[ ${#ACTIONS[@]} -eq 0 ]]; then
    echo -e "${GREEN}Nothing to do!${NC}"
    exit 0
fi

if [[ "$DRY_RUN" == true ]]; then
    echo "Dry run complete. Run without --dry-run to apply changes."
    exit 0
fi

if ! confirm ${FORCE:+-f} "Proceed with sync?"; then
    echo "Aborted."
    exit 2
fi

# Execute sync
echo ""
echo "============================================"
echo "  Executing Sync"
echo "============================================"
echo ""

# First-time setup
if [[ "$IS_FIRST_TIME" == true ]]; then
    prompt_first_time_config
    echo ""
fi

# Initialize backup system
init_backup

# Git pull
handle_git_pull "$IS_FIRST_TIME"

# Sync backups
if [[ -f "$BACKUP_MANIFEST" ]]; then
    echo "Syncing backups..."
    sync_backups
    echo ""
fi

# Create ~/.dotfiles symlink if needed
create_dotfiles_symlink
# Use symlink path for subsequent symlinks (more portable)
[[ "$DOTFILES_DIR" != "$HOME/.dotfiles" && -L "$HOME/.dotfiles" ]] && DOTFILES_DIR="$HOME/.dotfiles"

# Migrate: remove legacy profile system (profiles removed in v0.7.0)
[[ -f "$DOTFILES_DIR/.profile" ]] && rm -f "$DOTFILES_DIR/.profile"
migrate_private_overlay

# Migrate from ~/.claude directory symlink to granular linking
if [[ -L "$HOME/.claude" && -d "$HOME/.claude" ]]; then
    target=$(readlink "$HOME/.claude")
    if [[ "$target" == *"shared/claude/.claude"* || "$target" == *"configs/base/claude/.claude"* || "$target" == *".dotfiles"*"/.claude"* ]]; then
        echo "Migrating ~/.claude from directory link to granular links..."
        rm "$HOME/.claude"  # Remove symlink (not the directory it points to)
        mkdir -p "$HOME/.claude/skills"
        echo ""
    fi
fi

# Apply symlinks
echo "Creating symlinks..."
apply_symlinks "$SYMLINK_MAP"

# Install dendrik-built CLI tools from GitHub Releases (replaces the old committed-binary symlinks)
echo ""
echo "Installing CLI tools..."
install_tools

if has_private_overlay; then
    echo ""
    echo "Applying private overlay..."
    apply_private_symlinks "$(get_private_symlink_map)" "$PRIVATE_DIR"

    # Private skills
    if [[ -d "$PRIVATE_DIR/skills" ]]; then
        for skill in "$PRIVATE_DIR/skills"/*/; do
            [[ -d "$skill" ]] || continue
            skill_name=$(basename "$skill")
            [[ "$skill_name" == ".gitkeep" ]] && continue
            dest="$HOME/.claude/skills/$skill_name"
            mkdir -p "$(dirname "$dest")"
            create_private_symlink "skills/$skill_name" "$dest" "$PRIVATE_DIR"
        done
    fi
fi
echo ""

# Integrate shell configs
integrate_shell_configs

# Install/update brew packages
install_brew_packages "$IS_FIRST_TIME"

# Setup local configs
setup_local_configs "$IS_FIRST_TIME"

# Offer to init private overlay on first-time setup
if [[ "$IS_FIRST_TIME" == true ]] && ! has_private_overlay; then
    echo ""
    echo "The private overlay stores machine-specific configs (git identity,"
    echo "API keys, shell customizations) in a separate directory that can"
    echo "optionally be backed up to a private git remote."
    echo ""
    if confirm ${FORCE:+-f} "Initialize private overlay at $PRIVATE_DIR?"; then
        init_private_overlay
        echo ""
        echo "Applying private symlinks..."
        private_sync
    fi
elif has_private_overlay; then
    # Existing overlay: re-apply symlinks (may have new entries)
    echo "Applying private symlinks..."
    private_sync
fi

# Auto-resolve disk-ahead drift for plugin manifests (Claude self-updates these)
reconcile_plugin_drift "$DOTFILES_DIR/managed_map.txt"

# Check for managed file drift. Drifted files are SKIPPED by apply_managed_files (so we
# never clobber on-disk changes), but the rest of the sync continues — one drifted file
# no longer halts everything. Backfill drift to sources, or 'dot sync --force' to overwrite.
DRIFTED_FILES=()
if [[ "$FORCE" != true ]]; then
    check_managed_drift "$DOTFILES_DIR/managed_map.txt" || true
    if [[ ${#DRIFTED_FILES[@]} -gt 0 ]]; then
        echo ""
        warn "Skipping ${#DRIFTED_FILES[@]} drifted file(s) — backfill to sources, or --force to overwrite. Continuing."
    fi
fi

# Apply managed files (base + private overlay merge); drifted files are skipped.
apply_managed_files "$DOTFILES_DIR/managed_map.txt"

echo ""

# Setup mise
if command -v mise &>/dev/null; then
    echo "Setting up mise..."
    mise trust "$DOTFILES_DIR" 2>/dev/null || true
    mise install 2>/dev/null || true
    echo ""
fi

# Install Claude Code (native binary)
install_claude_code() {
    echo "Claude Code:"

    # Clean up legacy npm installation via mise
    if [[ -n "$(mise ls "npm:@anthropic-ai/claude-code" 2>/dev/null)" ]]; then
        info "Removing legacy npm installation from mise..."
        mise uninstall "npm:@anthropic-ai/claude-code" 2>/dev/null || true
        mise reshim 2>/dev/null || true
    fi

    # Check if native binary exists at expected location
    if [[ -x "$HOME/.local/bin/claude" ]] && file "$HOME/.local/bin/claude" 2>/dev/null | grep -q "Mach-O\|ELF"; then
        ok "Native binary installed ($(claude --version 2>/dev/null || echo 'unknown'))"
        return 0
    fi

    # Migrate an existing (e.g. legacy) claude on PATH to the native binary
    if command -v claude &>/dev/null; then
        info "Installing native binary via 'claude install'..."
        claude install || { err "Native install failed"; return 1; }
        ok "Native binary installed"
        return 0
    fi

    # Fresh machine: no claude anywhere. Bootstrap via the official installer.
    info "Installing native binary via official installer..."
    if curl -fsSL https://claude.ai/install.sh | bash; then
        ok "Native binary installed ($("$HOME/.local/bin/claude" --version 2>/dev/null || echo 'unknown'))"
    else
        err "Native install failed — install manually: https://docs.anthropic.com/en/docs/claude-code/getting-started"
        return 1
    fi
}
install_claude_code
echo ""

# Install/update the starship-claude statusline script (drives the Claude Code
# statusLine; reads ~/.claude/starship.toml which is symlinked from dotfiles).
# Pinned to a commit SHA (upstream ships no tags) — fetched from a mutable third
# party, so we pin for reproducibility and record the ref to stay idempotent.
STARSHIP_CLAUDE_REF="2029d40f084858103cc3062e307ca52a52e5d9f8"
install_starship_claude() {
    echo "Claude statusline (starship-claude):"

    if ! command -v starship &>/dev/null; then
        warn "starship not installed — skipping (install via Brewfile, then re-run dot sync)"
        return 0
    fi

    local dest="$HOME/.local/bin/starship-claude"
    local reffile="$HOME/.local/bin/.starship-claude.ref"

    # Idempotent: skip the fetch when the pinned ref is already installed.
    if [[ -x "$dest" && -f "$reffile" && "$(cat "$reffile" 2>/dev/null)" == "$STARSHIP_CLAUDE_REF" ]]; then
        ok "starship-claude installed (pinned ${STARSHIP_CLAUDE_REF:0:7})"
        return 0
    fi

    local url="https://raw.githubusercontent.com/martinemde/starship-claude/${STARSHIP_CLAUDE_REF}/plugin/bin/starship-claude"
    local tmp
    tmp="$(mktemp)"
    if curl -fsSL "$url" -o "$tmp" 2>/dev/null; then
        mkdir -p "$HOME/.local/bin"
        mv "$tmp" "$dest"
        chmod +x "$dest"
        printf '%s\n' "$STARSHIP_CLAUDE_REF" > "$reffile"
        ok "starship-claude installed (pinned ${STARSHIP_CLAUDE_REF:0:7})"
    else
        rm -f "$tmp"
        warn "Could not fetch starship-claude — check network or the upstream URL"
    fi
}
install_starship_claude
echo ""

echo "============================================"
echo -e "  ${GREEN}Sync complete!${NC}"
echo "============================================"
echo ""

echo ""
echo "Next steps:"
echo "  - Open nvim and run :Lazy update for plugins"
if [[ "$IS_FIRST_TIME" == true ]]; then
    echo "  - Run: dot health  (verify everything)"
    if has_private_overlay && has_private_git; then
        echo "  - Optionally push private overlay to a remote:"
        echo "      cd $PRIVATE_DIR && git remote add origin <url> && git push -u origin main"
    fi
fi
echo ""

