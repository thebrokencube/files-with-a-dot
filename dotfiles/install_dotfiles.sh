#!/usr/bin/env bash
set -euo pipefail

source ./dotfiles_helpers.sh

# $1 should be dotfile name from the dotfiles folder
symlink_dotfile () {
  local sourcefile_path=$(get_sourcefile $1)
  local dotfile_path=$(get_dotfile $1)
  local parent_dir=$(dirname $dotfile_path)

  if [[ -z $dotfile_path ]]; then
    echo "no mapping for $1; add to symlink_map.txt if missing"
  elif [[ -e $dotfile_path ]]; then
    echo "$dotfile_path: file already exists"
  else
    if ! [[ -e $parent_dir ]]; then
      echo "making parent dir..."
      mkdir -p $parent_dir
    fi
    ln -s $sourcefile_path $dotfile_path
    echo "symlinked $1 from $sourcefile_path -> $dotfile_path"
  fi
}

main () {
  # Adds all symlinks associated to dotfiles in the repo
  local dotfiles=`ls -A ./files`
  for dotfile in $dotfiles
  do
    echo "$dotfile"
    symlink_dotfile $dotfile
    echo
  done
}

main
