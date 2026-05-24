// cmd/install_hook.go
package cmd

import (
	"bufio"
	"fmt"
	"os"
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
# If there's a mismatch, it will prompt you before committing.
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
	Run: func(cmd *cobra.Command, args []string) {
		gitRoot, err := findGitRoot(".")
		if err != nil {
			fmt.Println("Error: Not a git repository (or any of the parent directories).")

			return
		}

		hooksDir := filepath.Join(gitRoot, ".git", "hooks")
		// It's possible the hooks directory doesn't exist in a fresh git init.
		if err := os.MkdirAll(hooksDir, executableFilePermissions); err != nil {
			fmt.Printf("Error: Could not create hooks directory: %v\n", err)

			return
		}

		hookPath := filepath.Join(hooksDir, "pre-commit")

		// --- New, smarter hook installation logic ---
		if _, err := os.Stat(hookPath); err == nil {
			// File exists, so we need to check its content.
			content, err := os.ReadFile(hookPath)
			if err != nil {
				fmt.Printf("Error: Could not read existing pre-commit hook: %v\n", err)

				return
			}

			if strings.Contains(string(content), "internal check-commit") {
				fmt.Printf("✓ %s pre-commit hook is already installed.\n", binaryName)

				return
			}

			// Hook exists but is missing our command. Ask to append.
			fmt.Printf("A pre-commit hook already exists. Append %s check? [Y/n]: ", binaryName)
			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')

			if strings.TrimSpace(strings.ToLower(response)) == "n" {
				fmt.Println("\nInstall cancelled. Please manually add the following line to your pre-commit hook:")
				fmt.Printf("  %s internal check-commit\n", binaryName)

				return
			}

			// User confirmed. Append to the existing file.
			f, err := os.OpenFile(hookPath, os.O_APPEND|os.O_WRONLY, executableFilePermissions)
			if err != nil {
				fmt.Printf("Error: Failed to open existing hook for appending: %v\n", err)

				return
			}
			defer func() {
				if err := f.Close(); err != nil {
					fmt.Printf("Warning: Failed to close hook file: %v\n", err)
				}
			}()

			if _, err := f.WriteString(hookScript()); err != nil {
				fmt.Printf("Error: Failed to append to existing hook: %v\n", err)

				return
			}
			fmt.Printf("✓ %s check appended successfully to %s\n", binaryName, hookPath)

		} else {
			// File does not exist, create a new one.
			// Prepend the shebang for a new script.
			newHookContent := "#!/bin/sh" + hookScript()
			err = os.WriteFile(hookPath, []byte(newHookContent), executableFilePermissions)
			if err != nil {
				fmt.Printf("Error installing hook: %v\n", err)

				return
			}
			fmt.Printf("✓ %s pre-commit hook installed successfully in %s\n", binaryName, hookPath)
		}
	},
}

// findGitRoot searches for the root of the git repository.
func findGitRoot(startDir string) (string, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}

		dir = parent
	}

	return "", fmt.Errorf("git root not found")
}

func init() {
	rootCmd.AddCommand(installHookCmd)
}
