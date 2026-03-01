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
| nvim-tree | File explorer | `<leader>e` (toggle), `<leader>f` (find current file) |
| telescope | Fuzzy finder | `<leader>sf` (search files), `<leader>sg` (grep) |
| treesitter | Syntax highlighting | automatic |
| mason + lspconfig | LSP support | automatic |
| tokyonight | Color scheme | automatic |

## LSP Server Management

LSP servers are managed in two ways:

| Type | Install Via | Examples | When to Use |
|------|-------------|----------|-------------|
| **Mason** | `ensure_installed` | `lua_ls`, `ts_ls`, `pyright` | Standalone or runtime-agnostic servers |
| **Project-local** | `vim.lsp.config` | `ruby_lsp` | Servers that need project's runtime/deps |

### Mason Servers (Global)
Defined in `mason_ensure` list in init.lua. These install once and work everywhere.

### Project-local Servers
Configured via `vim.lsp.config()` + `vim.lsp.enable()`. Run through bundler/venv to use project's versions.

Example for Ruby (requires `gem 'ruby-lsp'` in Gemfile):
```lua
vim.lsp.config('ruby_lsp', {
  cmd = { 'bundle', 'exec', 'ruby-lsp' },
  root_markers = { 'Gemfile', '.git' },
})
vim.lsp.enable('ruby_lsp')
```

### Adding a New LSP Server
1. **If standalone/runtime-agnostic**: Add to both `mason_ensure` (Mason name) and `ensure_installed` (lspconfig name)
2. **If project-sensitive**: Add `vim.lsp.config()` + `vim.lsp.enable()` block

### Cleanup
Run `:MasonCleanup` to remove Mason packages not in the `mason_ensure` list.

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
