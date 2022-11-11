if status is-interactive
  alias vim=nvim

  # Add dir for where homebrew applications get installed to if exists
  if test -e /opt/homebrew/bin/brew
    fish_add_path /opt/homebrew/bin
  end

  # Add /usr/local dirs to path if exist
  if test -e /usr/local/bin
    fish_add_path /usr/local/bin
  end
  if test -e /usr/local/sbin
    fish_add_path /usr/local/sbin
  end

  # environment variables for OOpinion development
  # TODO: uncomment after pass setup
  #set -x NODE_AUTH_TOKEN $(pass show OOPinion/NODE_AUTH_TOKEN)
  set -x VITE_API_URL http://localhost:3000
end
