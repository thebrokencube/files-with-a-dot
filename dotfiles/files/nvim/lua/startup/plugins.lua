-------------------------------------
-- bootstrap and initialize packer --
-------------------------------------

local ensure_packer = function()
  local fn = vim.fn
  local install_path = fn.stdpath('data') .. '/site/pack/packer/start/packer.nvim'
  if fn.empty(fn.glob(install_path)) > 0 then
    fn.system({ 'git', 'clone', '--depth', '1', 'https://github.com/wbthomason/packer.nvim', install_path })
    vim.cmd('packadd packer.nvim')
    return true
  end
  return false
end
local packer_bootstrap = ensure_packer()

-- Autocommand that reloads neovim whenever you save the plugins.lua file
vim.cmd([[
  augroup packer_user_config
    autocmd!
    autocmd BufWritePost plugins.lua source <afile> | PackerSync
  augroup end
]])

-- Use a protected call so we don't error out on first use
local status_ok, packer = pcall(require, 'packer')
if not status_ok then
  vim.notify('Unable to load packer')
  return
end

-- Have packer use a popup window
packer.init({
  display = {
    open_fn = function() return require('packer.util').float({ border = 'rounded' }) end,
  },
})

------------------------------
-- load plugins with packer --
------------------------------

packer.startup(function(use)
  use('wbthomason/packer.nvim') -- use packer to manage itself

  -------------------------
  -- syntax highlighting --
  -------------------------
  use({
    'nvim-treesitter/nvim-treesitter',
    run = function()
      local ts_update = require('nvim-treesitter.install').update({ with_sync = true })
      ts_update()
    end,
    config = function() require('config.treesitter') end,
    requires = {
      'andymass/vim-matchup', -- extended matchers for %
      'windwp/nvim-ts-autotag', -- auto-complete html tags
      'p00f/nvim-ts-rainbow', -- highlight parenthesis pairs w/ different colors
    },
  })

  -------------------------------------
  -- fuzzy finding and file browsing --
  -------------------------------------

  use({
    'nvim-telescope/telescope.nvim',
    requires = { 'nvim-lua/plenary.nvim' },
    config = function() require('config.telescope') end,
  })
  use({
    'nvim-telescope/telescope-fzf-native.nvim',
    run = 'make',
  })

  -------------------------
  -- tree file navigator --
  -------------------------

  use({
    'nvim-tree/nvim-tree.lua',
    requires = { 'nvim-tree/nvim-web-devicons' },
    config = function() require('nvim-tree').setup() end,
    tag = 'nightly', -- optional, updated every week. (see issue #1193)
  })

  ----------------
  -- bufferline --
  ----------------

  use({
    'akinsho/bufferline.nvim',
    config = function() require('config.bufferline') end,
    requires = { 'moll/vim-bbye' },
  })

  ----------------
  -- statusline --
  ----------------

  use({
    'hoob3rt/lualine.nvim',
    requires = {
      'kyazdani42/nvim-web-devicons',
      'arkav/lualine-lsp-progress',
    },
    config = function() require('config.lualine') end,
  })

  -----------------------------
  -- completion and snippets --
  -----------------------------

  use({
    'hrsh7th/nvim-cmp',
    config = function() require('config.cmp') end,
    requires = {
      'hrsh7th/cmp-buffer', -- buffer completion
      'hrsh7th/cmp-path', -- path completions
      'hrsh7th/cmp-cmdline', -- cmdline completions
      'saadparwaiz1/cmp_luasnip', -- snippet completions
      'hrsh7th/cmp-nvim-lsp', -- nvim-lsp completions
      'hrsh7th/cmp-nvim-lua', -- lua completions
    },
  })

  -- snippets
  use('L3MON4D3/LuaSnip') --snippet engine
  use('rafamadriz/friendly-snippets') -- a bunch of snippets to use

  ---------
  -- LSP --
  ---------

  use('neovim/nvim-lspconfig') -- enable LSP
  use('williamboman/mason.nvim') -- simple to use language server installer
  use('williamboman/mason-lspconfig.nvim') -- simple to use language server installer
  use('jose-elias-alvarez/null-ls.nvim') -- LSP diagnostics and code actions
  use('jose-elias-alvarez/typescript.nvim') -- enhanced typescript lsp server

  -- Quickfix list for lsp diagnostics
  use({
    'folke/trouble.nvim',
    requires = {
      'kyazdani42/nvim-web-devicons',
    },
    config = function() require('config.trouble') end,
  })

  -- Stabilize the trouble window
  use({
    'luukvbaal/stabilize.nvim',
    config = function() require('config.stabilize') end,
  })

  ---------------------------
  -- miscellaneous utilies --
  ---------------------------

  -- show function context as you scroll
  use({
    'romgrk/nvim-treesitter-context',
    requires = { 'nvim-treesitter/nvim-treesitter' },
  })

  use('itspriddle/vim-stripper') -- strip whitespace on save
  use('romainl/vim-cool') -- disable highlights automatically on cursor move

  -- easily comment stuff
  use({
    'numToStr/Comment.nvim',
    config = function() require('config.comment') end,
  })
  -- context comment strings
  use({
    'JoosepAlviste/nvim-ts-context-commentstring',
    requires = { 'nvim-treesitter/nvim-treesitter' },
  })

  --------------------------------
  -- miscellaneous tools/config --
  --------------------------------

  -- git and github basic integration
  use('tpope/vim-fugitive')
  use('tpope/vim-rhubarb')

  use('editorconfig/editorconfig-vim') -- editorconfig
  use('michalbachowski/vim-wombat256mod') -- colorscheme

  -- filetype icons
  use({
    'nvim-tree/nvim-web-devicons',
    config = function()
      require('nvim-web-devicons').setup({
        color_icons = true,
        default = true,
      })
    end,
  })

  -- toggleterm
  use({
    'akinsho/toggleterm.nvim',
    config = function() require('config.toggleterm') end,
  })

  ----------------------------------------
  -- initialize packer if bootstrapping --
  ----------------------------------------
  if packer_bootstrap then require('packer').sync() end
end)
