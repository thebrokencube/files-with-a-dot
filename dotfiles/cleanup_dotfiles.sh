#!/usr/bin/env bash
set -euo pipefail

source ./dotfiles_helpers.sh

# $1 should be dotfile name from the dotfiles folder
unlink_dotfile() {
  local dotfile_path=$(get_dotfile $1)

  if [ -z $dotfile_path ]; then
    echo "no mapping for $1; add to symlink_map.txt if missing"
  elif [ -L $dotfile_path ]; then
    unlink $dotfile_path
    echo "unlinked $dotfile_path"
  else
    echo "$dotfile_path: symbolic link does not exist"
  fi
}

main () {
  # Removes all symlinks associated to dotfiles in the repo
  local dotfiles=`ls -A ./files`
  for dotfile in $dotfiles
  do
    echo "$dotfile"
    unlink_dotfile $dotfile
    echo
  done
}

main
