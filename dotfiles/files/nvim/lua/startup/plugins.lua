-------------------------------------
-- bootstrap and initialize packer --
-------------------------------------

local ensure_packer = function()
  local fn = vim.fn
  local install_path = fn.stdpath('data')..'/site/pack/packer/start/packer.nvim'
  if fn.empty(fn.glob(install_path)) > 0 then
    fn.system({'git', 'clone', '--depth', '1', 'https://github.com/wbthomason/packer.nvim', install_path})
    vim.cmd('packadd packer.nvim')
    return true
  end
  return false
end
local packer_bootstrap = ensure_packer()

-- Autocommand that reloads neovim whenever you save the plugins.lua file
vim.cmd [[
  augroup packer_user_config
    autocmd!
    autocmd BufWritePost plugins.lua source <afile> | PackerSync
  augroup end
]]

-- Use a protected call so we don't error out on first use
local status_ok, packer = pcall(require, 'packer')
if not status_ok then
  vim.notify('Unable to require packer')
  return
end

-- Have packer use a popup window
packer.init {
  display = {
    open_fn = function()
      return require('packer.util').float { border = 'rounded' }
    end,
  },
}

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
    config = function()
      require('config.treesitter')
    end,
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
    config = function()
      local telescope = require('telescope')
      local builtins = require('telescope.builtin')
      telescope.load_extension('fzf')

      vim.keymap.set('n', '<leader>t', builtins.find_files)
      vim.keymap.set('n', '<leader>lg', builtins.live_grep)
      vim.keymap.set('n', '<leader>b', builtins.buffers)
      -- TODO: is there a way to do something similar to search for whatever your cursor
      -- is on with live_grep? before with ack.vim i had
      --   nnoremap S :Ack! '\b<C-R><C-W>\b'<CR>
    end,
  })
  use({
    'nvim-telescope/telescope-fzf-native.nvim',
    run = 'make'
  })

  -------------------------
  -- tree file navigator --
  -------------------------

  use {
    'nvim-tree/nvim-tree.lua',
    requires = { 'nvim-tree/nvim-web-devicons' },
    config = function()
      require('nvim-tree').setup()

      vim.keymap.set('n', '<leader>f', ':NvimTreeOpen %:p:h<CR>', { noremap = true })
    end,
    tag = 'nightly' -- optional, updated every week. (see issue #1193)
  }

  ----------------
  -- statusline --
  ----------------

  use({
    'hoob3rt/lualine.nvim',
    requires = {
      'kyazdani42/nvim-web-devicons',
    },
    config = function()
      require('lualine').setup({
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
          },
          lualine_x = {
            'filetype',
          },
        },
      })
    end,
  })

  -----------------------------
  -- completion and snippets --
  -----------------------------

  use({
    'hrsh7th/nvim-cmp',
    config = function()
      require('config.cmp')
    end,
    requires = {
      'hrsh7th/cmp-buffer', -- buffer completion
      'hrsh7th/cmp-path', -- path completions
      'hrsh7th/cmp-cmdline', -- cmdline completions
      'saadparwaiz1/cmp_luasnip', -- snippet completions
    }
  })

  -- snippets
  use "L3MON4D3/LuaSnip" --snippet engine
  use "rafamadriz/friendly-snippets" -- a bunch of snippets to use


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
        color_icons = true;
        default = true;
      })
    end,
  })

  ----------------------------------------
  -- initialize packer if bootstrapping --
  ----------------------------------------
  if packer_bootstrap then
    require('packer').sync()
  end
end)
