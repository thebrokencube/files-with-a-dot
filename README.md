# files-with-a-dot

This repository aims to:
* manage dotfiles and other machine-related configuration
* keep up-to-date instructions on [setting up a machine from scratch](./docs/machine_setup.md)
* keep up-to-date [playbooks for common tasks/issues](./docs/playbooks) for software managed by this repository

## Quickstart

If you're starting from a factory-reset installation (assuming the only thing you've done is run all the System Updates, you can run the following to setup everything (note: kinda; still a wip):

```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/thebrokencube/files-with-a-dot/main/setup.sh)"
```

## System Requirements

This repository assumes you are starting at nothing more than the following:

* a base macOS environment at the moment (apple or intel chip)
* `bash 3`

The following technologies end up being used for managing this repository:

* [Homebrew](https://brew.sh/) - Homebrew is used for managing packages on macOS, and we utilize [Homebrew bundle](https://github.com/Homebrew/homebrew-bundle) for helping define them in code via `~/.Brewfile`.
* `bash` scripts - For the initial scripts, we use `bash 3` compatible scripts so that we're not relying on anything that won't be there on a clean machine.
* [Git](https://www.git-scm.com/) - All of the dotfiles and related scripts/documentation are managed by `git`.
