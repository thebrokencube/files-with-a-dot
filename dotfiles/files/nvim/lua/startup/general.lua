-- don't use swap files
vim.opt.swapfile = false

-- indentation (2 spaces, no tabs)
local indent = 2
vim.opt.tabstop = indent
vim.opt.softtabstop = indent
vim.opt.shiftwidth = indent
vim.opt.expandtab = true
vim.opt.smarttab = true
vim.opt.smartindent = true
vim.opt.autoindent = true

-- show line numbers
vim.opt.number = true
vim.opt.relativenumber = true

-- disable background buffers
vim.opt.hidden = false

-- auto read files changed outside vim
vim.opt.autoread = true

-- auto write on as many de-focusing actions as possible
vim.opt.autowriteall = true

-- enable the mouse
vim.opt.mouse = 'a'

-- use the system clipboard
vim.opt.clipboard = 'unnamedplus'

-- start scrolling before hitting the bottom
vim.opt.scrolloff = 5

-- ignore these when autocompleting paths
vim.opt.wildignore = vim.opt.wildignore + 'node_modules/*,vendor/bundle/*,tmp/*'

-- only use case sensitive search when uppercase
vim.opt.ignorecase = true
vim.opt.smartcase = true

-- persistent undo files
vim.opt.undodir = vim.fn.getenv('HOME') .. '/.local/nvim/undofiles'
vim.opt.undofile = true

-- set termguicolors to enable highlight groups (should be default for most terminals)
vim.opt.termguicolors = true

-- split default open behavior
vim.opt.splitbelow = true
vim.opt.splitright = true

-- set soft linebreak for text wrapping
vim.cmd([[ set wrap linebreak ]])
