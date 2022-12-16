local status_ok, nvim_treesitter_configs = pcall(require, 'nvim-treesitter.configs')
if not status_ok then
  vim.notify('Unable to require nvim-treesitter.configs')
  return
end

nvim_treesitter_configs.setup({
  ensure_installed = {
    'bash',
    'css',
    'dockerfile',
    'go',
    'html',
    'javascript',
    'json',
    'lua',
    'python',
    'regex',
    'ruby',
    'svelte',
    'tsx',
    'typescript',
    'yaml',
  },
  highlight = { enable = true },
  textobjects = { enable = true },

  -- plugins
  autotag = { enable = true },
  autopairs = { enable = true },
  rainbow = {
    enable = true,
    extended_mode = true,
    max_file_lines = nil,
  },
  context_commentstring = {
    enable = true,
    enable_autocmd = false,
  },
})
