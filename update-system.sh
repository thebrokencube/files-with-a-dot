#!/usr/bin/env bash
set -euo pipefail

PASS_STORE_PATH="$HOME/.password-store"
DOTFILES_PATH="$HOME/.dotfiles"

print () {
  echo
  echo "[update-system.sh] $1"
  echo
}

update_pass_store () {
  if ! [[ -e $PASS_STORE_PATH ]]; then
    print "$PASS_STORE_PATH does not exist; nothing to update."
  else
    print "Updating pass store at $PASS_STORE_PATH..."
    cd $PASS_STORE_PATH
    git pull

    print "Importing keys (note: interactive)..."
    ./import-keys.sh

    print "Done updating pass store."
    cd $HOME
  fi
}

update_dotfiles () {
  if ! [[ -e $DOTFILES_PATH ]]; then
    print "$DOTFILES_PATH does not exist; nothing to update."
  else
    print "Updating dotfiles at $DOTFILES_PATH..."
    cd $DOTFILES_PATH
    git pull

    print "Resetting symlinks..."
    cd ./dotfiles
    ./cleanup_dotfiles.sh
    ./install_dotfiles.sh

    print "Done updating dotfiles"
    cd $HOME
  fi
}

update_mac_software_updates () {
  print "Updating Mac Software Updates..."
  softwareupdate -l

  print "Done running softwareupdate."
}

update_homebrew_packages () {
  print "Updating homebrew packages..."
  brew bundle --global install --verbose

  print "Done updating homebrew packages. It may be worth checking to see if cleanup would be useful."
}

update_mas_packages () {
  print "Updating Mac App Store packages..."
  mas upgrade

  print "Done updating Mac App Store packages."
}

show_potential_homebrew_clenup () {
  print "Running homebrew cleanup in dry-run mode..."
  brew bundle --global cleanup

  print "Run \`brew bundle --global cleanup --zap --force\` if you want to actually clean any out of date packages/config."
}

main () {
  update_pass_store
  update_dotfiles

  update_mac_software_updates
  update_homebrew_packages
  update_mas_packages

  show_potential_homebrew_clenup
}

main
