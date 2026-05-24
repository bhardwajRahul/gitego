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
func (ar *autoRunner) run(cmd *cobra.Command, args []string) {
	path := args[0]
	profileName := args[1]

	cfg, profile, err := ar.validateInputs(profileName)
	if err != nil {
		fmt.Println(err)

		return
	}

	cleanPath, err := ar.processPath(path)
	if err != nil {
		fmt.Printf("Error resolving path '%s': %v\n", path, err)

		return
	}

	if ar.ruleExists(cfg, cleanPath, profileName, path) {
		return
	}

	if err := ar.setupAutoRule(cfg, profileName, profile, cleanPath); err != nil {
		fmt.Println(err)

		return
	}

	fmt.Println("✓ Rule setup complete.")
}

func (ar *autoRunner) validateInputs(profileName string) (*config.Config, *config.Profile, error) {
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

	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	cleanPath := filepath.ToSlash(absPath)
	if !strings.HasSuffix(cleanPath, "/") {
		cleanPath += "/"
	}

	return cleanPath, nil
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
	Run: func(cmd *cobra.Command, args []string) {
		runner := &autoRunner{
			load:                   config.Load,
			save:                   func(c *config.Config) error { return c.Save() },
			ensureProfileGitconfig: config.EnsureProfileGitconfig,
			addIncludeIf:           config.AddIncludeIf,
		}
		runner.run(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(autoCmd)
}
