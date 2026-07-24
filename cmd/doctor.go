package cmd

import (
	"fmt"

	"github.com/bgreenwell/git-ego/config"
	"github.com/bgreenwell/git-ego/utils"
	"github.com/spf13/cobra"
)

var doctorRepair bool

var doctorCmd = &cobra.Command{Use: "doctor", Short: "Check safety markers, credentials, and generated Git configuration.", RunE: func(cmd *cobra.Command, _ []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if doctorRepair {
		return config.WithLock(func() error {
			cleanupFailures := 0
			for name, p := range cfg.Profiles {
				if p.CredentialID != "" {
					continue
				}
				id, err := config.NewCredentialID()
				if err != nil {
					return err
				}
				if token, getErr := config.GetToken(name); getErr == nil && token != "" {
					if err := config.SetToken(id, token); err != nil {
						return fmt.Errorf("migrate PAT for %q: %w", name, err)
					}
					if err := config.DeleteToken(name); err != nil && !config.IsTokenNotFound(err) {
						cleanupFailures++
						cmd.Printf("✗ migrated PAT for %q but could not remove its legacy key: %v\n", name, err)
					}
				} else if getErr != nil && !config.IsTokenNotFound(getErr) {
					return fmt.Errorf("read legacy PAT for %q: %w", name, getErr)
				}
				p.CredentialID = id
			}
			if err := saveAndReconcile(cfg); err != nil {
				return err
			}
			cmd.Println("✓ Repaired and migrated generated configuration.")
			if cleanupFailures > 0 {
				return fmt.Errorf("repair completed with %d legacy token cleanup failure(s)", cleanupFailures)
			}
			return nil
		})
	}
	problems := cfg.VerifyAutoRules()
	for name, p := range cfg.Profiles {
		if p.CredentialID == "" {
			problems = append(problems, fmt.Errorf("profile %q uses legacy credential storage", name))
		}
	}
	helpers, err := utils.GetGlobalGitConfigValues("credential.helper")
	if err != nil {
		problems = append(problems, err)
	}
	found := false
	for _, h := range helpers {
		if h == "!git-ego credential" || h == "!gitego credential" {
			found = true
		}
	}
	if !found {
		problems = append(problems, fmt.Errorf("git-ego credential helper is not configured"))
	}
	if len(problems) == 0 {
		cmd.Println("✓ gitego configuration is consistent.")
		return nil
	}
	for _, p := range problems {
		cmd.Printf("✗ %v\n", p)
	}
	return fmt.Errorf("found %d configuration problem(s); run '%s doctor --repair'", len(problems), binaryName)
}}

func init() {
	doctorCmd.Flags().BoolVar(&doctorRepair, "repair", false, "Migrate credentials and regenerate all managed Git files")
	rootCmd.AddCommand(doctorCmd)
}
