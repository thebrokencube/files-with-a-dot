if status is-interactive
  # Load the default .profile
  bass source ~/.profile

  alias vim=nvim

  # Load node version on directory change
  fnm env --use-on-cd | source

  # environment variables for OOpinion development
  set -x NODE_AUTH_TOKEN $(pass show files-with-a-dot/oopinion/NODE_AUTH_TOKEN)
  set -x VITE_API_URL http://localhost:3000
  set -x PASS_FIREBASE_PRIVATE_KEY $(pass show files-with-a-dot/hobby-honer/FIREBASE_PRIVATE_KEY)
  set -x PASS_FIREBASE_CLIENT_EMAIL $(pass show files-with-a-dot/hobby-honer/FIREBASE_CLIENT_EMAIL)
end
