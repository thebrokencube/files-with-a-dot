-- Sourced from https://github.com/LunarVim/Neovim-from-scratch/blob/122bedde844fcef84169889d7666af0592b58c46/lua/user/bufferline.lua

local status_ok, bufferline = pcall(require, 'bufferline')
if not status_ok then return end

bufferline.setup({
  options = {
    numbers = 'none', -- | "ordinal" | "buffer_id" | "both" | function({ ordinal, id, lower, raise }): string,
    close_command = 'Bdelete! %d', -- can be a string | function, see "Mouse actions"
    right_mouse_command = 'Bdelete! %d', -- can be a string | function, see "Mouse actions"
    left_mouse_command = 'buffer %d', -- can be a string | function, see "Mouse actions"
    middle_mouse_command = nil, -- can be a string | function, see "Mouse actions"
    indicator = { style = 'icon', icon = '▎' },
    buffer_close_icon = '',
    modified_icon = '●',
    close_icon = '',
    left_trunc_marker = '',
    right_trunc_marker = '',
    max_name_length = 30,
    max_prefix_length = 30, -- prefix used when a buffer is de-duplicated
    tab_size = 21,
    diagnostics = false, -- | "nvim_lsp" | "coc",
    diagnostics_update_in_insert = false,
    offsets = { { filetype = 'NvimTree', text = '', padding = 1 } },
    show_buffer_icons = true,
    show_buffer_close_icons = true,
    show_close_icon = true,
    show_tab_indicators = true,
    persist_buffer_sort = true, -- whether or not custom sorted buffers should persist
    separator_style = 'thick', -- | "thick" | "thin" | { 'any', 'any' },
    enforce_regular_tabs = true,
    always_show_bufferline = true,
  },
  highlights = {
    fill = {
      fg = { attribute = 'fg', highlight = '#ff0000' },
      bg = { attribute = 'bg', highlight = 'TabLine' },
    },
    background = {
      fg = { attribute = 'fg', highlight = 'TabLine' },
      bg = { attribute = 'bg', highlight = 'TabLine' },
    },

    buffer_selected = {
      fg = { attribute = 'fg', highlight = '#ff0000' },
      bg = { attribute = 'bg', highlight = '#0000ff' },
    },
    buffer_visible = {
      fg = { attribute = 'fg', highlight = 'TabLine' },
      bg = { attribute = 'bg', highlight = 'TabLine' },
    },

    close_button = {
      fg = { attribute = 'fg', highlight = 'TabLine' },
      bg = { attribute = 'bg', highlight = 'TabLine' },
    },
    close_button_visible = {
      fg = { attribute = 'fg', highlight = 'TabLine' },
      bg = { attribute = 'bg', highlight = 'TabLine' },
    },
    -- close_button_selected = {
    --   fg = {attribute='fg',highlight='TabLineSel'},
    --   bg ={attribute='bg',highlight='TabLineSel'}
    --   },

    tab_selected = {
      fg = { attribute = 'fg', highlight = 'Normal' },
      bg = { attribute = 'bg', highlight = 'Normal' },
    },
    tab = {
      fg = { attribute = 'fg', highlight = 'TabLine' },
      bg = { attribute = 'bg', highlight = 'TabLine' },
    },
    tab_close = {
      fg = { attribute = 'fg', highlight = 'TabLineSel' },
      bg = { attribute = 'bg', highlight = 'Normal' },
    },

    duplicate_selected = {
      fg = { attribute = 'fg', highlight = 'TabLineSel' },
      bg = { attribute = 'bg', highlight = 'TabLineSel' },
      italic = true,
    },
    duplicate_visible = {
      fg = { attribute = 'fg', highlight = 'TabLine' },
      bg = { attribute = 'bg', highlight = 'TabLine' },
      italic = true,
    },
    duplicate = {
      fg = { attribute = 'fg', highlight = 'TabLine' },
      bg = { attribute = 'bg', highlight = 'TabLine' },
      italic = true,
    },

    modified = {
      fg = { attribute = 'fg', highlight = 'TabLine' },
      bg = { attribute = 'bg', highlight = 'TabLine' },
    },
    modified_selected = {
      fg = { attribute = 'fg', highlight = 'Normal' },
      bg = { attribute = 'bg', highlight = 'Normal' },
    },
    modified_visible = {
      fg = { attribute = 'fg', highlight = 'TabLine' },
      bg = { attribute = 'bg', highlight = 'TabLine' },
    },

    separator = {
      fg = { attribute = 'bg', highlight = 'TabLine' },
      bg = { attribute = 'bg', highlight = 'TabLine' },
    },
    separator_selected = {
      fg = { attribute = 'bg', highlight = 'Normal' },
      bg = { attribute = 'bg', highlight = 'Normal' },
    },
    indicator_selected = {
      fg = { attribute = 'fg', highlight = 'LspDiagnosticsDefaultHint' },
      bg = { attribute = 'bg', highlight = 'Normal' },
    },
  },
})
