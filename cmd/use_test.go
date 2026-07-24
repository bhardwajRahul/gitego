// cmd/use_test.go

package cmd

import (
	"testing"

	"github.com/bgreenwell/git-ego/config"
)

func TestUseCommand(t *testing.T) {
	// 1. Setup mock config and state trackers
	mockCfg := &config.Config{
		Profiles: map[string]*config.Profile{
			"personal": {Name: "Test User", Email: "test@example.com"},
		},
	}

	var savedConfig bool

	var gitConfigCalls = make(map[string]string)

	// 2. Create the test runner with mock functions
	runner := &useRunner{
		load: func() (*config.Config, error) {
			return mockCfg, nil
		},
		save: func(c *config.Config) error {
			savedConfig = true
			mockCfg = c // "Save" to our in-memory object

			return nil
		},
		setGlobalGit: func(key, value string) error {
			gitConfigCalls[key] = value

			return nil
		},
	}

	// 3. Execute the command's logic
	args := []string{"personal"}
	if err := runner.run(useCmd, args); err != nil {
		t.Fatal(err)
	}

	// 4. Assertions
	if !savedConfig {
		t.Error("Expected config to be saved, but it wasn't.")
	}

	if mockCfg.ActiveProfile != "personal" {
		t.Errorf("Expected active profile to be 'personal', got '%s'", mockCfg.ActiveProfile)
	}

	if len(gitConfigCalls) != 0 {
		t.Errorf("global use should not write top-level Git identity: %v", gitConfigCalls)
	}

}
