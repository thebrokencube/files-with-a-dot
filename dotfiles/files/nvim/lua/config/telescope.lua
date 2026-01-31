local status_ok, telescope = pcall(require, 'telescope')
if not status_ok then
  vim.notify('Unable to load telescope')
  return
end

telescope.load_extension('fzf')
