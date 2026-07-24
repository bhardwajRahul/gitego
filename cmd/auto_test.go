// cmd/auto_test.go

package cmd

import (
	"path/filepath"
	"testing"

	"github.com/bgreenwell/git-ego/config"
)

// setupAutoTestConfig creates a mock config for auto command testing.
func setupAutoTestConfig() *config.Config {
	return &config.Config{
		Profiles: map[string]*config.Profile{
			"work": {Name: "Work User", Email: "work@example.com"},
		},
		AutoRules: []*config.AutoRule{},
	}
}

func TestAutoCommand(t *testing.T) {
	// Setup
	mockCfg := setupAutoTestConfig()

	var savedConfig bool

	// Create the test runner with mocked dependencies
	runner := &autoRunner{
		load: func() (*config.Config, error) {
			cfgCopy := *mockCfg

			return &cfgCopy, nil
		},
		save: func(c *config.Config) error {
			savedConfig = true
			*mockCfg = *c

			return nil
		},
		ensureProfileGitconfig: func(profileName string, p *config.Profile) error {
			return nil
		},
		addIncludeIf: func(profileName, path string) error {
			return nil
		},
	}

	// Execute the command's logic
	testPath := filepath.Join("tmp", "work")
	args := []string{testPath, "work"}
	if err := runner.run(autoCmd, args); err != nil {
		t.Fatal(err)
	}

	// Assertions
	validateAutoRuleCreation(t, mockCfg, testPath)
	if !savedConfig {
		t.Fatal("expected authoritative config to be saved")
	}
}

// validateAutoRuleCreation validates that the auto rule was created correctly.
func validateAutoRuleCreation(t *testing.T, mockCfg *config.Config, testPath string) {
	t.Helper()

	if len(mockCfg.AutoRules) != 1 {
		t.Fatalf("Expected 1 auto-rule to be added, but found %d", len(mockCfg.AutoRules))
	}

	rule := mockCfg.AutoRules[0]
	if rule.Profile != "work" {
		t.Errorf("Expected rule to be for profile 'work', got '%s'", rule.Profile)
	}

	// Check that the path stored in the rule is absolute and has forward slashes
	absTestPath, _ := filepath.Abs(testPath)

	expectedPath := filepath.ToSlash(absTestPath) + "/"
	if rule.Path != expectedPath {
		t.Errorf("Expected rule path to be '%s', got '%s'", expectedPath, rule.Path)
	}
}

// validateAutoCommandEffects validates all side effects of the auto command.
