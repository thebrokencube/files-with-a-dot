if status is-interactive
  alias vim=nvim

  # environment variables for OOpinion development
  set -x NODE_AUTH_TOKEN $(pass show OOPinion/NODE_AUTH_TOKEN)
  set -x VITE_API_URL http://localhost:3000
end
