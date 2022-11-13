if status is-interactive
  # Load the default .profile
  bass source ~/.profile

  alias vim=nvim

  # Load node version on directory change
  fnm env --use-on-cd | source

  # environment variables for OOpinion development
  set -x NODE_AUTH_TOKEN $(pass show files-with-a-dot/oopinion/NODE_AUTH_TOKEN)
  set -x VITE_API_URL http://localhost:3000
end
