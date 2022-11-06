#!/usr/bin/env bash
set -euo pipefail

### These are helpers for use in the install and cleanup scripts

# Map containing all source to destination files
#TODO: make this not reliant on where it's called from
SYMLINK_MAP_FILE="$PWD/symlink_map.txt"

# $1 should be the dotfile
# $2 should be the source_path
get_dotfile () {
  # find the matching line in the symlink map file
  local dotfile_unparsed=$(cat $SYMLINK_MAP_FILE | egrep "^$1:" | sed -r "s|^$1:([^ ]+).*|\\1|")
  # replace $HOME text with resolved $HOME value
  local dotfile_parsed=$(echo $dotfile_unparsed | sed "s|^\$HOME|$HOME|g")

  echo $dotfile_parsed
}

# $1 should be the dotfile name from the dotfiles folder
get_sourcefile () {
  #TODO: make this not reliant on where it's called from
  echo "$PWD/files/$1"
}
