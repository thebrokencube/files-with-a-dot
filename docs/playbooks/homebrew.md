# Homebrew Playbooks

[Homebrew bundle](https://github.com/Homebrew/homebrew-bundle)

I haven't found very simple workflows around homebrew bundle, so these playbooks represent some of the processes i've gone through when testing.

## Playbook: Sanity Checks

1. To check for packages that may get removed based on not being part of the `Brewfile`, run:
```bash
brew bundle cleanup --global
```

2. To check for packages that haven't been installed yet from the `Brewfile`, run:
```bash
brew bundle install --global
```

## Playbook: Add a package

1. Run through the Sanity Checks (see above).
2. Add the desired package to `dotfiles/Brewfile`.
3. Install packages by running:
```bash
brew bundle install --global
```
4. As a dry-run to figure out what dependencies are missing, run:
```bash
brew bundle cleanup --global
```
5. Add those dependencies back into `dotfiles/Brewfile`.
6. Run through the Sanity Checks (see above).

## Playbook: Remove a package

1. Run through the Sanity Checks (see above).
2. Remove the desired package from `dotfiles/Brewfile`
3. As a dry-run to figure out what dependencies will be deleted, run:
```bash
brew bundle cleanup --global
```
4. Remove those dependencies from `dotfiles/Brewfile`, and then verify that the output is as expected by running:
```bash
brew bundle cleanup --global
```
6. Uninstall packages and their preferences by running:
```bash
brew bundle cleanup --global --zap --force
```
7. Run through the Sanity Checks (see above).
