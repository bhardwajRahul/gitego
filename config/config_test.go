// config/config_test.go

package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestLoad_Success tests the successful loading and parsing of a valid config file.
func TestLoad_Success(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gitego-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Warning: Failed to remove temp directory (this is common on Windows): %v", err)
		}
	}()

	testConfigContent := `
profiles:
  work:
    name: Test User
    email: test@work.com
    username: testuser
    ssh_key: ~/.ssh/id_work
  personal:
    name: Test Personal
    email: test@personal.com
active_profile: personal
auto_rules:
  - path: /tmp/work/
    profile: work
`
	tempConfigFile := filepath.Join(tempDir, "config.yaml")

	if err := os.WriteFile(tempConfigFile, []byte(testConfigContent), 0644); err != nil {
		t.Fatalf("Failed to write temp config file: %v", err)
	}

	originalConfigPath := gitegoConfigPath
	gitegoConfigPath = tempConfigFile

	defer func() {
		gitegoConfigPath = originalConfigPath
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned an unexpected error: %v", err)
	}

	if len(cfg.Profiles) != 2 {
		t.Errorf("Expected 2 profiles, but got %d", len(cfg.Profiles))
	}

	if profile, ok := cfg.Profiles["work"]; !ok {
		t.Error("Expected 'work' profile to exist, but it doesn't")
	} else if profile.Name != "Test User" {
		t.Errorf("Expected work profile name to be 'Test User', got '%s'", profile.Name)
	}

	if cfg.ActiveProfile != "personal" {
		t.Errorf("Expected active profile to be 'personal', got '%s'", cfg.ActiveProfile)
	}

	if len(cfg.AutoRules) != 1 {
		t.Errorf("Expected 1 auto_rule, but got %d", len(cfg.AutoRules))
	}
}

// TestLoad_NonExistentFile tests the behavior when the config file does not exist.
func TestLoad_NonExistentFile(t *testing.T) {
	// Point to a config file in a non-existent directory.
	originalConfigPath := gitegoConfigPath
	gitegoConfigPath = "/tmp/non/existent/path/config.yaml"

	defer func() {
		gitegoConfigPath = originalConfigPath
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned an unexpected error for a non-existent file: %v", err)
	}

	if cfg == nil {
		t.Fatal("Load() returned a nil config for a non-existent file.")
	}

	// Expect a new, empty config struct.
	if len(cfg.Profiles) != 0 || len(cfg.AutoRules) != 0 || cfg.ActiveProfile != "" {
		t.Error("Expected an empty config struct when file does not exist, but got data.")
	}
}

// TestLoad_EmptyFile tests the behavior with an empty config file.
func TestLoad_EmptyFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gitego-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Warning: Failed to remove temp directory (this is common on Windows): %v", err)
		}
	}()

	// Create an empty file.
	tempConfigFile := filepath.Join(tempDir, "config.yaml")
	if _, err := os.Create(tempConfigFile); err != nil {
		t.Fatalf("Failed to create empty config file: %v", err)
	}

	originalConfigPath := gitegoConfigPath
	gitegoConfigPath = tempConfigFile

	defer func() {
		gitegoConfigPath = originalConfigPath
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned an unexpected error for an empty file: %v", err)
	}

	if cfg == nil {
		t.Fatal("Load() returned a nil config for an empty file.")
	}

	if len(cfg.Profiles) != 0 || len(cfg.AutoRules) != 0 || cfg.ActiveProfile != "" {
		t.Error("Expected an empty config struct for an empty file, but got data.")
	}
}

// TestLoad_MalformedYAML tests that an error is returned for an invalid YAML file.
func TestLoad_MalformedYAML(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gitego-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Warning: Failed to remove temp directory (this is common on Windows): %v", err)
		}
	}()

	// Malformed YAML with an unclosed quote.
	malformedContent := `profiles: { work: { name: "Test User }`

	tempConfigFile := filepath.Join(tempDir, "config.yaml")
	if err := os.WriteFile(tempConfigFile, []byte(malformedContent), 0644); err != nil {
		t.Fatalf("Failed to write malformed config file: %v", err)
	}

	originalConfigPath := gitegoConfigPath
	gitegoConfigPath = tempConfigFile

	defer func() {
		gitegoConfigPath = originalConfigPath
	}()

	_, err = Load()
	if err == nil {
		t.Error("Expected an error when loading malformed YAML, but got nil.")
	}
}

// TestRemoveIncludeIf_MultipleRulesWithSpaces verifies that the correct rule is removed
// from a .gitconfig file that has multiple gitego rules separated by blank lines.
func TestRemoveIncludeIf_MultipleRulesWithSpaces(t *testing.T) {
	// 1. Setup
	tempDir, err := os.MkdirTemp("", "gitego-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Warning: Failed to remove temp directory (this is common on Windows): %v", err)
		}
	}()

	// Override global paths to use our temp directory.
	originalGitConfigPath := gitConfigPath
	originalProfilesDir := profilesDir
	gitConfigPath = filepath.Join(tempDir, ".gitconfig")
	profilesDir = filepath.Join(tempDir, ".gitego", "profiles")

	defer func() {
		gitConfigPath = originalGitConfigPath
		profilesDir = originalProfilesDir
	}()

	// Create a mock .gitconfig with two rules and spacing, similar to the user's file.
	// IMPORTANT: Use forward slashes as this is what gitego writes.
	initialGitconfigContent := `[user]
	email = test@example.com
	name = Test User

# gitego auto-switch rule
[includeIf "gitdir:C:/test/personal/"]
    path = ` + filepath.ToSlash(filepath.Join(profilesDir, "personal.gitconfig")) + `

# gitego auto-switch rule
[includeIf "gitdir:C:/test/work/"]
    path = ` + filepath.ToSlash(filepath.Join(profilesDir, "work.gitconfig")) + `
`
	if err := os.WriteFile(gitConfigPath, []byte(initialGitconfigContent), 0644); err != nil {
		t.Fatalf("Failed to write initial .gitconfig: %v", err)
	}

	// 2. Execute the function to remove the "work" profile's rule.
	err = RemoveIncludeIf("work")
	if err != nil {
		t.Fatalf("RemoveIncludeIf returned an unexpected error: %v", err)
	}

	// 3. Assert the result
	finalContent, err := os.ReadFile(gitConfigPath)
	if err != nil {
		t.Fatalf("Failed to read final .gitconfig: %v", err)
	}

	finalContentStr := string(finalContent)

	// The "work" profile's path should be gone.
	if strings.Contains(finalContentStr, "work.gitconfig") {
		t.Errorf("Expected 'work.gitconfig' rule to be removed, but it still exists.\nContent:\n%s", finalContentStr)
	}

	// The "personal" profile's path should still be present.
	if !strings.Contains(finalContentStr, "personal.gitconfig") {
		t.Errorf("Expected 'personal.gitconfig' rule to remain, but it was removed.\nContent:\n%s", finalContentStr)
	}

	// The user section should be untouched.
	if !strings.Contains(finalContentStr, "[user]") {
		t.Error("The [user] section was unexpectedly removed.")
	}
}

// TestEnsureProfileGitconfig_WithSigningKey tests that a profile with a signing key
// generates the correct .gitconfig file content.
func TestEnsureProfileGitconfig_WithSigningKey(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gitego-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Warning: Failed to remove temp directory (this is common on Windows): %v", err)
		}
	}()

	// Override the global profilesDir to use our temp directory.
	originalProfilesDir := profilesDir
	profilesDir = tempDir

	defer func() {
		profilesDir = originalProfilesDir
	}()

	// Create a profile with a signing key.
	profile := &Profile{
		Name:       "Test User",
		Email:      "test@example.com",
		SigningKey: "ABCD1234",
	}

	// Call the function.
	err = EnsureProfileGitconfig("test-profile", profile)
	if err != nil {
		t.Fatalf("EnsureProfileGitconfig returned an error: %v", err)
	}

	// Read the generated file.
	generatedFile := filepath.Join(tempDir, "test-profile.gitconfig")
	content, err := os.ReadFile(generatedFile)
	if err != nil {
		t.Fatalf("Failed to read generated gitconfig file: %v", err)
	}

	contentStr := string(content)

	// Assert that the signing key is present.
	if !strings.Contains(contentStr, "signingkey = \"ABCD1234\"") {
		t.Errorf("Expected 'signingkey = ABCD1234' in gitconfig, but got:\n%s", contentStr)
	}

	// Assert that the user section is present.
	if !strings.Contains(contentStr, "name = \"Test User\"") {
		t.Errorf("Expected 'name = Test User' in gitconfig, but got:\n%s", contentStr)
	}

	if !strings.Contains(contentStr, "email = \"test@example.com\"") {
		t.Errorf("Expected 'email = test@example.com' in gitconfig, but got:\n%s", contentStr)
	}
}

// TestEnsureProfileGitconfig_WithoutSigningKey tests that a profile without a signing key
// does not include the signingkey line.
func TestEnsureProfileGitconfig_WithoutSigningKey(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gitego-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Warning: Failed to remove temp directory (this is common on Windows): %v", err)
		}
	}()

	// Override the global profilesDir to use our temp directory.
	originalProfilesDir := profilesDir
	profilesDir = tempDir

	defer func() {
		profilesDir = originalProfilesDir
	}()

	// Create a profile without a signing key.
	profile := &Profile{
		Name:  "Test User",
		Email: "test@example.com",
	}

	// Call the function.
	err = EnsureProfileGitconfig("test-profile", profile)
	if err != nil {
		t.Fatalf("EnsureProfileGitconfig returned an error: %v", err)
	}

	// Read the generated file.
	generatedFile := filepath.Join(tempDir, "test-profile.gitconfig")
	content, err := os.ReadFile(generatedFile)
	if err != nil {
		t.Fatalf("Failed to read generated gitconfig file: %v", err)
	}

	contentStr := string(content)

	// Assert that the signing key is NOT present.
	if strings.Contains(contentStr, "signingkey") {
		t.Errorf("Did not expect 'signingkey' in gitconfig when SigningKey is empty, but got:\n%s", contentStr)
	}

	// Assert that the user section is still present.
	if !strings.Contains(contentStr, "name = \"Test User\"") {
		t.Errorf("Expected 'name = Test User' in gitconfig, but got:\n%s", contentStr)
	}
}

func TestRemoveIncludeIfPreservesFollowingSection(t *testing.T) {
	tempDir := t.TempDir()
	originalGitConfigPath, originalProfilesDir := gitConfigPath, profilesDir
	gitConfigPath = filepath.Join(tempDir, ".gitconfig")
	profilesDir = filepath.Join(tempDir, ".gitego", "profiles")
	t.Cleanup(func() { gitConfigPath, profilesDir = originalGitConfigPath, originalProfilesDir })

	profilePath, err := ProfileGitconfigPath("work")
	if err != nil {
		t.Fatal(err)
	}
	content := "# gitego auto-switch rule\n[includeIf \"gitdir:/tmp/work/\"]\n    path = \"" + filepath.ToSlash(profilePath) + "\"\n[user]\n    name = \"Unrelated User\"\n"
	if err := os.WriteFile(gitConfigPath, []byte(content), filePermissions); err != nil {
		t.Fatal(err)
	}
	if err := RemoveIncludeIf("work"); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(gitConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "[user]\n    name = \"Unrelated User\"") {
		t.Fatalf("following section was corrupted:\n%s", got)
	}
}

func TestRemoveIncludeIfPreservesAdjacentIncludeRule(t *testing.T) {
	tempDir := t.TempDir()
	originalGitConfigPath, originalProfilesDir := gitConfigPath, profilesDir
	gitConfigPath = filepath.Join(tempDir, ".gitconfig")
	profilesDir = filepath.Join(tempDir, ".gitego", "profiles")
	t.Cleanup(func() { gitConfigPath, profilesDir = originalGitConfigPath, originalProfilesDir })
	workPath, _ := ProfileGitconfigPath("work")
	personalPath, _ := ProfileGitconfigPath("personal")
	content := "[includeIf \"gitdir:/tmp/work/\"]\n path = \"" + filepath.ToSlash(workPath) + "\"\n[includeIf \"gitdir:/tmp/personal/\"]\n path = \"" + filepath.ToSlash(personalPath) + "\"\n"
	if err := os.WriteFile(gitConfigPath, []byte(content), filePermissions); err != nil {
		t.Fatal(err)
	}
	if err := RemoveIncludeIf("work"); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(gitConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "gitdir:/tmp/personal/") || !strings.Contains(string(got), filepath.ToSlash(personalPath)) {
		t.Fatalf("adjacent include rule was removed:\n%s", got)
	}
}

func TestEnsureProfileGitconfigEscapesAndRejectsUnsafeValues(t *testing.T) {
	tempDir := t.TempDir()
	originalProfilesDir := profilesDir
	profilesDir = tempDir
	t.Cleanup(func() { profilesDir = originalProfilesDir })

	profile := &Profile{Name: "A \"quoted\" name", Email: "user@example.com", SSHKey: "/tmp/My Key"}
	if err := EnsureProfileGitconfig("work", profile); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(filepath.Join(tempDir, "work.gitconfig"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), `name = "A \"quoted\" name"`) || !strings.Contains(string(content), `sshCommand = "ssh -i '/tmp/My Key'"`) {
		t.Fatalf("values were not safely quoted:\n%s", content)
	}
	if err := EnsureProfileGitconfig("work", &Profile{Name: "bad\n[include]", Email: "user@example.com"}); err == nil {
		t.Fatal("expected control character to be rejected")
	}
}

func TestEnsureProfileGitconfigSSHSigningKeySetsFormat(t *testing.T) {
	tempDir := t.TempDir()
	originalProfilesDir := profilesDir
	profilesDir = tempDir
	t.Cleanup(func() { profilesDir = originalProfilesDir })
	if err := EnsureProfileGitconfig("work", &Profile{Name: "User", Email: "user@example.com", SigningKey: "~/.ssh/id_signing.pub"}); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(filepath.Join(tempDir, "work.gitconfig"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "[gpg]\n    format = ssh") {
		t.Fatalf("missing SSH signing format:\n%s", content)
	}
}

func TestAddIncludeIfAllowsMultiplePathsForOneProfile(t *testing.T) {
	tempDir := t.TempDir()
	originalGitConfigPath, originalProfilesDir := gitConfigPath, profilesDir
	gitConfigPath = filepath.Join(tempDir, ".gitconfig")
	profilesDir = filepath.Join(tempDir, ".gitego", "profiles")
	t.Cleanup(func() { gitConfigPath, profilesDir = originalGitConfigPath, originalProfilesDir })

	if err := AddIncludeIf("work", "/tmp/first/"); err != nil {
		t.Fatal(err)
	}
	if err := AddIncludeIf("work", "/tmp/second/"); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(gitConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(content), "[includeIf") != 2 || !strings.Contains(string(content), "gitdir:/tmp/second/") {
		t.Fatalf("expected two distinct rules:\n%s", content)
	}
}
