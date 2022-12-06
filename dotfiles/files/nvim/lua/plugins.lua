----------------------
-- bootstrap packer --
----------------------

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

------------------------------
-- load plugins with packer --
------------------------------

require('packer').startup(function(use)
  -- use packer to manage itself
  use('wbthomason/packer.nvim')

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

      vim.keymap.set('n', '<leader>p', builtins.find_files)
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

  ---------------------------
  -- miscellaneous utilies --
  ---------------------------

  -- show function context as you scroll
  use({
    'romgrk/nvim-treesitter-context',
    requires = { 'nvim-treesitter/nvim-treesitter' },
  })

  -- strip whitespace on save
  use('itspriddle/vim-stripper')

  -- disable highlights automatically on cursor move
  use('romainl/vim-cool')

  --------------------------------
  -- miscellaneous tools/config --
  --------------------------------

  -- editorconfig
  use('editorconfig/editorconfig-vim')

  -- git and github basic integration
  use('tpope/vim-fugitive')
  use('tpope/vim-rhubarb')

  -- colorscheme
  use('michalbachowski/vim-wombat256mod')

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

-- set colorscheme
vim.cmd([[
  set termguicolors
  colorscheme wombat256mod
]])
