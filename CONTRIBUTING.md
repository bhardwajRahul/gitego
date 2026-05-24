# Contributing to gitego

First off, thank you for considering contributing to `gitego`! Your help is invaluable in making this tool better.

This document provides guidelines for contributing to the project. Please feel free to propose changes to this document in a pull request.

## Code of Conduct

This project and everyone participating in it is governed by our [Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code. Please report unacceptable behavior.

## How Can I Contribute?

### Reporting Bugs

If you find a bug, please ensure the bug was not already reported by searching on GitHub under [Issues](https://github.com/bgreenwell/git-ego/issues).

When you are creating a bug report, please include as many details as possible:
* Your operating system and version.
* The version of `gitego` you are using (`gitego --version`).
* The exact steps to reproduce the problem.
* The output you expected to see, and what you saw instead.

### Suggesting Enhancements

Enhancement suggestions are tracked as [GitHub issues](https://github.com/bgreenwell/git-ego/issues).
* Use a clear and descriptive title.
* Provide a step-by-step description of the suggested enhancement in as many details as possible.
* Explain why this enhancement would be useful to other `gitego` users.

### Pull Requests

We actively welcome your pull requests.

1.  Fork the repo and create your branch from `main`.
2.  If you've added code that should be tested, add tests.
3.  If you've changed APIs, update the documentation.
4.  Ensure the test suite passes (`go test ./...`).
5.  Make sure your code lints (`golangci-lint run`).
6.  Issue that pull request!

## Local Development Setup

To get started with development, you'll need [Go](https://go.dev/dl/) (version 1.24 or newer) installed.

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/bgreenwell/git-ego.git
    cd gitego
    ```

2.  **Build the binary:**
    ```bash
    go build .
    ```
    This will create a `gitego` (or `gitego.exe`) executable in the current directory.

3.  **Run the tests:**
    We have both unit tests and integration tests. To run all of them, use the following command from the root of the project:
    ```bash
    go test ./...
    ```
    The `-v` flag can be used for verbose output.

## Styleguides

### Git Commit Messages

* Use the present tense ("Add feature" not "Added feature").
* Use the imperative mood ("Move cursor to..." not "Moves cursor to...").
* Limit the first line to 72 characters or less.
* Reference issues and pull requests liberally after the first line.

### Go Styleguide

We adhere to the standard Go formatting and style guidelines. Please run `go fmt` on your code before committing. A linter is also configured for this project to ensure consistency.

---

Thank you again for your contribution!