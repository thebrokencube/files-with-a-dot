#!/usr/bin/env bash
set -euo pipefail

source ./dotfiles_helpers.sh

print () {
  echo
  echo "[cleanup_dotfiles.sh] $1"
  echo
}

# $1 should be dotfile name from the dotfiles folder
unlink_dotfile() {
  local dotfile_path=$(get_dotfile $1)

  if [[ -z $dotfile_path ]]; then
    print "no mapping for $1; add to symlink_map.txt if missing"
  elif [[ -L $dotfile_path ]]; then
    unlink $dotfile_path
    print "unlinked $dotfile_path"
  else
    print "$dotfile_path: symbolic link does not exist"
  fi
}

main () {
  # Removes all symlinks associated to dotfiles in the repo
  local dotfiles=`ls -A ./files`
  for dotfile in $dotfiles
  do
    unlink_dotfile $dotfile
  done
}

main
