-- Sourced from https://github.com/LunarVim/Neovim-from-scratch/blob/b9a2bc855bbd7a7d54c9e280ee875393e30cf1a6/lua/user/autopairs.lua
local status_ok, npairs = pcall(require, 'nvim-autopairs')
if not status_ok then
  vim.notify('Unable to load nvim-autopairs')
  return
end

npairs.setup({
  check_ts = true,
  ts_config = {
    lua = { 'string', 'source' },
    javascript = { 'string', 'template_string' },
    java = false,
  },
  disable_filetype = { 'TelescopePrompt', 'spectre_panel' },
  fast_wrap = {
    map = '<M-e>',
    chars = { '{', '[', '(', '"', "'" },
    pattern = string.gsub([[ [%'%"%)%>%]%)%}%,] ]], '%s+', ''),
    offset = 0, -- Offset from pattern match
    end_key = '$',
    keys = 'qwertyuiopzxcvbnmasdfghjkl',
    check_comma = true,
    highlight = 'PmenuSel',
    highlight_grey = 'LineNr',
  },
})

local ca_status_ok, cmp_autopairs = pcall(require, 'nvim-autopairs.completion.cmp')
if not ca_status_ok then
  vim.notify('Unable to load nvim-autopairs.completion.cmp')
  return
end

local cmp_status_ok, cmp = pcall(require, 'cmp')
if not cmp_status_ok then
  vim.notify('Unable to load cmp')
  return
end

cmp.event:on('confirm_done', cmp_autopairs.on_confirm_done({ map_char = { tex = '' } }))
