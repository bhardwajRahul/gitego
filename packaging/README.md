# Distribution packaging

This directory contains upstream-maintained packaging definitions. Release
automation should replace version, URLs, and SHA-256 values from GitHub assets
before submitting updates to the relevant package repositories.

The release-triggered workflows publish to the same owner-maintained channels
as the related projects in this workspace. Configure these secrets before the
next release:

- `HOMEBREW_TAP_TOKEN` for `bgreenwell/homebrew-tap`.
- `SCOOP_BUCKET_TOKEN` for `bgreenwell/scoop-bucket`.
- `WINGET_TOKEN` for `bgreenwell/winget-pkgs` and WinGet pull requests.

- `homebrew/git-ego.rb` — Homebrew formula template
- `scoop/git-ego.json` — Scoop manifest template
- `nix/flake.nix` — Nix flake for building from source
- `debian/` — Debian source-package metadata
