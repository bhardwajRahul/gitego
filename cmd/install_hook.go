// cmd/install_hook.go
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// hookScript returns the shell script fragment to write or append.
// Using the actual binary name ensures the hook works when invoked as git-ego.
func hookScript() string {
	return fmt.Sprintf(`
# gitego pre-commit hook
# This command checks your commit author against the expected profile.
# A mismatch aborts safely; correct it with "git ego use <profile> --local".
%s internal check-commit
`, binaryName)
}

const (
	// executableFilePermissions are the permissions for an executable file.
	executableFilePermissions = 0755
)

var installHookCmd = &cobra.Command{
	Use:   "install-hook",
	Short: "Installs the pre-commit hook to safeguard against misattributed commits.",
	Long: `Installs a pre-commit hook in the current Git repository.

This hook automatically runs before every commit to verify that your
commit author details match the expected profile for this directory.
This provides a powerful safety net against accidental misattributed commits.
If a pre-commit hook already exists, you will be asked whether to append.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		hooksDir, err := gitHooksPath()
		if err != nil {
			return fmt.Errorf("not a Git repository: %w", err)
		}

		// It's possible the hooks directory doesn't exist in a fresh git init.
		if err := os.MkdirAll(hooksDir, executableFilePermissions); err != nil {
			return fmt.Errorf("create hooks directory: %w", err)
		}

		hookPath := filepath.Join(hooksDir, "pre-commit")

		// --- New, smarter hook installation logic ---
		if _, err := os.Stat(hookPath); err == nil {
			// File exists, so we need to check its content.
			content, err := os.ReadFile(hookPath)
			if err != nil {
				return fmt.Errorf("read existing pre-commit hook: %w", err)
			}

			if strings.Contains(string(content), "internal check-commit") {
				fmt.Printf("✓ %s pre-commit hook is already installed.\n", binaryName)

				return nil
			}

			// Hook exists but is missing our command. Ask to append.
			fmt.Printf("A pre-commit hook already exists. Append %s check? [Y/n]: ", binaryName)
			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')

			if strings.TrimSpace(strings.ToLower(response)) == "n" {
				fmt.Println("\nInstall cancelled. Please manually add the following line to your pre-commit hook:")
				fmt.Printf("  %s internal check-commit\n", binaryName)

				return nil
			}

			// User confirmed. Append to the existing file.
			f, err := os.OpenFile(hookPath, os.O_APPEND|os.O_WRONLY, executableFilePermissions)
			if err != nil {
				return fmt.Errorf("open hook for append: %w", err)
			}
			defer func() {
				if err := f.Close(); err != nil {
					fmt.Printf("Warning: Failed to close hook file: %v\n", err)
				}
			}()

			if _, err := f.WriteString(hookScript()); err != nil {
				return fmt.Errorf("append hook: %w", err)
			}
			fmt.Printf("✓ %s check appended successfully to %s\n", binaryName, hookPath)
			return nil

		} else {
			// File does not exist, create a new one.
			// Prepend the shebang for a new script.
			newHookContent := "#!/bin/sh" + hookScript()
			err = os.WriteFile(hookPath, []byte(newHookContent), executableFilePermissions)
			if err != nil {
				return fmt.Errorf("install hook: %w", err)
			}
			fmt.Printf("✓ %s pre-commit hook installed successfully in %s\n", binaryName, hookPath)
			return nil
		}
	},
}

// gitHooksPath asks Git for the repository-specific hooks directory. This
// works for linked worktrees, submodules, and bare repositories.
func gitHooksPath() (string, error) {
	custom := exec.Command("git", "config", "--get", "core.hooksPath")
	if output, err := custom.Output(); err == nil {
		path := strings.TrimSpace(string(output))
		if filepath.IsAbs(path) {
			return path, nil
		}
		rootOutput, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
		if err != nil {
			return "", err
		}
		return filepath.Join(strings.TrimSpace(string(rootOutput)), path), nil
	}

	cmd := exec.Command("git", "rev-parse", "--git-path", "hooks")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	path := strings.TrimSpace(string(output))
	if !filepath.IsAbs(path) {
		path, err = filepath.Abs(path)
		if err != nil {
			return "", err
		}
	}
	return path, nil
}

func init() {
	rootCmd.AddCommand(installHookCmd)
}
