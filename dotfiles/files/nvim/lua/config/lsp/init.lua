local status_ok, _ = pcall(require, 'lspconfig')
if not status_ok then
  vim.notify('Unable to require lspconfig')
  return
end

-- gutter space for lsp info on left
vim.cmd([[set signcolumn=yes]])
-- 300ms (instead of default 4000ms before CursorHold events fire (like hover text on errors)
vim.cmd([[set updatetime=300]])

require('config.lsp.mason')
require('config.lsp.handlers').setup()
