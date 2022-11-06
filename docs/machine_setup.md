# Machine Setup
## 0. HACK - Setup Brewfile Manually

We run into a small bootstrapping issue that this repository is managed via git, and we don't have git setup at this point. Since we only need one file at the beginning of the process (`~/.Brewfile`), we can take some manual steps in the short-term until we find a better solution.

1. Download the [`Brewfile`](./dotfiles/files/Brewfile) to the home folder as `~/.Brewfile`.
2. Once Homebrew has been installed and all packages are installed, come back and delete `~/.Brewfile`. We'll afterwards have the file managed by the dotfile scripts.

## 1. Install Homebrew

[[Homebrew Playbooks]](./playbooks/homebrew.md)

We aim to use Homebrew to manage as many packages as possible (even ones like `git`), and to help with that we're utilizing [Homebrew bundle](https://github.com/Homebrew/homebrew-bundle) for helping install/uninstall packages. For now, we can start with the following:

1. Install [homebrew](http://brew.sh).
2. Run `brew bundle install --global` (using `~/.Brewfile` which we setup above). This should get a lot of the system setup for use with all the base packages and programs that are used on a daily basis.
3. HACK - If you manually downloaded the `Brewfile` above, remove it before proceeding.

## 2. Install Dotfiles

Now that we have all the base packages installed, we can use this repository to get them configured.

1. Since `git` is now installed, we can first start with cloning this repository somewhere (e.g. `~/.dotfiles`).
2. Navigate to that folder, navigate to the `dotfiles` subdirectory, and then run `./install_dotfiles.sh` to create symlinks for all the configuration stored in this repository.

## 3. (WIP) Setup GPG and `pass`

[[`pass` Playbooks]](./playbooks/pass.md)
We use [`pass`](http://passwordstore.org) for managing secrets locally, and we use them in env variables during shell setup. There's still some things to figure out (e.g. how to automate setting up the gpg keys, how to distribute shared `pass` stores between machines), so referring to the [playbook](./playbooks/pass.md) may be useful as it evolves.

## 4. Setup Shells

At this point, GPG and `pass` should be set up, so we can finally set up our alternative shell environments.

### `zsh`

If you want to use `zsh`, you'll need to run the following to switch the default terminal to it (Note: it will ask for the user's password):

```bash
chsh -s /bin/zsh
```
