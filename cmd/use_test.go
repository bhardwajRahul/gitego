// cmd/use_test.go

package cmd

import (
	"fmt"
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

	var setCredentialCalls = make(map[string]string)

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
		setGitCredential: func(username, token string) error {
			setCredentialCalls[username] = token

			return nil
		},
		getOS:    func() string { return "linux" }, // Test non-darwin case first
		getToken: func(pn string) (string, error) { return "", fmt.Errorf("not found") },
	}

	// 3. Execute the command's logic
	args := []string{"personal"}
	runner.run(useCmd, args)

	// 4. Assertions
	if !savedConfig {
		t.Error("Expected config to be saved, but it wasn't.")
	}

	if mockCfg.ActiveProfile != "personal" {
		t.Errorf("Expected active profile to be 'personal', got '%s'", mockCfg.ActiveProfile)
	}

	if gitConfigCalls["user.name"] != "Test User" {
		t.Errorf("Expected user.name to be set to 'Test User', got '%s'", gitConfigCalls["user.name"])
	}

	if gitConfigCalls["user.email"] != "test@example.com" {
		t.Errorf("Expected user.email to be set to 'test@example.com', got '%s'", gitConfigCalls["user.email"])
	}

	if len(setCredentialCalls) > 0 {
		t.Error("Expected SetGitCredential not to be called on linux, but it was.")
	}
}

func TestUseCommand_macOS(t *testing.T) {
	// Test the macOS-specific path
	mockCfg := &config.Config{
		Profiles: map[string]*config.Profile{
			"work": {Name: "Mac User", Email: "mac@example.com", Username: "mac-user"},
		},
	}

	var setCredentialCalls = make(map[string]string)

	runner := &useRunner{
		load:         func() (*config.Config, error) { return mockCfg, nil },
		save:         func(c *config.Config) error { return nil },
		setGlobalGit: func(k, v string) error { return nil },
		getToken:     func(pn string) (string, error) { return "mac-token", nil },
		getOS:        func() string { return "darwin" },
		setGitCredential: func(username, token string) error {
			setCredentialCalls[username] = token

			return nil
		},
	}

	runner.run(useCmd, []string{"work"})

	if setCredentialCalls["mac-user"] != "mac-token" {
		t.Error("Expected SetGitCredential to be called with correct username and token on macOS")
	}
}
