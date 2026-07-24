package cmd

import (
	"fmt"

	"github.com/bgreenwell/git-ego/config"
	"github.com/bgreenwell/git-ego/utils"
	"github.com/spf13/cobra"
)

type statusRunner struct {
	load         func() (*config.Config, error)
	getGitConfig func(string) (string, error)
	resolve      func(*config.Config) profileResolution
}

func (sr *statusRunner) run(cmd *cobra.Command, _ []string) error {
	cfg, err := sr.load()
	if err != nil {
		return fmt.Errorf("load gitego configuration: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid gitego configuration: %w", err)
	}
	resolver := sr.resolve
	if resolver == nil {
		resolver = resolveProfiles
	}
	r := resolver(cfg)
	cmd.Println("--- Git Identity Status ---")
	cmd.Printf("  Name:      %s\n", r.Name)
	cmd.Printf("  Email:     %s\n", r.Email)
	cmd.Printf("  Effective: %s\n", valueOrNone(r.Effective))
	cmd.Printf("  Origin:    %s\n", valueOrNone(r.Origin))
	if r.Expected != "" || r.ExpectationSource != "" {
		cmd.Printf("  Expected:  %s (%s)\n", valueOrNone(r.Expected), r.ExpectationSource)
	}
	if r.Legacy {
		cmd.Println("  Legacy:    validated compatibility fallback")
	}
	if !r.Consistent {
		cmd.Printf("  Consistent: no (%s)\n", r.Problem)
		return fmt.Errorf("git identity is inconsistent")
	}
	cmd.Println("  Consistent: yes")
	return nil
}

func valueOrNone(v string) string {
	if v == "" {
		return "(none)"
	}
	return v
}

var statusCmd = &cobra.Command{
	Use: "status", Short: "Display the effective profile, origin, expectation, and consistency.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return (&statusRunner{load: config.Load, getGitConfig: utils.GetEffectiveGitConfig, resolve: resolveProfiles}).run(cmd, args)
	},
}

func init() { rootCmd.AddCommand(statusCmd) }
