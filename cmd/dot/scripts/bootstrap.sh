#!/bin/bash
# bootstrap.sh - First-time setup for a fresh machine
# Run via: /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/thebrokencube/files-with-a-dot/main/bootstrap.sh)"
#
# What it does:
#   1. Checks for git (Xcode CLT or standalone)
#   2. Clones dotfiles repo to ~/.dotfiles
#   3. Checks for Homebrew (offers to install if missing)
#   4. Runs sync.sh if Homebrew is available
#
# Usage:
#   ./bootstrap.sh                                    # Interactive mode
#   ./bootstrap.sh --machine conservative             # Non-interactive
#   ./bootstrap.sh --machine aggressive --git-name "Name" --git-email "email@example.com"
#
# All flags are passed through to sync.sh

set -e

DOTFILES_REPO="git@github.com:thebrokencube/files-with-a-dot.git"
DOTFILES_TARGET="$HOME/.dotfiles"
FORCE_CLEAN=false

# Parse arguments
INSTALL_ARGS=""
while [[ $# -gt 0 ]]; do
    case $1 in
        --clean)
            FORCE_CLEAN=true
            shift
            ;;
        --clone-dir)
            DOTFILES_TARGET="$2"
            INSTALL_ARGS="$INSTALL_ARGS $1 $2"
            shift 2
            ;;
        --help|-h)
            echo "Usage: ./bootstrap.sh [OPTIONS]"
            echo ""
            echo "Bootstrap dotfiles on a fresh machine."
            echo ""
            echo "Options:"
            echo "  --clean              Remove existing dotfiles before starting"
            echo "  --machine TYPE       Set machine type (aggressive|conservative)"
            echo "  --git-name NAME      Set git user name"
            echo "  --git-email EMAIL    Set git user email"
            echo "  --clone-dir DIR      Clone repo to DIR instead of ~/.dotfiles"
            echo ""
            echo "Environment variables:"
            echo "  DOTFILES_MACHINE, DOTFILES_GIT_NAME, DOTFILES_GIT_EMAIL"
            echo "  NONINTERACTIVE=1     Skip prompts"
            exit 0
            ;;
        *)
            INSTALL_ARGS="$INSTALL_ARGS $1"
            shift
            ;;
    esac
done

echo "============================================"
echo "  Dotfiles Bootstrap"
echo "============================================"
echo ""

# ============================================================================
# Install Xcode Command Line Tools (macOS)
# ============================================================================

if [[ "$(uname)" == "Darwin" ]]; then
    if xcode-select -p &>/dev/null; then
        echo "Xcode Command Line Tools installed."
    else
        echo "Installing Xcode Command Line Tools..."
        echo "  A dialog will appear - please click Install and wait for it to finish."
        echo ""

        set +e
        xcode-select --install 2>/dev/null
        XCODE_RESULT=$?
        set -e

        if [[ $XCODE_RESULT -eq 0 ]]; then
            echo "  Waiting for installation..."
            until xcode-select -p &>/dev/null; do
                sleep 5
            done
            echo "  Xcode Command Line Tools installed."
        else
            echo "  Note: xcode-select --install needs to be run manually."
            echo "  Continuing anyway - Homebrew will install Command Line Tools if needed."
        fi
    fi
    echo ""
fi

# ============================================================================
# Pre-flight cleanup check
# ============================================================================

# Check if dotfiles directory already exists
if [[ -e "$DOTFILES_TARGET" ]] && [[ "$FORCE_CLEAN" == false ]] && [[ "${NONINTERACTIVE:-0}" != "1" ]]; then
    echo "Existing dotfiles directory detected: $DOTFILES_TARGET"
    echo ""
    echo "Options:"
    echo "  1) Continue (pull latest if git repo, or abort if not)"
    echo "  2) Clean and start fresh (remove and re-clone)"
    echo "  3) Abort"
    read -p "Choose [1-3]: " choice

    case "$choice" in
        1)
            echo "Continuing..."
            ;;
        2)
            FORCE_CLEAN=true
            echo "Removing $DOTFILES_TARGET..."
            rm -rf "$DOTFILES_TARGET"
            ;;
        3|*)
            echo "Aborted."
            exit 0
            ;;
    esac
    echo ""
elif [[ "$FORCE_CLEAN" == true ]] && [[ -e "$DOTFILES_TARGET" ]]; then
    echo "Cleaning existing dotfiles..."
    rm -rf "$DOTFILES_TARGET"
    echo ""
fi

# Note: Detailed conflict detection (shell configs, symlinks, etc.) is handled
# by sync.sh, which has access to symlink_map.txt and can do proper friction checks.

# ============================================================================
# Clone dotfiles repository (do this early, before needing Homebrew)
# ============================================================================

# First check if we have git available
HAS_GIT=false
if command -v git &>/dev/null; then
    HAS_GIT=true
    echo "Git already installed."
elif [[ "$(uname)" == "Darwin" ]] && xcode-select -p &>/dev/null; then
    # On macOS, Xcode Command Line Tools includes git
    HAS_GIT=true
    echo "Git available via Xcode Command Line Tools."
fi

if [[ "$HAS_GIT" == false ]]; then
    echo ""
    echo "============================================"
    echo "  Git Not Available"
    echo "============================================"
    echo ""
    echo "Bootstrap requires git to clone the dotfiles repository."
    echo ""
    echo "On macOS, git is included with Xcode Command Line Tools."
    echo "Install with: xcode-select --install"
    echo ""
    echo "Or install Homebrew first (includes git):"
    echo "  https://brew.sh"
    echo ""
    exit 1
fi

if [[ -L "$DOTFILES_TARGET" ]]; then
    # It's a symlink - follow it to the real repo
    REAL_DIR="$(readlink "$DOTFILES_TARGET")"
    echo ""
    echo "Dotfiles symlink exists at $DOTFILES_TARGET -> $REAL_DIR"
    echo "Pulling latest changes..."
    git -C "$REAL_DIR" pull
    DOTFILES_DIR="$REAL_DIR"
elif [[ -d "$DOTFILES_TARGET" ]]; then
    echo ""
    echo "Dotfiles directory already exists at $DOTFILES_TARGET"
    echo "Pulling latest changes..."
    git -C "$DOTFILES_TARGET" pull
    DOTFILES_DIR="$DOTFILES_TARGET"
else
    echo ""
    echo "Cloning dotfiles repository to $DOTFILES_TARGET..."
    git clone "$DOTFILES_REPO" "$DOTFILES_TARGET"
    DOTFILES_DIR="$DOTFILES_TARGET"
fi

echo ""

# ============================================================================
# Check Homebrew (required for sync.sh to install packages)
# ============================================================================

BREW_AVAILABLE=false

if command -v brew &>/dev/null; then
    BREW_AVAILABLE=true
    echo "Homebrew already installed."

    # Add Homebrew to PATH for this session
    if [[ -f "/opt/homebrew/bin/brew" ]]; then
        eval "$(/opt/homebrew/bin/brew shellenv)"
    elif [[ -f "/usr/local/bin/brew" ]]; then
        eval "$(/usr/local/bin/brew shellenv)"
    fi

    # Check if there are packages that might conflict (interactive only)
    PACKAGE_COUNT=$(brew list --formula 2>/dev/null | wc -l | tr -d ' ')
    if [[ "$PACKAGE_COUNT" -gt 0 ]] && [[ "${NONINTERACTIVE:-0}" != "1" ]]; then
        echo ""
        echo "Found $PACKAGE_COUNT Homebrew packages already installed."
        echo "You may want to clean up packages not in your Brewfiles before continuing."
        echo ""
        read -p "Run cleanup check now? [y/N] " cleanup_choice
        if [[ "$cleanup_choice" == "y" || "$cleanup_choice" == "Y" ]]; then
            echo ""
            "$DOTFILES_DIR/cleanup.sh" --force
            echo ""
        fi
    fi
else
    echo "Homebrew not found."
    echo ""
    echo "Homebrew is required to install packages (nvim, starship, etc.)"
    echo "Install from: https://brew.sh"
    echo ""
    echo "After installing Homebrew, run:"
    echo "  cd ~/.dotfiles && ./sync.sh $INSTALL_ARGS"
    echo ""

    if [[ "${NONINTERACTIVE:-0}" != "1" ]]; then
        echo "You can also try installing Homebrew now (requires sudo):"
        read -p "Install Homebrew now? [y/N] " install_brew

        if [[ "$install_brew" == "y" || "$install_brew" == "Y" ]]; then
            echo ""
            echo "Installing Homebrew..."
            echo "  Note: You will be prompted for your password"
            echo ""

            # Try to install Homebrew
            set +e
            /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
            BREW_RESULT=$?
            set -e

            if [[ $BREW_RESULT -eq 0 ]]; then
                # Add to PATH and verify
                if [[ -f "/opt/homebrew/bin/brew" ]]; then
                    eval "$(/opt/homebrew/bin/brew shellenv)"
                elif [[ -f "/usr/local/bin/brew" ]]; then
                    eval "$(/usr/local/bin/brew shellenv)"
                fi

                if command -v brew &>/dev/null; then
                    BREW_AVAILABLE=true
                    echo ""
                    echo "Homebrew installed successfully!"
                fi
            else
                echo ""
                echo "Homebrew installation failed or was cancelled."
                echo "You can install it manually later and then run sync.sh"
            fi
        fi
    fi
fi

echo ""

# ============================================================================
# Run sync script (if Homebrew is available)
# ============================================================================

if [[ "$BREW_AVAILABLE" == true ]]; then
    echo "Running sync..."
    "$DOTFILES_DIR/sync.sh" $INSTALL_ARGS

    echo ""
    echo "============================================"
    echo "  Bootstrap complete!"
    echo "============================================"
    echo ""
    echo "Run ~/.dotfiles/health.sh to verify setup."
else
    echo "============================================"
    echo "  Dotfiles Repository Cloned"
    echo "============================================"
    echo ""
    echo "Repository location: $DOTFILES_DIR"
    echo ""
    echo "To complete setup:"
    echo "  1. Install Homebrew: https://brew.sh"
    echo "  2. Run: cd ~/.dotfiles && ./sync.sh $INSTALL_ARGS"
    echo "  3. Open a new terminal (or run: exec \$SHELL -l)"
    echo ""
fi
