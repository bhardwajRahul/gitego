// cmd/auto.go

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bgreenwell/git-ego/config"
	"github.com/spf13/cobra"
)

const (
	// exactArgs is the number of arguments for the auto command.
	exactArgs = 2
)

// autoRunner holds the dependencies for the auto command for mocking.
type autoRunner struct {
	load                   func() (*config.Config, error)
	save                   func(*config.Config) error
	ensureProfileGitconfig func(string, *config.Profile) error
	addIncludeIf           func(string, string) error
}

// run is the core logic for the auto command.
func (ar *autoRunner) run(cmd *cobra.Command, args []string) error {
	path := args[0]
	profileName := args[1]

	cfg, profile, err := ar.validateInputs(profileName)
	if err != nil {
		return err
	}

	cleanPath, err := ar.processPath(path)
	if err != nil {
		return fmt.Errorf("resolve path %q: %w", path, err)
	}

	if ar.ruleExists(cfg, cleanPath, profileName, path) {
		return nil
	}

	if err := ar.setupAutoRule(cfg, profileName, profile, cleanPath); err != nil {
		return err
	}

	fmt.Println("✓ Rule setup complete.")
	return nil
}

func (ar *autoRunner) validateInputs(profileName string) (*config.Config, *config.Profile, error) {
	if err := config.ValidateProfileName(profileName); err != nil {
		return nil, nil, err
	}
	cfg, err := ar.load()
	if err != nil {
		return nil, nil, fmt.Errorf("error loading configuration: %v", err)
	}

	profile, exists := cfg.Profiles[profileName]
	if !exists {
		return nil, nil, fmt.Errorf("profile '%s' not found in gitego", profileName)
	}

	return cfg, profile, nil
}

func (ar *autoRunner) processPath(path string) (string, error) {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		path = filepath.Join(home, path[2:])
	}

	return config.NormalizeAutoRulePath(path)
}

func (ar *autoRunner) ruleExists(cfg *config.Config, cleanPath, profileName, originalPath string) bool {
	for _, rule := range cfg.AutoRules {
		if rule.Path == cleanPath && rule.Profile == profileName {
			fmt.Printf("✓ Auto-switch rule for profile '%s' on path '%s' already exists.\n", profileName, originalPath)

			return true
		}
	}

	return false
}

func (ar *autoRunner) setupAutoRule(
	cfg *config.Config,
	profileName string,
	profile *config.Profile,
	cleanPath string,
) error {
	fmt.Printf("Setting up new auto-switch rule for profile '%s'...\n", profileName)

	if err := ar.ensureProfileGitconfig(profileName, profile); err != nil {
		return fmt.Errorf("error creating profile gitconfig: %v", err)
	}

	if err := ar.addIncludeIf(profileName, cleanPath); err != nil {
		return fmt.Errorf("error updating global .gitconfig: %v", err)
	}

	newRule := &config.AutoRule{
		Path:    cleanPath,
		Profile: profileName,
	}

	cfg.AutoRules = append(cfg.AutoRules, newRule)
	if err := ar.save(cfg); err != nil {
		return fmt.Errorf("warning: Git config updated, but failed to save rule to gitego config: %v", err)
	}

	return nil
}

var autoCmd = &cobra.Command{
	Use:   "auto <path> <profile_name>",
	Short: "Automatically switch profiles based on directory.",
	Long: `Configures your global .gitconfig to automatically use a specific
profile whenever you are working inside the given directory path.`,
	Args: cobra.ExactArgs(exactArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		runner := &autoRunner{
			load:                   config.Load,
			save:                   func(c *config.Config) error { return c.Save() },
			ensureProfileGitconfig: config.EnsureProfileGitconfig,
			addIncludeIf:           config.AddIncludeIf,
		}
		return runner.run(cmd, args)
	},
}

var autoListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured auto-switch rules.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		if len(cfg.AutoRules) == 0 {
			cmd.Println("No auto-switch rules configured.")
			return nil
		}
		for _, rule := range cfg.AutoRules {
			cmd.Printf("%s\t%s\n", rule.Path, rule.Profile)
		}
		return nil
	},
}

var autoRemoveCmd = &cobra.Command{
	Use:     "rm <path>",
	Aliases: []string{"remove"},
	Short:   "Remove one auto-switch rule.",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cleanPath, err := (&autoRunner{}).processPath(args[0])
		if err != nil {
			return fmt.Errorf("resolve path %q: %w", args[0], err)
		}
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		for i, rule := range cfg.AutoRules {
			if rule.Path != cleanPath {
				continue
			}
			if err := config.RemoveIncludeIfAt(rule.Profile, cleanPath); err != nil {
				return err
			}
			cfg.AutoRules = append(cfg.AutoRules[:i], cfg.AutoRules[i+1:]...)
			if err := cfg.Save(); err != nil {
				return err
			}
			cmd.Printf("✓ Removed auto-switch rule for %q.\n", args[0])
			return nil
		}
		return fmt.Errorf("no auto-switch rule for %q", args[0])
	},
}

func init() {
	autoCmd.AddCommand(autoListCmd, autoRemoveCmd)
	rootCmd.AddCommand(autoCmd)
}
