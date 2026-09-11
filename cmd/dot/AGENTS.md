# dot

Use Bash 3.2 syntax. Keep the functional core in `lib/` and side effects in entry scripts. Every mutating path detects state, supports noninteractive execution, and offers `--dry-run` when preview is meaningful.

Run `dot validate` after script or map changes. `symlink_map.txt` is the source of symlink state; managed files are compiled from their declared sources, never edited at the destination.
