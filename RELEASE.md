# Release Checklist

This document outlines the steps to prepare and publish a new release of git-ego.

## Pre-Release Checklist

### 1. Code Quality
- [ ] All tests pass: `go test -v ./...`
- [ ] Linting passes: `golangci-lint run` 
- [ ] Code is formatted: `gofmt -d .` (should show no output)
- [ ] Static analysis passes: `go vet ./...`

### 2. Version Management
- [ ] Update version in `cmd/root.go`
- [ ] Update CHANGELOG.md with new version and date
- [ ] Ensure go.mod and GitHub Actions versions are aligned

### 3. Documentation
- [ ] README.md reflects current features and requirements
- [ ] Installation instructions are up to date
- [ ] Examples work with current version

### 4. Build Verification
- [ ] Project builds successfully: `go build -v ./...`
- [ ] Binary works: `./git-ego --version`
- [ ] GitHub Actions CI passes

## Release Process

### 1. Create Release Tag
```bash
git tag -a v0.1.2 -m "Release version 0.1.1"
git push origin v0.1.2
```

### 2. GitHub Release
- [ ] Create GitHub release from tag
- [ ] Include changelog entry in release notes
- [ ] Upload pre-built binaries (optional)

### 3. Verify Installation
```bash
go install github.com/bgreenwell/git-ego@latest
git ego --version
```

### 4. Post-Release
- [ ] Update version back to "dev" for continued development
- [ ] Create "Unreleased" section in CHANGELOG.md
- [ ] Announce on social media/relevant channels

## Quality Gates

All items in the Pre-Release Checklist must be completed before creating a release tag. Any failing checks should be addressed before proceeding.