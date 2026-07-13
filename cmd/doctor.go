package cmd

import (
	"fmt"

	"github.com/bgreenwell/git-ego/config"
	"github.com/bgreenwell/git-ego/utils"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Checks auto-rule configuration for drift.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		if doctorRepair {
			for _, rule := range cfg.AutoRules {
				profile, ok := cfg.Profiles[rule.Profile]
				if !ok {
					continue
				}
				if err := config.EnsureProfileGitconfig(rule.Profile, profile); err != nil {
					return err
				}
				if err := config.AddIncludeIf(rule.Profile, rule.Path); err != nil {
					return err
				}
			}
			cmd.Println("✓ Repaired generated profile includes and missing auto-rules where possible.")
		}
		problems := cfg.VerifyAutoRules()
		helpers, err := utils.GetGlobalGitConfigValues("credential.helper")
		if err != nil {
			problems = append(problems, fmt.Errorf("inspect credential helpers: %w", err))
		} else {
			found := false
			for _, helper := range helpers {
				if helper == "!git-ego credential" || helper == "!gitego credential" {
					found = true
					break
				}
			}
			if !found {
				problems = append(problems, fmt.Errorf("git-ego credential helper is not configured"))
			}
		}
		if len(problems) == 0 {
			cmd.Println("✓ gitego configuration is consistent.")
			return nil
		}
		for _, problem := range problems {
			cmd.Printf("✗ %v\n", problem)
		}
		return fmt.Errorf("found %d configuration problem(s)", len(problems))
	},
}

var doctorRepair bool

func init() {
	doctorCmd.Flags().BoolVar(&doctorRepair, "repair", false, "Regenerate profile includes and restore missing auto-rules")
	rootCmd.AddCommand(doctorCmd)
}
