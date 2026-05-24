// cmd/rm.go

package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bgreenwell/git-ego/config"
	"github.com/spf13/cobra"
)

var (
	forceFlag bool
)

// rmRunner holds the dependencies for the rm command for mocking.
type rmRunner struct {
	load             func() (*config.Config, error)
	save             func(*config.Config) error
	removeIncludeIf  func(string) error
	removeProfileCfg func(string) error
	deleteToken      func(string) error
}

// run is the core logic for the rm command.
func (r *rmRunner) run(cmd *cobra.Command, args []string) {
	profileName := args[0]

	cfg, err := r.load()
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)

		return
	}

	if _, exists := cfg.Profiles[profileName]; !exists {
		fmt.Printf("Error: Profile '%s' not found.\n", profileName)

		return
	}

	if !forceFlag {
		fmt.Printf("Are you sure you want to remove the profile '%s' and all its rules?\n", profileName)
		fmt.Print("This cannot be undone. [y/N]: ")

		reader := bufio.NewReader(os.Stdin)

		response, _ := reader.ReadString('\n')
		if strings.TrimSpace(strings.ToLower(response)) != "y" {
			fmt.Println("Removal cancelled.")

			return
		}
	}

	// 1. Remove the includeIf directive from the global .gitconfig.
	if err := r.removeIncludeIf(profileName); err != nil {
		fmt.Printf("Warning: Failed to remove rule from .gitconfig: %v\n", err)
	}

	// 2. Delete the profile-specific .gitconfig file.
	if err := r.removeProfileCfg(profileName); err != nil {
		fmt.Printf("Warning: Failed to remove profile config file: %v\n", err)
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

	if err := r.save(cfg); err != nil {
		fmt.Printf("Error saving configuration: %v\n", err)

		return
	}

	// 5. Remove the PAT from the OS keychain.
	_ = r.deleteToken(profileName)

	fmt.Printf("✓ Profile '%s' and all associated rules removed successfully.\n", profileName)
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
	Run: func(cmd *cobra.Command, args []string) {
		runner := &rmRunner{
			load:            config.Load,
			save:            func(c *config.Config) error { return c.Save() },
			removeIncludeIf: config.RemoveIncludeIf,
			deleteToken:     config.DeleteToken,
			removeProfileCfg: func(profileName string) error {
				home, err := os.UserHomeDir()
				if err != nil {
					return err
				}
				path := filepath.Join(home, ".gitego", "profiles", fmt.Sprintf("%s.gitconfig", profileName))

				return os.Remove(path)
			},
		}
		runner.run(cmd, args)
	},
}

func init() {
	rmCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Force removal without confirmation")
	rootCmd.AddCommand(rmCmd)
}
