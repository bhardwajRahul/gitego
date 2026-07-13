// cmd/use.go
package cmd

import (
	"fmt"
	"runtime"

	"github.com/bgreenwell/git-ego/config"
	"github.com/bgreenwell/git-ego/utils"
	"github.com/spf13/cobra"
)

// useRunner holds the dependencies for the use command for mocking.
type useRunner struct {
	load             func() (*config.Config, error)
	save             func(*config.Config) error
	setGlobalGit     func(string, string) error
	unsetGlobalGit   func(string) error
	setGitCredential func(string, string) error
	getOS            func() string
	getToken         func(string) (string, error)
}

// run is the core logic for the use command.
func (u *useRunner) run(cmd *cobra.Command, args []string) error {
	profileName := args[0]
	if err := config.ValidateProfileName(profileName); err != nil {
		return err
	}

	cfg, err := u.load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	profile, exists := cfg.Profiles[profileName]
	if !exists {
		return fmt.Errorf("profile %q not found", profileName)
	}

	if err := applyProfileToGlobal(profile, u.setGlobalGit, u.unsetGlobalGit); err != nil {
		return fmt.Errorf("apply Git profile: %w", err)
	}

	// Action 2: Set this profile as the active one in gitego's config.
	cfg.ActiveProfile = profileName
	if err := u.save(cfg); err != nil {
		return fmt.Errorf("save active profile: %w", err)
	}

	// Action 3: If on macOS, also preemptively set the credential
	// in the keychain to prevent the osxkeychain helper from prompting.
	if u.getOS() == "darwin" {
		token, err := u.getToken(profileName)
		if err == nil && token != "" && profile.Username != "" {
			_ = u.setGitCredential(profile.Username, token)
		}
	}

	fmt.Printf("✓ Set active profile to '%s'.\n", profileName)
	return nil
}

func applyProfileToGlobal(profile *config.Profile, set func(string, string) error, unset func(string) error) error {
	for key, value := range map[string]string{"user.name": profile.Name, "user.email": profile.Email} {
		if err := set(key, value); err != nil {
			return fmt.Errorf("setting %s: %w", key, err)
		}
	}

	if profile.SigningKey != "" {
		if err := set("user.signingkey", profile.SigningKey); err != nil {
			return fmt.Errorf("setting user.signingkey: %w", err)
		}
		format := "openpgp"
		if config.IsSSHSigningKey(profile.SigningKey) {
			format = "ssh"
		}
		if err := set("gpg.format", format); err != nil {
			return fmt.Errorf("setting gpg.format: %w", err)
		}
	} else if unset != nil {
		if err := unset("user.signingkey"); err != nil {
			return err
		}
		if err := unset("gpg.format"); err != nil {
			return err
		}
	}

	if profile.SSHKey != "" {
		if err := set("core.sshCommand", config.SSHCommand(profile.SSHKey)); err != nil {
			return fmt.Errorf("setting core.sshCommand: %w", err)
		}
	} else if unset != nil {
		if err := unset("core.sshCommand"); err != nil {
			return err
		}
	}

	return nil
}

var useCmd = &cobra.Command{
	Use:   "use <profile_name>",
	Short: "Sets a profile as the active default for gitego.",
	Long: `Sets a profile as the active default. This profile will be used
for any repository that does not have a specific auto-switch rule.
This command updates your global .gitconfig, sets the active profile for the
credential helper, and preemptively updates the macOS Keychain.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		runner := &useRunner{
			load:             config.Load,
			save:             func(c *config.Config) error { return c.Save() },
			setGlobalGit:     utils.SetGlobalGitConfig,
			unsetGlobalGit:   utils.UnsetGlobalGitConfig,
			setGitCredential: config.SetGitCredential,
			getOS:            func() string { return runtime.GOOS },
			getToken:         config.GetToken,
		}
		return runner.run(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(useCmd)
}
