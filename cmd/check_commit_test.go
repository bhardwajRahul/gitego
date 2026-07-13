// cmd/check_commit_test.go

package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/bgreenwell/git-ego/config"
	"github.com/spf13/cobra"
)

// runCheckCommitTest is a helper to execute the check-commit command with mocks.
func runCheckCommitTest(t *testing.T, cfg *config.Config, gitEmail, userInput string) (exitCode int, stderr string) {
	t.Helper()

	exitSignal := make(chan int, 1)

	mockExit := func(code int) {
		exitSignal <- code
	}

	var stderrBuf bytes.Buffer

	runner := &checkCommitRunner{
		getGitConfig: func(key string) (string, error) {
			if key == "user.email" {
				return gitEmail, nil
			}

			return "", nil
		},
		loadConfig: func() (*config.Config, error) { return cfg, nil },
		stdin:      strings.NewReader(userInput),
		stderr:     &stderrBuf,
		exit:       mockExit,
	}

	runner.run(&cobra.Command{}, []string{})

	// Block until the mock exit function has been called.
	exitCode = <-exitSignal

	return exitCode, stderrBuf.String()
}

func TestCheckCommitCommand(t *testing.T) {
	// --- Definitive Fix: Create a realistic temporary directory structure for the test ---
	tempDir, err := os.MkdirTemp("", "gitego-check-commit-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Warning: Failed to remove temp directory (this is common on Windows): %v", err)
		}
	}()

	originalWd, _ := os.Getwd()

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalWd); err != nil {
			t.Errorf("Failed to restore original working directory: %v", err)
		}
	}()

	// Setup a mock config using an absolute path for the rule.
	mockCfg := &config.Config{
		Profiles: map[string]*config.Profile{
			"work": {Email: "work@example.com"},
		},
		AutoRules: []*config.AutoRule{
			{Path: tempDir, Profile: "work"},
		},
	}

	t.Run("when emails match", func(t *testing.T) {
		exitCode, _ := runCheckCommitTest(t, mockCfg, "work@example.com", "")
		if exitCode != 0 {
			t.Errorf("Expected exit code 0 for matching emails, but got %d", exitCode)
		}
	})

	t.Run("when emails mismatch and user aborts", func(t *testing.T) {
		// User types "y" or just presses Enter.
		exitCode, stderr := runCheckCommitTest(t, mockCfg, "other@email.com", "\n")

		if exitCode != 1 {
			t.Errorf("Expected exit code 1 when user aborts, but got %d", exitCode)
		}

		if !strings.Contains(stderr, "Commit aborted by user") {
			t.Errorf("Expected 'aborted' message in stderr, but it was missing. Got:\n%s", stderr)
		}
	})

	t.Run("when emails mismatch and user proceeds", func(t *testing.T) {
		// User types "n".
		exitCode, stderr := runCheckCommitTest(t, mockCfg, "other@email.com", "n\n")

		if exitCode != 0 {
			t.Errorf("Expected exit code 0 when user proceeds, but got %d", exitCode)
		}

		if !strings.Contains(stderr, "Commit proceeding with mismatched user") {
			t.Errorf("Expected 'proceeding' message in stderr, but it was missing. Got:\n%s", stderr)
		}
	})

	t.Run("when active profile email is overridden", func(t *testing.T) {
		mockCfg.ActiveProfile = "work"
		exitCode, stderr := runCheckCommitTest(t, mockCfg, "overridden@example.com", "\n")
		if exitCode != 1 || !strings.Contains(stderr, "Commit aborted by user") {
			t.Fatalf("expected active profile override to be checked, code=%d stderr=%q", exitCode, stderr)
		}
	})
}
