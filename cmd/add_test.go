// cmd/add_test.go

package cmd

import (
	"testing"

	"github.com/bgreenwell/git-ego/config"
)

func TestAddCommand(t *testing.T) {
	// 1. Setup our in-memory state
	mockCfg := &config.Config{
		Profiles: make(map[string]*config.Profile),
	}

	var capturedTokenProfile, capturedToken string

	// 2. Create an instance of our command runner with mock functions
	a := &adder{
		load: func() (*config.Config, error) {
			return mockCfg, nil
		},
		save: func(c *config.Config) error {
			mockCfg = c // "Save" to our in-memory object

			return nil
		},
		setToken: func(profileName, token string) error {
			capturedTokenProfile = profileName
			capturedToken = token

			return nil
		},
	}

	// 3. Execute the command's logic
	args := []string{"work"}
	// Set the flag values programmatically for the test
	addName = "Test User"
	addEmail = "test@work.com"
	addPAT = "test_pat_123"

	a.run(addCmd, args)

	// 4. Assert the results
	if len(mockCfg.Profiles) != 1 {
		t.Fatalf("Expected 1 profile to be added, but found %d", len(mockCfg.Profiles))
	}

	profile, ok := mockCfg.Profiles["work"]
	if !ok {
		t.Fatal("Profile 'work' was not added to the config")
	}

	if profile.Name != "Test User" {
		t.Errorf("Expected profile name to be 'Test User', got '%s'", profile.Name)
	}

	if profile.Email != "test@work.com" {
		t.Errorf("Expected profile email to be 'test@work.com', got '%s'", profile.Email)
	}

	if capturedTokenProfile != "work" || capturedToken != "test_pat_123" {
		t.Error("SetToken was not called with the correct profile name and PAT")
	}
}
