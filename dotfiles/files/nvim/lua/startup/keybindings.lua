--- use space as the leader key
vim.g.mapleader = ' '

---- split navigation for dvorak, no movement
vim.keymap.set('n', '<c-h>', '<c-w>h')
vim.keymap.set('n', '<c-t>', '<c-w>j')
vim.keymap.set('n', '<c-n>', '<c-w>k')
vim.keymap.set('n', '<c-s>', '<c-w>l')

-- telescope
local builtins = require('telescope.builtin')
vim.keymap.set('n', '<leader>t', builtins.find_files)
vim.keymap.set('n', '<leader>g', builtins.live_grep)
vim.keymap.set('n', '<leader>b', builtins.buffers)

-- nvim-tree
vim.keymap.set('n', '<leader>f', ':NvimTreeOpen %:p:h<CR>', { noremap = true })

-- buffers
vim.keymap.set('n', '<c-d>', ':bd<CR>')
