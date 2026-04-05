# CLAUDE.md — sshpub/dotfiles-cli

Go CLI for the sshpub/dotfiles framework. Manages profiles, modules, caches, and symlinks.

## Build

- `make build` — builds to `dist/dotfiles`
- `make build-all` — cross-compiles for 9 targets
- `make test` — runs tests
- `make clean` — removes dist/

## Architecture

- `cmd/` — Cobra commands (one file per command group)
- `pkg/` — library packages (module, profile, symlink, installer, registry, migrate, selfupdate, ui)
- Version injected via `-ldflags` at build time

## Conventions

- JSON for all config, never YAML
- Companion repo: `sshpub/dotfiles` (the shell framework this CLI manages)
- Design specs and plans live in the companion repo under `docs/superpowers/`
