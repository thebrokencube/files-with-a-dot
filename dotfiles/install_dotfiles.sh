#!/usr/bin/env bash
set -euo pipefail

source ./dotfiles_helpers.sh

print () {
  echo
  echo "[install_dotfiles.sh] $1"
  echo
}

# $1 should be dotfile name from the dotfiles folder
symlink_dotfile () {
  local sourcefile_path=$(get_sourcefile $1)
  local dotfile_path=$(get_dotfile $1)
  local parent_dir=$(dirname $dotfile_path)

  if [[ -z $dotfile_path ]]; then
    print "no mapping for $1; add to symlink_map.txt if missing"
  elif [[ -e $dotfile_path ]]; then
    print "$dotfile_path: file already exists"
  else
    if ! [[ -e $parent_dir ]]; then
      print "making parent dir..."
      mkdir -p $parent_dir
    fi
    ln -s $sourcefile_path $dotfile_path
    print "symlinked $1 from $sourcefile_path -> $dotfile_path"
  fi
}

main () {
  # Adds all symlinks associated to dotfiles in the repo
  local dotfiles=`ls -A ./files`
  for dotfile in $dotfiles
  do
    symlink_dotfile $dotfile
  done
}

main
