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

// GetEffectiveGitConfigWithOrigin returns the winning value and the file or
// scope that supplied it. Git quotes unusual origins, so split only once.
func GetEffectiveGitConfigWithOrigin(key string) (value, origin string, err error) {
	cmd := execCommand("git", "config", "--show-origin", "--get", key)
	output, err := cmd.Output()
	if err != nil {
		return "", "", err
	}
	line := strings.TrimSpace(string(output))
	fields := strings.SplitN(line, "\t", 2)
	if len(fields) != 2 {
		fields = strings.Fields(line)
		if len(fields) < 2 {
			return "", "", fmt.Errorf("unexpected git config origin output")
		}
		return strings.Join(fields[1:], " "), fields[0], nil
	}
	return strings.TrimSpace(fields[1]), strings.TrimSpace(fields[0]), nil
}

func RepositoryRoot() (string, error) {
	cmd := execCommand("git", "rev-parse", "--show-toplevel")
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

// SetLocalGitConfig sets a repository-local Git configuration value.
func SetLocalGitConfig(key, value string) error {
	cmd := execCommand("git", "config", "--local", key, value)
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

// UnsetLocalGitConfig removes a repository-local configuration value.
func UnsetLocalGitConfig(key string) error {
	cmd := execCommand("git", "config", "--local", "--unset", key)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 5 {
			return nil
		}
		return fmt.Errorf("git command failed: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// GetGlobalGitConfigValues returns every global value for a multi-valued key.
func GetGlobalGitConfigValues(key string) ([]string, error) {
	cmd := execCommand("git", "config", "--global", "--get-all", key)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return nil, nil
		}
		return nil, err
	}
	values := strings.FieldsFunc(string(output), func(r rune) bool { return r == '\n' || r == '\r' })
	return values, nil
}
