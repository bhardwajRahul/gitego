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

func init() { rootCmd.AddCommand(doctorCmd) }
