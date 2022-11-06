# `fish` Playbooks

As powerful as [`fish` shell](http;//www.fishshell.com) is, it doesn't always support everything that `bash` or `zsh` do. For this, we have some quirks and limitations to highlight here.

TODO:
- [ ] `nvm` tab completion

#### Quirks: NVM

One such quirk is with [`nvm` in `fish` shell](https://eshlox.net/2019/01/27/how-to-use-nvm-with-fish-shell). To resolve these issues:
* We use [`bass`](https://github.com/edc/bass) as a utility to run shell scripts in fish
* A [wrapper function](../dotfiles/files/fish/functions/nvm.fish) was written so that we can call `nvm` from `fish`.
* Of note, this configuration is all separate from what `bash` and `zsh` are using for `nvm` configuration/setup/startup.
* Tab completion is not supported in this setup currently for `nvm` in `fish`.
