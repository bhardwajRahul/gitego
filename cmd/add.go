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
	addHosts      []string
)

// adder holds the dependencies for the add command, allowing them to be mocked for testing.
type adder struct {
	load func() (*config.Config, error)
	save func(*config.Config) error
}

// run is the core logic for the add command.
func (a *adder) run(cmd *cobra.Command, args []string) error {
	profileName := args[0]
	if err := config.ValidateProfileName(profileName); err != nil {
		return err
	}

	cfg, err := a.load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	if _, exists := cfg.Profiles[profileName]; exists {
		return fmt.Errorf("profile %q already exists; use '%s edit %s' or '%s rm %s'", profileName, binaryName, profileName, binaryName, profileName)
	}

	newProfile := &config.Profile{
		Name:       addName,
		Email:      addEmail,
		Username:   addUsername,
		SSHKey:     addSSHKey,
		SigningKey: addSigningKey,
		Hosts:      addHosts,
	}

	cfg.Profiles[profileName] = newProfile

	if err := a.save(cfg); err != nil {
		return fmt.Errorf("save configuration: %w", err)
	}

	fmt.Printf("✓ Profile '%s' added successfully.\n", profileName)
	return nil
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
	RunE: func(cmd *cobra.Command, args []string) error {
		a := &adder{
			load: config.Load,
			save: func(c *config.Config) error { return c.Save() },
		}
		return a.run(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(addCmd)

	addCmd.Flags().StringVarP(&addName, "name", "n", "", "The user.name for the profile")
	addCmd.Flags().StringVarP(&addEmail, "email", "e", "", "The user.email for the profile")
	addCmd.Flags().StringVar(&addUsername, "username", "", "Login username for the service (e.g., GitHub username)")
	addCmd.Flags().StringVar(&addSSHKey, "ssh-key", "", "Path to the SSH key for this profile (optional)")
	addCmd.Flags().StringVar(&addSigningKey, "signing-key", "", "GPG key ID or SSH key path for commit signing (optional)")
	addCmd.Flags().StringSliceVar(&addHosts, "host", nil, "HTTPS host this profile may authenticate to (repeat or comma-separate; defaults to github.com)")

	if err := addCmd.MarkFlagRequired("name"); err != nil {
		log.Fatalf("Failed to mark name flag as required: %v", err)
	}
	if err := addCmd.MarkFlagRequired("email"); err != nil {
		log.Fatalf("Failed to mark email flag as required: %v", err)
	}
}
