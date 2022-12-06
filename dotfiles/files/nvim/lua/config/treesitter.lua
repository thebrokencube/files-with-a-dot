require('nvim-treesitter.configs').setup({
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
    'tsx',
    'typescript',
    'yaml',
  },
  highlight = { enable = true },
  textobjects = { enable = true },

  -- Plugins
  autotag = { enable = true },
  rainbow = {
    enable = true,
    extended_mode = true,
    max_file_lines = nil,
  },
})
