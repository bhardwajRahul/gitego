package cmd

import (
	"fmt"
	"os"

	"github.com/bgreenwell/git-ego/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var profileIOCmd = &cobra.Command{Use: "profiles", Short: "Import or export profile configuration."}

var profilesExportCmd = &cobra.Command{
	Use: "export <file>", Args: cobra.ExactArgs(1), Short: "Export profiles and auto-rules (never PATs).",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		data, err := yaml.Marshal(cfg)
		if err != nil {
			return err
		}
		if err := os.WriteFile(args[0], data, 0600); err != nil {
			return err
		}
		cmd.Printf("✓ Exported profiles to %s.\n", args[0])
		return nil
	},
}
var profilesImportCmd = &cobra.Command{
	Use: "import <file>", Args: cobra.ExactArgs(1), Short: "Import profiles and auto-rules (PATs are never imported).",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := os.ReadFile(args[0])
		if err != nil {
			return err
		}
		var imported config.Config
		if err := yaml.Unmarshal(data, &imported); err != nil {
			return fmt.Errorf("parse import: %w", err)
		}
		if imported.Profiles == nil {
			imported.Profiles = map[string]*config.Profile{}
		}
		for name := range imported.Profiles {
			if err := config.ValidateProfileName(name); err != nil {
				return err
			}
		}
		if err := imported.Save(); err != nil {
			return err
		}
		cmd.Printf("✓ Imported profiles from %s.\n", args[0])
		return nil
	},
}

func init() {
	profileIOCmd.AddCommand(profilesExportCmd, profilesImportCmd)
	rootCmd.AddCommand(profileIOCmd)
}
