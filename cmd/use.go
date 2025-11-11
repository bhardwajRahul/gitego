// cmd/use.go
package cmd

import (
	"fmt"
	"runtime"

	"github.com/bgreenwell/gitego/config"
	"github.com/bgreenwell/gitego/utils"
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
func (u *useRunner) run(cmd *cobra.Command, args []string) {
	profileName := args[0]

	cfg, err := u.load()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)

		return
	}

	profile, exists := cfg.Profiles[profileName]
	if !exists {
		fmt.Printf("Error: Profile '%s' not found.\n", profileName)

		return
	}

	// Action 1: Set the global git config for user name and email.
	if err := u.setGlobalGit("user.name", profile.Name); err != nil {
		fmt.Printf("Error setting git user.name: %v\n", err)

		return
	}

	if err := u.setGlobalGit("user.email", profile.Email); err != nil {
		fmt.Printf("Error setting git user.email: %v\n", err)

		return
	}

	if profile.SigningKey != "" {
		if err := u.setGlobalGit("user.signingkey", profile.SigningKey); err != nil {
			fmt.Printf("Error setting git user.signingkey: %v\n", err)

			return
		}
	} else if u.unsetGlobalGit != nil {
		_ = u.unsetGlobalGit("user.signingkey")
	}

	if profile.SSHKey != "" {
		sshCommand := fmt.Sprintf("ssh -i %s", profile.SSHKey)
		if err := u.setGlobalGit("core.sshCommand", sshCommand); err != nil {
			fmt.Printf("Error setting git core.sshCommand: %v\n", err)

			return
		}
	} else if u.unsetGlobalGit != nil {
		_ = u.unsetGlobalGit("core.sshCommand")
	}

	// Action 2: Set this profile as the active one in gitego's config.
	cfg.ActiveProfile = profileName
	if err := u.save(cfg); err != nil {
		fmt.Printf("Error saving active profile setting: %v\n", err)

		return
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
}

var useCmd = &cobra.Command{
	Use:   "use <profile_name>",
	Short: "Sets a profile as the active default for gitego.",
	Long: `Sets a profile as the active default. This profile will be used
for any repository that does not have a specific auto-switch rule.
This command updates your global .gitconfig, sets the active profile for the
credential helper, and preemptively updates the macOS Keychain.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runner := &useRunner{
			load:             config.Load,
			save:             func(c *config.Config) error { return c.Save() },
			setGlobalGit:     utils.SetGlobalGitConfig,
			unsetGlobalGit:   utils.UnsetGlobalGitConfig,
			setGitCredential: config.SetGitCredential,
			getOS:            func() string { return runtime.GOOS },
			getToken:         config.GetToken,
		}
		runner.run(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(useCmd)
}
