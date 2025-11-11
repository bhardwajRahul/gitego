// utils/git.go

package utils

import (
	"fmt"
	"os/exec"
	"strings"
)

// execCommand is a package-level variable that can be overridden in tests.
var execCommand = exec.Command

// GetEffectiveGitConfig runs 'git config <key>' without the --global flag.
// This correctly resolves the config value from local > global > system.
func GetEffectiveGitConfig(key string) (string, error) {
	// Use the package-level variable instead of exec.Command directly.
	cmd := execCommand("git", "config", key)

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

// SetGlobalGitConfig runs 'git config --global <key> <value>'.
// It sets a configuration value in the user's global .gitconfig file.
func SetGlobalGitConfig(key, value string) error {
	// Use the package-level variable here as well.
	cmd := execCommand("git", "config", "--global", key, value)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git command failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// UnsetGlobalGitConfig runs 'git config --global --unset <key>'.
// If the key is not set, git exits with status code 5; this is ignored.
func UnsetGlobalGitConfig(key string) error {
	cmd := execCommand("git", "config", "--global", "--unset", key)

	output, err := cmd.CombinedOutput()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			// "you try to unset an option which does not exist (ret=5)"
			if exitErr.ExitCode() == 5 {
				return nil
			}
		}

		return fmt.Errorf("git command failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}
