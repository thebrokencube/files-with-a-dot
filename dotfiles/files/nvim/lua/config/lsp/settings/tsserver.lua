local status_ok, lspconfig = pcall(require, 'lspconfig')
if not status_ok then
  vim.notify('Unable to load lspconfig')
  return
end

return {
  setup = {
    root_dir = lspconfig.util.root_pattern('tsconfig.json'),
  },
}
