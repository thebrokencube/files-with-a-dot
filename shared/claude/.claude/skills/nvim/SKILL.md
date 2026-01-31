---
name: nvim
description: Neovim configuration help - plugins, keybindings, LSP setup, troubleshooting. Use when asked about nvim, vim, editor config, or plugins.
---

# Neovim Configuration

## Commands
- `/nvim` - Show overview
- `/nvim plugins` - List installed plugins
- `/nvim keys` - Show key keybindings
- `/nvim add <plugin>` - Help add a plugin

## Config Location
`~/.config/nvim/init.lua` (symlinked from `~/.dotfiles/shared/nvim/.config/nvim/`)

## Stack
- **Base**: kickstart.nvim
- **Plugin manager**: lazy.nvim

## Key Plugins
| Plugin | Purpose | Key Binding |
|--------|---------|-------------|
| nvim-tree | File explorer | `<leader>e` |
| telescope | Fuzzy finder | `<leader>sf` (files), `<leader>sg` (grep) |
| treesitter | Syntax highlighting | automatic |
| mason + lspconfig | LSP support | automatic |
| lazygit.nvim | Git TUI | `<leader>gg` |
| tokyonight | Color scheme | automatic |

## To List Plugins (`/nvim plugins`)
Read `~/.config/nvim/init.lua` and extract plugins from the `lazy.setup()` call.

## To Add a Plugin (`/nvim add <plugin>`)
1. Edit `~/.dotfiles/shared/nvim/.config/nvim/init.lua`
2. Add plugin spec to the lazy.setup plugins table:
   ```lua
   { 'author/plugin-name', opts = {} },
   ```
3. Open nvim, run `:Lazy install`

## Troubleshooting
- **Plugins not loading**: Open nvim, run `:Lazy`
- **LSP not working**: Run `:LspInfo`, check `:Mason` for servers
- **Health check**: Run `:checkhealth` in nvim
