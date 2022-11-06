# `vim` Playbooks

`vim` is used for all console text editing, and has a few workflows for keeping things up to date. There's a lot to fix here, but the basics should be covered. This assumes running `nvim` in the base Terminal application on macOS.

* [Neovim](https://neovim.io/) is the editor of choice for `vim` usage (it's remapped via an alias in all shells as well).
* [`vim-plug`](https://github.com/junegunn/vim-plug) is used for managing plugins.
* [`the_silver_searcher`](https://github.com/ggreer/the_silver_searcher) is used for powering search

**TODO**
- [ ] Update vim-plug installation & usage instructions from a clean slate, figure out why i have to keep closing and re-opening to see changes.
- [ ] `~/.vimrc` settings cleanup
- [ ] replace `ag` with `rg` ([example](https://phelipetls.github.io/posts/extending-vim-with-ripgrep/))
- [ ] LSP server integration ([docs](https://neovim.io/doc/user/lsp.html#lsp))

## Playbook: Install/Remove a plugin

1. Edit `~/.vimrc` and add or remove the plugin in the correct section.
2. Close and re-open `vim` and run `:PlugInstall`.
3. Close and re-open `vim`; you should be set with the new plugin

## Playbook: Update plugins

1. Open `vim` and run `:PluginUpdate`.
2. Close and re-open `vim`; you should be set with fully re-installed plugins.
