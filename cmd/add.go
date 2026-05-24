// cmd/add.go

package cmd

import (
	"errors"
	"fmt"
	"log"

	"github.com/bgreenwell/git-ego/config"
	"github.com/spf13/cobra"
)

var (
	addName       string
	addEmail      string
	addUsername   string
	addSSHKey     string
	addSigningKey string
	addPAT        string
)

// adder holds the dependencies for the add command, allowing them to be mocked for testing.
type adder struct {
	load     func() (*config.Config, error)
	save     func(*config.Config) error
	setToken func(string, string) error
}

// run is the core logic for the add command.
func (a *adder) run(cmd *cobra.Command, args []string) {
	profileName := args[0]

	cfg, err := a.load()
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)

		return
	}

	if _, exists := cfg.Profiles[profileName]; exists {
		fmt.Printf("Error: Profile '%s' already exists.\n", profileName)
		fmt.Printf("Use '%s edit %s' to modify it, or '%s rm %s' to remove it.\n", binaryName, profileName, binaryName, profileName)

		return
	}

	newProfile := &config.Profile{
		Name:       addName,
		Email:      addEmail,
		Username:   addUsername,
		SSHKey:     addSSHKey,
		SigningKey: addSigningKey,
	}

	cfg.Profiles[profileName] = newProfile

	if err := a.save(cfg); err != nil {
		fmt.Printf("Error saving configuration: %v\n", err)

		return
	}

	if addPAT != "" {
		if err := a.setToken(profileName, addPAT); err != nil {
			fmt.Printf("Warning: Profile saved, but failed to store PAT securely: %v\n", err)

			return
		}
	}

	fmt.Printf("✓ Profile '%s' added successfully.\n", profileName)
}

var addCmd = &cobra.Command{
	Use:   "add <profile_name>",
	Short: "Adds a new user profile to the gitego config.",
	Long: `Adds a new user profile, associating a profile name (e.g., "work")
with a specific Git user name and email address.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("requires exactly one argument: the profile name")
		}

		return nil
	},
	// The Run function is a wrapper around our testable run method.
	Run: func(cmd *cobra.Command, args []string) {
		a := &adder{
			load:     config.Load,
			save:     func(c *config.Config) error { return c.Save() },
			setToken: config.SetToken,
		}
		a.run(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(addCmd)

	addCmd.Flags().StringVarP(&addName, "name", "n", "", "The user.name for the profile")
	addCmd.Flags().StringVarP(&addEmail, "email", "e", "", "The user.email for the profile")
	addCmd.Flags().StringVar(&addUsername, "username", "", "Login username for the service (e.g., GitHub username)")
	addCmd.Flags().StringVar(&addSSHKey, "ssh-key", "", "Path to the SSH key for this profile (optional)")
	addCmd.Flags().StringVar(&addSigningKey, "signing-key", "", "GPG key ID or SSH key path for commit signing (optional)")
	addCmd.Flags().StringVar(&addPAT, "pat", "", "Personal Access Token for this profile (stored securely)")

	if err := addCmd.MarkFlagRequired("name"); err != nil {
		log.Fatalf("Failed to mark name flag as required: %v", err)
	}
	if err := addCmd.MarkFlagRequired("email"); err != nil {
		log.Fatalf("Failed to mark email flag as required: %v", err)
	}
}
