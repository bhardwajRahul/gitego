# Distribution packaging

This directory contains upstream-maintained packaging definitions. Release
automation should replace version, URLs, and SHA-256 values from GitHub assets
before submitting updates to the relevant package repositories.

- `homebrew/git-ego.rb` — Homebrew formula template
- `scoop/git-ego.json` — Scoop manifest template
- `nix/flake.nix` — Nix flake for building from source
- `debian/` — Debian source-package metadata
