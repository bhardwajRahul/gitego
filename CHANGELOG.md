# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.3] - 2026-07-14

### Changed

- Consolidated PAT handling on the secure git-ego keyring and scoped credential
  helper; profile switching no longer modifies unrelated macOS credentials.
- Clarified that `git ego` and `git-ego` are equivalent user-facing commands,
  and documented host-scoped credential-helper setup.

### Added

- Added a live, opt-in GitHub two-account smoke test and expanded integration
  coverage for profile persistence, Git configuration, hooks, and credentials.

## [0.2.2] - 2026-07-13

### Changed

- Simplified README language and removed the retired Go Report Card badge.
- Added release-asset checksums to the Scoop and Homebrew package definitions.

## [0.2.1] - 2026-07-13

### Added

- Repository `.gitego` profile selection, `doctor --repair`, and opt-in PAT checks for `list`.
- Distribution packaging definitions for Homebrew, Scoop, and Nix.

## [0.2.0] - 2026-07-13

### Added

- Host-scoped HTTPS credentials via profile `hosts` and `--host`.
- `pat set`/`pat delete`, `auto list`/`auto rm`, `doctor`, `use --local`, and profile import/export commands.

### Changed

- PATs are now supplied only through standard input; `add --pat` and `edit --pat` were removed.
- User-facing command failures now return non-zero exit codes.

### Fixed

- Credential helpers no longer disclose tokens to unrelated hosts or on `store`/`erase` operations.
- Auto-rule, profile-edit, hook/worktree, signing-key, quoting, and macOS keychain drift issues.

## [0.1.2] - 2026-05-24

### Added

- **Signing key support**: New `--signing-key` flag on `add` and `edit` sets `user.signingkey` in the profile's gitconfig, so commits are properly signed and verified when switching identities
- **SSH key fix on profile switch**: `git-ego use` now correctly updates the SSH key when switching the global default profile

### Changed

- Renamed Go module to `github.com/bgreenwell/git-ego` to match the repository name
- Binary name is resolved at runtime, so the tool works correctly when invoked as either `gitego` or `git-ego` (git subcommand form)
- Updated all user-facing messages to use the actual binary name

### Fixed

- Restored `.golangci.yml` linter configuration (had been left as a stale backup file)
- Corrected `gitego` → `git-ego` references in README, CONTRIBUTING, RELEASE, and build script

## [0.1.1] - 2025-08-13

### Changed

- **Code Quality**: Comprehensive linting and error handling improvements
  - All error return values are now properly checked throughout the codebase
  - Error messages follow Go conventions (lowercase, no trailing punctuation)
  - Added proper error handling for file operations, formatting functions, and system calls
  - Enhanced defer blocks with proper error checking for resource cleanup
- **CI/CD**: Updated GitHub Actions workflow to use Go 1.24 for consistency
- **Documentation**: Enhanced developer guidelines in GEMINI.md with error handling best practices
- **Project Requirements**: Updated minimum Go version requirement to 1.24+

### Fixed

- Resolved all golangci-lint issues (errcheck and staticcheck violations)
- Fixed version mismatch between go.mod and GitHub Actions workflow

## [0.1.0] - 2025-06-25

### Added

- **`edit` Command**: New `gitego edit <profile_name>` command allows for modification of existing profiles, including their name, email, username, SSH key, and PAT.
- **`--version` Flag**: Added a `-v` / `--version` flag to the root command to print the application's version number.
- **Shell Completion**: Introduced a `gitego completion [shell]` command to generate auto-completion scripts for Bash, Zsh, Fish, and PowerShell, improving user experience and discoverability.
- **Configuration Validation**: Implemented a validation step on application startup that warns users about inconsistencies in their `config.yaml`, such as auto-switch rules that point to non-existent profiles.

### Changed

- **Enhanced `list` Command**: The `gitego list` command output is now a more informative table, indicating the active global profile with an asterisk (`*`) and showing which profiles have associated SSH keys or PATs.
- **Smarter Hook Installation**: The `gitego install-hook` command now detects existing `pre-commit` hooks. If a hook exists, it will prompt the user for permission to append the `gitego` command instead of failing.
- **Destructive Command Confirmation**: The `gitego rm` command now requires user confirmation before deleting a profile. A `--force` flag was added to allow bypassing this safety check in scripts.
- **Refactored Codebase**:
    - Centralized the logic for determining the active profile based on directory rules into a single `config.GetActiveProfileForCurrentDir()` function, removing duplicated code from the `status` and `check-commit` commands.
    - Consolidated platform-specific keychain logic. Common functions for `gitego`'s internal token vault are now shared, with only OS-specific credential helper logic remaining in separate files.
    - Moved all Git configuration functions (`get` and `set`) into the `utils` package for better code organization.

### Removed

- Removed redundant, local implementations of Git configuration and profile-finding logic from individual command files in favor of centralized helper functions.
