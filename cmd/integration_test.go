// cmd/integration_test.go

package cmd_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var (
	// This will be the path to the compiled test binary.
	gitegoBinary string
)

// TestMain is a special function that runs once for the entire package.
// We use it to compile our gitego binary for the integration tests.
func TestMain(m *testing.M) {
	// Build the gitego binary into a temporary location.
	tempDir, err := os.MkdirTemp("", "gitego-integration-")
	if err != nil {
		panic("Failed to create temp dir for binary: " + err.Error())
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			panic("Failed to remove temp dir: " + err.Error())
		}
	}()

	gitegoBinary = filepath.Join(tempDir, "gitego.exe") // .exe is safe for non-Windows OSes

	buildCmd := exec.Command("go", "build", "-o", gitegoBinary, "..")
	if err := buildCmd.Run(); err != nil {
		panic("Failed to build gitego binary: " + err.Error())
	}

	// Run the actual tests.
	exitCode := m.Run()
	os.Exit(exitCode)
}

// setupTestEnvironment creates a clean temporary home directory for a test run.
func setupTestEnvironment(t *testing.T) (homeDir string) {
	t.Helper()

	homeDir, err := os.MkdirTemp("", "gitego-home-")
	if err != nil {
		t.Fatalf("Failed to create temp home dir: %v", err)
	}

	t.Setenv("HOME", homeDir)
	t.Setenv("USERPROFILE", homeDir) // for Windows

	return homeDir
}

func TestIntegration_FullWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode.")
	}

	homeDir := setupTestEnvironment(t)
	defer func() {
		if err := os.RemoveAll(homeDir); err != nil {
			t.Logf("Warning: Failed to remove test home directory (this is common on Windows): %v", err)
		}
	}()

	paths := setupIntegrationPaths(homeDir)

	runIntegrationAddProfile(t)
	runIntegrationUseProfile(t)
	validateIntegrationUseProfile(t, paths.gitConfig)
	runIntegrationAutoRule(t, homeDir)
	validateIntegrationAutoRule(t, paths.gitConfig)
	runIntegrationRemoveProfile(t)
	validateIntegrationRemoveProfile(t, paths.gitConfig, paths.gitegoConfig)
}

func setupIntegrationPaths(homeDir string) struct{ gitConfig, gitegoConfig string } {
	return struct{ gitConfig, gitegoConfig string }{
		gitConfig:    filepath.Join(homeDir, ".gitconfig"),
		gitegoConfig: filepath.Join(homeDir, ".gitego", "config.yaml"),
	}
}

func runIntegrationAddProfile(t *testing.T) {
	addCmd := exec.Command(gitegoBinary, "add", "work", "--name", "Test User", "--email", "test@work.com")
	addCmd.Env = os.Environ()

	if err := addCmd.Run(); err != nil {
		t.Fatalf("gitego add command failed: %v", err)
	}
}

func runIntegrationUseProfile(t *testing.T) {
	useCmd := exec.Command(gitegoBinary, "use", "work")
	useCmd.Env = os.Environ()

	if err := useCmd.Run(); err != nil {
		t.Fatalf("gitego use command failed: %v", err)
	}
}

func validateIntegrationUseProfile(t *testing.T, gitConfigPath string) {
	profileBytes, _ := os.ReadFile(filepath.Join(filepath.Dir(gitConfigPath), ".gitego", "profiles", "work.gitconfig"))
	profileContent := string(profileBytes)
	if !strings.Contains(profileContent, "name = \"Test User\"") || !strings.Contains(profileContent, "email = \"test@work.com\"") || !strings.Contains(profileContent, "profile = \"work\"") {
		t.Fatal("managed profile config was not updated correctly by 'use' command")
	}
}

func runIntegrationAutoRule(t *testing.T, homeDir string) {
	autoDir := filepath.Join(homeDir, "work-projects")
	if err := os.Mkdir(autoDir, 0755); err != nil {
		t.Fatalf("Failed to create auto directory: %v", err)
	}

	autoCmd := exec.Command(gitegoBinary, "auto", autoDir, "work")
	autoCmd.Env = os.Environ()

	if err := autoCmd.Run(); err != nil {
		t.Fatalf("gitego auto command failed: %v", err)
	}
}

func validateIntegrationAutoRule(t *testing.T, gitConfigPath string) {
	gitConfigBytes, _ := os.ReadFile(gitConfigPath)

	gitConfigContent := string(gitConfigBytes)
	if !strings.Contains(gitConfigContent, "includes.gitconfig") {
		t.Fatal(".gitconfig was not updated with the managed include")
	}
}

func runIntegrationRemoveProfile(t *testing.T) {
	rmCmd := exec.Command(gitegoBinary, "rm", "work", "--force")
	rmCmd.Env = os.Environ()

	if err := rmCmd.Run(); err != nil {
		t.Fatalf("gitego rm command failed: %v", err)
	}
}

func validateIntegrationRemoveProfile(t *testing.T, gitConfigPath, gitegoConfigPath string) {
	gitegoConfigBytes, _ := os.ReadFile(gitegoConfigPath)

	gitegoConfigContent := string(gitegoConfigBytes)
	if strings.Contains(gitegoConfigContent, "work:") {
		t.Fatal("Profile 'work' was not removed from gitego's config.yaml.")
	}

	gitConfigBytes, _ := os.ReadFile(gitConfigPath)

	gitConfigContent := string(gitConfigBytes)
	if strings.Contains(gitConfigContent, "[includeIf") {
		t.Fatal("The includeIf rule was not cleaned up from .gitconfig by 'rm' command.")
	}
}
