local status_ok, stabilize = pcall(require, 'stabilize')
if not status_ok then
  vim.notify('Unable to require stabilize')
  return
end

stabilize.setup({
  -- This is necessary to get stabilize to work in some cases.
  -- See https://github.com/luukvbaal/stabilize.nvim#note
  nested = 'QuickFixCmdPost,DiagnosticChanged *',
})
