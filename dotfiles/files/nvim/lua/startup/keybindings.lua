--- use space as the leader key
vim.g.mapleader = ' '

---- split navigation for dvorak, no movement
vim.keymap.set('n', '<c-h>', '<c-w>h')
vim.keymap.set('n', '<c-t>', '<c-w>j')
vim.keymap.set('n', '<c-n>', '<c-w>k')
vim.keymap.set('n', '<c-s>', '<c-w>l')

-- telescope
local builtin_status_ok, builtin = pcall(require, 'telescope.builtin')
if not builtin_status_ok then
  vim.notify('Unable to load telescope.builtin')
  return
end
vim.keymap.set('n', '<leader>t', builtin.find_files)
vim.keymap.set('n', '<leader>g', builtin.live_grep)
vim.keymap.set('n', '<leader>b', builtin.buffers)

-- nvim-tree
vim.keymap.set('n', '<leader>f', ':NvimTreeOpen %:p:h<CR>', { noremap = true })

-- buffers
vim.keymap.set('n', '<c-d>', ':bd<CR>')
