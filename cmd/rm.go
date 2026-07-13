// cmd/rm.go

package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/bgreenwell/git-ego/config"
	"github.com/bgreenwell/git-ego/utils"
	"github.com/spf13/cobra"
)

var (
	forceFlag bool
)

// rmRunner holds the dependencies for the rm command for mocking.
type rmRunner struct {
	load                func() (*config.Config, error)
	save                func(*config.Config) error
	removeIncludeIf     func(string) error
	removeProfileCfg    func(string) error
	deleteToken         func(string) error
	deleteGitCredential func(string) error
	unsetGlobalGit      func(string) error
}

// run is the core logic for the rm command.
func (r *rmRunner) run(cmd *cobra.Command, args []string) error {
	profileName := args[0]
	if err := config.ValidateProfileName(profileName); err != nil {
		return err
	}

	cfg, err := r.load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	profile, exists := cfg.Profiles[profileName]
	if !exists {
		return fmt.Errorf("profile %q not found", profileName)
	}

	if !forceFlag {
		fmt.Printf("Are you sure you want to remove the profile '%s' and all its rules?\n", profileName)
		fmt.Print("This cannot be undone. [y/N]: ")

		reader := bufio.NewReader(os.Stdin)

		response, _ := reader.ReadString('\n')
		if strings.TrimSpace(strings.ToLower(response)) != "y" {
			return nil
		}
	}

	// 1. Remove the includeIf directive from the global .gitconfig.
	if err := r.removeIncludeIf(profileName); err != nil {
		return fmt.Errorf("remove Git include rule: %w", err)
	}

	// 2. Delete the profile-specific .gitconfig file.
	if err := r.removeProfileCfg(profileName); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove profile config: %w", err)
	}

	// 3. Remove any auto-rules from gitego's config that use this profile.
	var keptRules []*config.AutoRule

	for _, rule := range cfg.AutoRules {
		if rule.Profile != profileName {
			keptRules = append(keptRules, rule)
		}
	}

	cfg.AutoRules = keptRules

	// 4. Delete the profile itself.
	delete(cfg.Profiles, profileName)
	wasActive := cfg.ActiveProfile == profileName
	if wasActive {
		cfg.ActiveProfile = ""
	}

	if err := r.save(cfg); err != nil {
		return fmt.Errorf("save configuration: %w", err)
	}

	// 5. Remove the PAT from the OS keychain.
	// A profile may not have a token, and some headless keyring backends do not
	// support deletion. The profile/config removal is still authoritative.
	_ = r.deleteToken(profileName)
	if wasActive {
		if r.unsetGlobalGit != nil {
			for _, key := range []string{"user.name", "user.email", "user.signingkey", "gpg.format", "core.sshCommand"} {
				if err := r.unsetGlobalGit(key); err != nil {
					return fmt.Errorf("unset %s: %w", key, err)
				}
			}
		}
		if r.deleteGitCredential != nil {
			if err := r.deleteGitCredential(profile.Username); err != nil {
				return fmt.Errorf("remove active Git credential: %w", err)
			}
		}
	}

	fmt.Printf("✓ Profile '%s' and all associated rules removed successfully.\n", profileName)
	return nil
}

// rmCmd represents the rm command.
var rmCmd = &cobra.Command{
	Use:   "rm <profile_name>",
	Short: "Removes a saved user profile and all associated rules.",
	Long: `Removes a profile, its associated credentials, any auto-switch 
	rules that use it from the gitego config, and cleans up corresponding 
	rules from your global .gitconfig file.`,
	Aliases: []string{"remove"},
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		runner := &rmRunner{
			load:                config.Load,
			save:                func(c *config.Config) error { return c.Save() },
			removeIncludeIf:     config.RemoveIncludeIf,
			deleteToken:         config.DeleteToken,
			deleteGitCredential: config.DeleteGitCredential,
			unsetGlobalGit:      utils.UnsetGlobalGitConfig,
			removeProfileCfg: func(profileName string) error {
				path, err := config.ProfileGitconfigPath(profileName)
				if err != nil {
					return err
				}
				return os.Remove(path)
			},
		}
		return runner.run(cmd, args)
	},
}

func init() {
	rmCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Force removal without confirmation")
	rootCmd.AddCommand(rmCmd)
}
