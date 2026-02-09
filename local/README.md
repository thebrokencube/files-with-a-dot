# Local Configuration

This directory contains machine-specific configuration files.

## What's Tracked vs Ignored

| File | Tracked | Purpose |
|------|---------|---------|
| `README.md` | ✅ Yes | This documentation |
| `.gitkeep` | ✅ Yes | Ensures directory exists on fresh clones |
| `shell.managed` | ✅ Yes | Repo-controlled shell config (aggressive mode only) |
| `gitconfig.local` | ❌ No | Git user identity and signing configuration |
| `env.local` | ❌ No | Environment variables (API keys, tokens) |
| `shell.local` | ❌ No | Private shell aliases and functions |

## How It Works

**Aggressive mode**: Sources `shell.managed` (tracked) + `shell.local` (private)
**Conservative mode**: Sources only `shell.local` (private)

This allows the dotfiles repo to ship useful shell config in aggressive mode while
keeping your private customizations separate.

## Setup

Private files (`*.local`, `env.local`) are created automatically by `sync.sh` during
first-time setup. You can also create them manually from the templates in `templates/`.

## Security

Never commit secrets. All `*.local` files are gitignored.
