---------------------
-- general options --
---------------------
---- disable netrw
vim.g.loaded_netrw = 1
vim.g.loaded_netrwPlugin = 1

---- don't use swap files
vim.opt.swapfile = false

---- indentation (2 spaces, no tabs)
local indent = 2
vim.opt.tabstop = indent
vim.opt.softtabstop = indent
vim.opt.shiftwidth = indent
vim.opt.expandtab = true
vim.opt.smarttab = true
vim.opt.smartindent = true
vim.opt.autoindent = true

---- show line numbers
vim.opt.number = true
vim.opt.relativenumber = true

---- disable background buffers
vim.opt.hidden = false

---- auto read files changed outside vim
vim.opt.autoread = true

--- auto write on as many de-focusing actions as possible
vim.opt.autowriteall = true

---- enable the mouse
vim.opt.mouse = 'a'

---- use the system clipboard
vim.opt.clipboard = 'unnamedplus'

---- start scrolling before hitting the bottom
vim.opt.scrolloff = 5

---- ignore these when autocompleting paths
vim.opt.wildignore = vim.opt.wildignore + 'node_modules/*,vendor/bundle/*,tmp/*'

---- only use case sensitive search when uppercase
vim.opt.ignorecase = true
vim.opt.smartcase = true

---- persistent undo files
vim.opt.undodir = vim.fn.getenv('HOME')..'/.local/nvim/undofiles'
vim.opt.undofile = true

-- set termguicolors to enable highlight groups
vim.opt.termguicolors = true

-----------------
-- keybindings --
-----------------

--- use space as the leader key
vim.g.mapleader = ' '

---- split navigation
vim.keymap.set('n', '<c-h>', ':wincmd h<cr>')
vim.keymap.set('n', '<c-j>', ':wincmd j<cr>')
vim.keymap.set('n', '<c-k>', ':wincmd k<cr>')
vim.keymap.set('n', '<c-l>', ':wincmd l<cr>')

------------------------
-- initialize plugins --
------------------------

require('plugins')
