// cmd/status_test.go

package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bgreenwell/git-ego/config"
	"github.com/spf13/cobra"
)

// setupStatusTestEnvironment creates a temporary directory structure and mock config for testing.
func setupStatusTestEnvironment(t *testing.T) (tempDir, workDir string, mockCfg *config.Config, cleanup func()) {
	tempDir, err := os.MkdirTemp("", "gitego-status-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	workDir = filepath.Join(tempDir, "work", "project")
	if err := os.MkdirAll(workDir, 0755); err != nil {
		t.Fatalf("Failed to create work dir: %v", err)
	}

	// Save original working directory to restore later
	originalWd, _ := os.Getwd()

	mockCfg = &config.Config{
		Profiles: map[string]*config.Profile{
			"work": {
				Name:  "Work User",
				Email: "work@example.com",
			},
			"global": {
				Name:  "Global User",
				Email: "global@example.com",
			},
		},
		AutoRules: []*config.AutoRule{
			// FIX: Add a trailing slash to the path to ensure correct prefix matching.
			{Path: filepath.Join(tempDir, "work") + string(os.PathSeparator), Profile: "work"},
		},
		ActiveProfile: "global",
	}

	cleanup = func() {
		if err := os.Chdir(originalWd); err != nil {
			t.Errorf("Failed to restore original working directory: %v", err)
		}
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Warning: Failed to remove temp directory (this is common on Windows): %v", err)
		}
	}

	return tempDir, workDir, mockCfg, cleanup
}

func TestStatusCommand(t *testing.T) {
	// 1. Setup test environment
	tempDir, workDir, mockCfg, cleanup := setupStatusTestEnvironment(t)
	defer cleanup()

	// 2. Create a test runner with mocked dependencies
	runner := &statusRunner{
		load: func() (*config.Config, error) {
			return mockCfg, nil
		},
		getGitConfig: func(key string) (string, error) { return "", nil },
	}

	// --- Scenario 1: Test inside the auto-rule directory ---
	t.Run("inside auto-rule directory", func(t *testing.T) {
		runner.getGitConfig = func(key string) (string, error) {
			if key == "user.name" {
				return "Work User", nil
			}

			return "work@example.com", nil
		}

		runStatusTestScenario(t, runner, workDir, "gitego auto-rule for profile 'work'", "Work User")
	})

	// --- Scenario 2: Test outside any auto-rule directory ---
	t.Run("outside auto-rule directory", func(t *testing.T) {
		runner.getGitConfig = func(key string) (string, error) {
			if key == "user.name" {
				return "Global User", nil
			}

			return "global@example.com", nil
		}

		runStatusTestScenario(t, runner, tempDir, "Global gitego default", "Global User")
	})
}

// runStatusTestScenario executes a status command test scenario with given parameters.
func runStatusTestScenario(t *testing.T, runner *statusRunner, targetDir, expectedSource, expectedUser string) {
	if err := os.Chdir(targetDir); err != nil {
		t.Fatalf("Failed to change directory to %s: %v", targetDir, err)
	}

	var buf bytes.Buffer

	statusCmd := &cobra.Command{}
	statusCmd.SetOut(&buf)
	if err := runner.run(statusCmd, []string{}); err != nil {
		t.Fatal(err)
	}

	output := buf.String()

	if !strings.Contains(output, expectedSource) {
		t.Errorf("Expected output to contain source '%s', but it didn't.\nOutput:\n%s", expectedSource, output)
	}

	if !strings.Contains(output, expectedUser) {
		t.Errorf("Expected output to contain name '%s', but it didn't.\nOutput:\n%s", expectedUser, output)
	}
}
