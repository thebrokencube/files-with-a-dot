-- Sourced from https://github.com/LunarVim/Neovim-from-scratch/blob/06-LSP/lua/user/lsp/null-ls.lua

local status_ok, null_ls = pcall(require, 'null-ls')
if not status_ok then
  vim.notify('Unable to load null-ls')
  return
end

-- https://github.com/jose-elias-alvarez/null-ls.nvim/tree/main/lua/null-ls/builtins/formatting
local formatting = null_ls.builtins.formatting

null_ls.setup({
  debug = false,
  sources = {
    formatting.prettier,
    formatting.stylua,
  },
})
