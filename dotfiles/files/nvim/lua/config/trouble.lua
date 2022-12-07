local status_ok, trouble = pcall(require, 'trouble')
if not status_ok then
  vim.notify('Unable to require trouble')
  return
end

trouble.setup({
  auto_close = true,
  auto_open = true,
  height = 5,
  mode = 'document_diagnostics',
  group = false,
  padding = false,
  indent_lines = false,
})
