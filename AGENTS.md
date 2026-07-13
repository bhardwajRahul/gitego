# AGENTS.md for git-ego

@../CLAUDE.md

`git-ego` is a command-line tool that solves the common problem of managing
multiple git identities. It allows you to define separate profiles for work,
personal projects, and clients, and then automatically switch between them
based on your working directory.

The tool manages `user.name`, `user.email`, SSH keys, and personal access
tokens (PATs), acting as a unified and intelligent manager for your git
identity. It is built on native git features like `includeif` and credential
helpers, ensuring it works seamlessly without fighting against git's own
mechanisms.

## Architecture

The project is organized into three main packages, following standard Go
practices:

- **`cmd/`**: All CLI logic. Each command is its own file (e.g., `add.go`,
  `list.go`). Uses `spf13/cobra`; `root.go` sets up the main command.
- **`config/`**: Loading, saving, and managing user profiles and application
  settings, including secure keychain interaction. `config.go` defines the
  data structures for profiles and auto-activation rules, serialized with
  `gopkg.in/yaml.v3`. `keyring.go` (plus platform-specific
  `keyring_darwin.go`/`keyring_other.go`) manages secure credential storage
  via `github.com/zalando/go-keyring`.
- **`utils/`**: Helper functions that interact with the git CLI and file
  system. `git.go` executes git commands and parses their output — the core
  logic for setting git configuration values.

## Main dependencies

- `github.com/spf13/cobra` — CLI framework
- `github.com/zalando/go-keyring` — cross-platform credential store access
- `gopkg.in/yaml.v3` — YAML parsing/emitting
