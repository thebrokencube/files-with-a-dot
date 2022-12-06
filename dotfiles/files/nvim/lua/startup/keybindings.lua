--- use space as the leader key
vim.g.mapleader = ' '

---- split navigation for dvorak, no movement
vim.keymap.set('n', '<c-h>', '<c-w>h')
vim.keymap.set('n', '<c-t>', '<c-w>j')
vim.keymap.set('n', '<c-n>', '<c-w>k')
vim.keymap.set('n', '<c-s>', '<c-w>l')
