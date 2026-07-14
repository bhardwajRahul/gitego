// cmd/edit_test.go

package cmd

import (
	"log"
	"testing"

	"github.com/bgreenwell/git-ego/config"
	"github.com/spf13/cobra"
)

// setupEditTestConfig creates a mock config for edit command testing.
func setupEditTestConfig() *config.Config {
	return &config.Config{
		Profiles: map[string]*config.Profile{
			"work": {
				Name:     "Original Name",
				Email:    "original@example.com",
				Username: "original_user",
			},
		},
	}
}

func TestEditCommandSynchronizesActiveProfile(t *testing.T) {
	mockCfg := setupEditTestConfig()
	mockCfg.ActiveProfile = "work"

	command := &cobra.Command{}
	command.Flags().String("email", "", "")
	if err := command.Flags().Set("email", "updated@example.com"); err != nil {
		t.Fatal(err)
	}

	previousEmail := editEmail
	editEmail = "updated@example.com"
	t.Cleanup(func() { editEmail = previousEmail })

	var ensuredEmail string
	global := make(map[string]string)
	runner := &editor{
		load: func() (*config.Config, error) { return mockCfg, nil },
		save: func(*config.Config) error { return nil },
		ensureProfileGitconfig: func(_ string, profile *config.Profile) error {
			ensuredEmail = profile.Email
			return nil
		},
		setGlobalGit: func(key, value string) error {
			global[key] = value
			return nil
		},
		unsetGlobalGit: func(string) error { return nil },
	}

	if err := runner.run(command, []string{"work"}); err != nil {
		t.Fatal(err)
	}
	if ensuredEmail != "updated@example.com" {
		t.Fatalf("generated profile gitconfig used %q, want updated email", ensuredEmail)
	}
	if global["user.email"] != "updated@example.com" {
		t.Fatalf("active global profile used %q, want updated email", global["user.email"])
	}
}

func TestEditCommand(t *testing.T) {
	// Setup: Create an initial mock config
	mockCfg := setupEditTestConfig()

	// Create the test runner with mocked dependencies
	runner := &editor{
		load: func() (*config.Config, error) {
			cfgCopy := *mockCfg

			return &cfgCopy, nil
		},
		save: func(c *config.Config) error {
			*mockCfg = *c

			return nil
		},
	}

	// Execute the command's logic
	args := []string{"work"}

	cleanup := setEditCommandFlags("new-email@example.com")
	defer cleanup()

	if err := runner.run(editCmd, args); err != nil {
		t.Fatal(err)
	}

	// Assertions
	validateEditCommandResults(t, mockCfg)
}

// setEditCommandFlags sets the command flags for testing.
func setEditCommandFlags(email string) func() {
	if err := editCmd.Flags().Set("email", email); err != nil {
		log.Fatalf("Failed to set email flag: %v", err)
	}

	// Return cleanup function
	return func() {
		if err := editCmd.Flags().Set("email", ""); err != nil {
			log.Printf("Warning: Failed to reset email flag: %v", err)
		}
	}
}

// validateEditCommandResults validates the edit command results.
func validateEditCommandResults(t *testing.T, mockCfg *config.Config) {
	t.Helper()

	updatedProfile, ok := mockCfg.Profiles["work"]
	if !ok {
		t.Fatal("The 'work' profile was unexpectedly deleted.")
	}

	// This field should be changed
	if updatedProfile.Email != "new-email@example.com" {
		t.Errorf("Expected email to be updated to 'new-email@example.com', got '%s'", updatedProfile.Email)
	}

	// These fields should NOT have changed
	if updatedProfile.Name != "Original Name" {
		t.Errorf("Expected name to remain 'Original Name', but it was changed to '%s'", updatedProfile.Name)
	}

	if updatedProfile.Username != "original_user" {
		t.Errorf("Expected username to remain 'original_user', but it was changed to '%s'", updatedProfile.Username)
	}

}
