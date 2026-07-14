# Release checklist

This document outlines how to prepare and publish a git-ego release. Replace
`X.Y.Z` below with the release version.

## Pre-Release Checklist

### 1. Code Quality
- [ ] All tests pass: `go test -v ./...`
- [ ] Linting passes: `golangci-lint run`
- [ ] Code is formatted: `gofmt -s -d .` (should show no output)
- [ ] Static analysis passes: `go vet ./...`

### 2. Version Management
- [ ] Update version in `cmd/root.go`
- [ ] Update CHANGELOG.md with new version and date
- [ ] Ensure Go versions in `go.mod` and GitHub Actions are aligned

### 3. Documentation
- [ ] README.md reflects current features and requirements
- [ ] Installation instructions are up to date
- [ ] Examples work with current version

### 4. Build Verification
- [ ] Project builds successfully: `go build -v ./...`
- [ ] Binary works: `./git-ego --version`
- [ ] GitHub Actions CI passes

For credential-related changes, run the opt-in live smoke test after installing
the checkout. It creates and pushes a test commit, then restores the selected
profile:

```bash
go install .
GIT_EGO_BIN="$(go env GOPATH)/bin/git-ego" \
  GIT_EGO_SMOKE_RETURN_PROFILE=personal \
  scripts/gh-account-switch-smoke-test.sh --push
```

## Release Process

### 1. Create Release Tag
```bash
git tag -a vX.Y.Z -m "Release vX.Y.Z"
git push origin vX.Y.Z
```

### 2. GitHub Release
- [ ] The tag workflow creates a draft release and uploads platform archives.
- [ ] Review the generated notes and publish the draft release.
- [ ] Verify the Homebrew and Scoop workflows triggered by publishing succeed.
- [ ] Verify the WinGet workflow succeeds once the initial manifest PR has merged.

### 3. Verify Installation
```bash
go install github.com/bgreenwell/git-ego@latest
git ego --version
```

### 4. Post-Release
- [ ] Create "Unreleased" section in CHANGELOG.md
- [ ] Announce on social media/relevant channels

## Quality Gates

All items in the Pre-Release Checklist must be completed before creating a release tag. Any failing checks should be addressed before proceeding.
