// cmd/edit.go

package cmd

import (
	"fmt"

	"github.com/bgreenwell/git-ego/config"
	"github.com/bgreenwell/git-ego/utils"
	"github.com/spf13/cobra"
)

var (
	// These variables will hold the values from the flags for the 'edit' command.
	editName       string
	editEmail      string
	editUsername   string
	editSSHKey     string
	editSigningKey string
	editHosts      []string
)

// editor holds the dependencies for the edit command for mocking.
type editor struct {
	load                   func() (*config.Config, error)
	save                   func(*config.Config) error
	ensureProfileGitconfig func(string, *config.Profile) error
	setGlobalGit           func(string, string) error
	unsetGlobalGit         func(string) error
}

// run is the core logic for the edit command.
func (e *editor) run(cmd *cobra.Command, args []string) error {
	profileName := args[0]
	if err := config.ValidateProfileName(profileName); err != nil {
		return err
	}

	cfg, err := e.load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	profile, exists := cfg.Profiles[profileName]
	if !exists {
		return fmt.Errorf("profile %q not found", profileName)
	}

	// Update fields only if the corresponding flag was set by the user.
	if cmd.Flags().Changed("name") {
		profile.Name = editName
	}

	if cmd.Flags().Changed("email") {
		profile.Email = editEmail
	}

	if cmd.Flags().Changed("username") {
		profile.Username = editUsername
	}

	if cmd.Flags().Changed("ssh-key") {
		profile.SSHKey = editSSHKey
	}

	if cmd.Flags().Changed("signing-key") {
		profile.SigningKey = editSigningKey
	}
	if cmd.Flags().Changed("host") {
		profile.Hosts = editHosts
	}

	// Keep the generated include file in sync before advertising the update in
	// config.yaml. Auto-switching otherwise continues using stale values.
	if e.ensureProfileGitconfig != nil {
		if err := e.ensureProfileGitconfig(profileName, profile); err != nil {
			return fmt.Errorf("update profile gitconfig: %w", err)
		}
	}
	if cfg.ActiveProfile == profileName {
		if err := applyProfileToGlobal(profile, e.setGlobalGit, e.unsetGlobalGit); err != nil {
			return fmt.Errorf("apply updated active profile: %w", err)
		}
	}

	// Save the updated configuration.
	if err := e.save(cfg); err != nil {
		return fmt.Errorf("save configuration: %w", err)
	}

	fmt.Printf("✓ Profile '%s' updated successfully.\n", profileName)
	return nil
}

// editCmd represents the edit command.
var editCmd = &cobra.Command{
	Use:   "edit <profile_name>",
	Short: "Edits an existing user profile.",
	Long: `Edits an existing user profile. You can update the user name, email,
username, SSH key, or Personal Access Token (PAT).
Only the flags you provide will be updated.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		e := &editor{
			load:                   config.Load,
			save:                   func(c *config.Config) error { return c.Save() },
			ensureProfileGitconfig: config.EnsureProfileGitconfig,
			setGlobalGit:           utils.SetGlobalGitConfig,
			unsetGlobalGit:         utils.UnsetGlobalGitConfig,
		}
		return e.run(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(editCmd)

	// Define the flags for the 'edit' command.
	editCmd.Flags().StringVarP(&editName, "name", "n", "", "The new user.name for the profile")
	editCmd.Flags().StringVarP(&editEmail, "email", "e", "", "The new user.email for the profile")
	editCmd.Flags().StringVar(&editUsername, "username", "", "The new login username for the service")
	editCmd.Flags().StringVar(&editSSHKey, "ssh-key", "", "The new path to the SSH key for this profile")
	editCmd.Flags().StringVar(&editSigningKey, "signing-key", "", "The new GPG key ID or SSH key path for commit signing")
	editCmd.Flags().StringSliceVar(&editHosts, "host", nil, "HTTPS hosts this profile may authenticate to (repeat or comma-separate)")
}
