local status_ok, lualine = pcall(require, 'lualine')
if not status_ok then
  vim.notify('Unable to require lualine')
  return
end

lualine.setup({
  options = {
    section_separators = '',
    component_separators = '',
    theme = 'wombat',
    globalstatus = true,
  },
  extensions = {
    'quickfix',
    'fugitive',
  },
  sections = {
    lualine_c = {
      {
        'filename',
        path = 1,
      },
      'lsp_progress',
    },
    lualine_x = {
      'filetype',
    },
  },
})
