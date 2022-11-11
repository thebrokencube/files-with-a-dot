#!/usr/bin/env bash
set -euo pipefail

BREWFILE_URL="https://raw.githubusercontent.com/thebrokencube/files-with-a-dot/main/dotfiles/files/Brewfile"

print () {
  echo "[setup.sh] $1"
}

# Validate starting state

validate_home_exists () {
  # Temporarily remove unset variable check while validating environment
  # variable existence
  set +u

  if [[ -z "$HOME" ]]; then
    print "ERROR: \$HOME environment variable does not exist; cannot run script."
    exit 1
  fi

  # Re-enable unset variable check
  set -u
}

validate_dotfiles_absent () {
  validate_home_exists

  #local dotfiles_folder="$HOME/.dotfiles.tmp"
  local dotfiles_folder="$HOME/.dotfiles"

  if [[ -e $dotfiles_folder ]]; then
    print "$dotfiles_folder exists."
    print "Check the docs in $dotfiles_folder for further instructions."
    exit 0
  fi
}

# Homebrew

get_tmp_brewfile_path () {
  echo "$HOME/.Brewfile.tmp"
}

download_brewfile () {
  local brewfile_path=$(get_tmp_brewfile_path)

  if [[ -e $brewfile_path ]]; then
    print "$brewfile_path already exists."
  else
    print "$brewfile_path does not exist. Downloading..."
    curl -o $brewfile_path $BREWFILE_URL
    print "$brewfile_path downloaded."
  fi
}

remove_brewfile () {
  local brewfile_path=$(get_tmp_brewfile_path)

  if [[ -e $brewfile_path ]]; then
    print "$brewfile_path exists. Removing..."
    rm $brewfile_path
    print "$brewfile_path removed."
  else
    print "$brewfile_path does not exist, nothing to remove."
  fi
}

ensure_homebrew_exists () {
  local brew_exists=$(command -v brew)

  if [[ -z "$brew_exists" ]]; then
    print "Homebrew not installed. Installing..."
    echo
    /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
    print "Homebrew installed."
  fi
}

# System Setup

install_with_homebrew () {
  print "Installing packages via Homebrew..."
  echo
  /opt/homebrew/bin/brew bundle install --verbose --file $(get_tmp_brewfile_path)
}

test_if_git_exists () {
  print "Running git -v to validate git is installed..."
  git -v
  print "Proceed to downloading the repository with git"
}

main () {
  validate_dotfiles_absent
  download_brewfile
  ensure_homebrew_exists
  install_with_homebrew
  remove_brewfile
  test_if_git_exists
}

main
