# Machine Setup Guide

## 1. Install Homebrew & Packages

[[Homebrew Playbooks]](./playbooks/homebrew.md)

We aim to use Homebrew to manage as many packages as possible (even ones like `git`), and to help with that we're utilizing [Homebrew bundle](https://github.com/Homebrew/homebrew-bundle) for helping install/uninstall packages. From a fresh machine, we can do the following:

```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/thebrokencube/files-with-a-dot/main/first-time-setup.sh)"
```

## 2. Configure GitHub access

Now that we have all the base packages installed, we can use this repository to get them configured. However, first we'll need to get SSH access setup for GitHub.

Go [here in the GitHub settings](https://github.com/settings/keys) and go through the process of setting up an SSH key for this computer and github. In particular:

1. [Generate an SSH key and add it to the ssh-agent](https://docs.github.com/en/authentication/connecting-to-github-with-ssh/generating-a-new-ssh-key-and-adding-it-to-the-ssh-agent)
2. [Add the new SSH key to GitHub](https://docs.github.com/en/authentication/connecting-to-github-with-ssh/adding-a-new-ssh-key-to-your-github-account)
3. [Test the SSH connection](https://docs.github.com/en/authentication/connecting-to-github-with-ssh/testing-your-ssh-connection)

Github has some good docs about how to manage this all [here](https://docs.github.com/en/authentication/connecting-to-github-with-ssh).

## 3. Set up shared `pass` store

If you are using [`pass`](https://www.passwordstore.org) for managing any secrets, you should make sure it's set up on the machine (whether it's a new one or a shared one, like [from this blog post](https://medium.com/@davidpiegza/using-pass-in-a-team-1aa7adf36592)). It's useful to do now if possible before the shells get set up (though you can work around it if you can't for some reason).

## 4. Install Dotfiles

1. Since `git` is now installed, we can first start with cloning this repository somewhere, such as:
```bash
git clone git@github.com:thebrokencube/files-with-a-dot.git ~/.dotfiles
```

2. Navigate to `~/.dotfiles/dotfiles` and then run `./install_dotfiles.sh` to create symlinks for all the configuration stored in this repository.

## 5. Setup Shells

At this point, GPG and `pass` should be set up, so we can finally set up our alternative shell environments.

### 5.1. `zsh`

`zsh` is the default terminal for macOS. However, if you want to use `zsh`, you'll need to run the following to switch the default terminal to it (Note: it will ask for the user's password):

```bash
chsh -s /bin/zsh
```

### 5.2. `fish`

[[`fish` Playbooks]](./playbooks/fish.md)

If you want to use [`fish` shell](http://www.fishshell.com), you'll need to do the following to switch the default terminal to it. First, figure out where homebrew installed fish to with:
```bash
which fish
```

Whatever that is, add it to the end of `/etc/shells` with:
```bash
sudo echo "<insert path>" >> /etc/shells
```

And then run the following to change the default terminal (Note: it will ask for your password):
```bash
chsh -s <insert path>
```

## 6. Setup `vim`

[[`vim` Playbooks]](./playbooks/vim.md)

Once the shells are setup, we can setup vim and its dependencies. Of note, all shell configurations have `alias vim=nvim`, so we just use `vim` in all of these examples but are actually using Neovim.

1. Install [Vundle](https://github.com/VundleVim/Vundle.vim)
1. Open `vim` and run `:PluginInstall`. You may see some warnings on the first time, but they should go away after this step.
2. Once finished, close and re-open `vim` and it should be set up.

## 7. Setup Karabiner Elements

We'll need to start up Karabiner Elements to ensure that the settings have been loaded properly now that it's both installed and configured. To do so, simply open up Karabiner Elements via Spotlight and give it all the permissions it needs. For more information or debugging info, please see [the docs](https://karabiner-elements.pqrs.org/docs/manual/).

## 8. Amethyst (tiling window manager)

You should just need to open up Amethyst and manage preferences within the application for now. In the future maybe we can look into using [`.amethyst.yml`](https://github.com/ianyh/Amethyst/issues/1283), but for now I haven't made any keyboard configuration changes yet.

## 9. Update System

At this point, everything should be installed and setup, and you will want to just run the `update-system.sh` script. This will:
* update the pass-store git repo if it exists in `~/.password-store`
* import public keys for `~/.password-store/files-with-a-dot` into GPG interactively
* update the dotfiles git repo if it exists in `~/.dotfiles`
* reset all symlinks (run the cleanup and then install scripts)
* install mac software updates
* install brew package updates
* install mac app store package updates
* print out brew cleanup output if there would be any

This should be enough to run pretty regularly and keep systems up to date with minimal manual work.
