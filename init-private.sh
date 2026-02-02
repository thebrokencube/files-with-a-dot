#!/bin/bash
# init-private.sh - Initialize a private dotfiles overlay repo
#
# Creates ~/.dotfiles.private/ from the template structure and optionally
# initializes it as a git repo.
#
# Usage:
#   ./init-private.sh                    # Create private overlay
#   ./init-private.sh --no-git           # Create without git init
#   ./init-private.sh --dir /path/to/dir # Create in custom location

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PRIVATE_DIR="$HOME/.dotfiles.private"
INIT_GIT=true

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --no-git) INIT_GIT=false; shift ;;
        --dir) PRIVATE_DIR="$2"; shift 2 ;;
        --help|-h)
            echo "Usage: ./init-private.sh [OPTIONS]"
            echo ""
            echo "Initialize a private dotfiles overlay repository."
            echo ""
            echo "Options:"
            echo "  --no-git     Don't initialize as git repo"
            echo "  --dir PATH   Create in PATH instead of ~/.dotfiles.private"
            exit 0
            ;;
        *) echo "Unknown option: $1"; exit 1 ;;
    esac
done

# Colors
GREEN='\033[0;32m'
CYAN='\033[0;36m'
YELLOW='\033[0;33m'
NC='\033[0m'

echo "============================================"
echo "  Initialize Private Dotfiles Overlay"
echo "============================================"
echo ""
echo "Location: $PRIVATE_DIR"
echo ""

# Check if already exists
if [[ -d "$PRIVATE_DIR" ]]; then
    echo -e "${YELLOW}Warning: $PRIVATE_DIR already exists${NC}"
    read -p "Overwrite? This will delete existing content. [y/N] " confirm
    if [[ "$confirm" != "y" && "$confirm" != "Y" ]]; then
        echo "Aborted."
        exit 0
    fi
    rm -rf "$PRIVATE_DIR"
fi

# Copy template
echo "Creating directory structure..."
cp -R "$SCRIPT_DIR/templates/private" "$PRIVATE_DIR"

# Create .gitkeep files for empty directories
touch "$PRIVATE_DIR/work/skills/.gitkeep"
touch "$PRIVATE_DIR/personal/skills/.gitkeep"

# Create empty shared files if not from template
[[ ! -f "$PRIVATE_DIR/symlink_map.txt" ]] && echo "# Shared private symlinks" > "$PRIVATE_DIR/symlink_map.txt"
[[ ! -f "$PRIVATE_DIR/Brewfile" ]] && echo "# Shared private packages" > "$PRIVATE_DIR/Brewfile"

# Initialize git repo
if [[ "$INIT_GIT" == true ]]; then
    echo ""
    echo "Initializing git repository..."
    cd "$PRIVATE_DIR"
    git init

    # Create .gitignore
    cat > .gitignore << 'EOF'
# OS files
.DS_Store
*.swp
*~

# Secrets (if any accidentally added)
*.secret
*.key
EOF

    # Initial commit
    git add -A
    git commit -m "Initial private dotfiles overlay"

    echo ""
    echo -e "${GREEN}Git repository initialized.${NC}"
    echo ""
    echo "To push to a remote:"
    echo "  cd $PRIVATE_DIR"
    echo "  git remote add origin <your-private-repo-url>"
    echo "  git push -u origin main"
fi

echo ""
echo "============================================"
echo -e "  ${GREEN}Private overlay created!${NC}"
echo "============================================"
echo ""
echo "Structure:"
echo "  $PRIVATE_DIR/"
echo "  ├── shared/         # All profiles"
echo "  ├── work/           # Work profile"
echo "  │   ├── skills/     # Work Claude skills"
echo "  │   ├── Brewfile    # Work packages"
echo "  │   └── symlink_map.txt"
echo "  └── personal/       # Personal profile"
echo ""
echo "Next steps:"
echo "  1. Add your private configs to work/ or personal/"
echo "  2. Run: ~/.dotfiles/sync.sh --profile work"
echo "  3. Optionally push to a private git remote"
