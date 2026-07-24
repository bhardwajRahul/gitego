package cmd

import (
	"bytes"
	"fmt"
	"os"

	"github.com/bgreenwell/git-ego/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var profileIOCmd = &cobra.Command{Use: "profiles", Short: "Import or export profile configuration."}
var profilesImportReplace bool

var profilesExportCmd = &cobra.Command{Use: "export <file>", Args: cobra.ExactArgs(1), Short: "Export profiles and auto-rules (never PATs or credential IDs).", RunE: func(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	portable := *cfg
	portable.Profiles = make(map[string]*config.Profile, len(cfg.Profiles))
	for name, p := range cfg.Profiles {
		copy := *p
		copy.CredentialID = ""
		portable.Profiles[name] = &copy
	}
	data, err := yaml.Marshal(&portable)
	if err != nil {
		return err
	}
	if err := os.WriteFile(args[0], data, 0600); err != nil {
		return err
	}
	cmd.Printf("✓ Exported profiles to %s.\n", args[0])
	return nil
}}

var profilesImportCmd = &cobra.Command{Use: "import <file>", Args: cobra.ExactArgs(1), Short: "Import validated profiles and rules (PATs are never imported).", RunE: func(cmd *cobra.Command, args []string) error {
	data, err := os.ReadFile(args[0])
	if err != nil {
		return err
	}
	var imported config.Config
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&imported); err != nil {
		return fmt.Errorf("parse import: %w", err)
	}
	if imported.Profiles == nil {
		imported.Profiles = map[string]*config.Profile{}
	}
	for name, p := range imported.Profiles {
		if p != nil && p.CredentialID != "" {
			return fmt.Errorf("import must not contain credential_id for profile %q", name)
		}
		id, err := config.NewCredentialID()
		if err != nil {
			return err
		}
		if p != nil {
			p.CredentialID = id
		}
	}
	if err := imported.Validate(); err != nil {
		return fmt.Errorf("validate import: %w", err)
	}
	return config.WithLock(func() error {
		current, err := config.Load()
		if err != nil {
			return err
		}
		if (len(current.Profiles) > 0 || len(current.AutoRules) > 0 || current.ActiveProfile != "") && !profilesImportReplace {
			return fmt.Errorf("configuration is not empty; use --replace to import")
		}
		if profilesImportReplace {
			type removed struct{ account, token string }
			var deleted []removed
			for name, p := range current.Profiles {
				account := p.CredentialID
				if account == "" {
					account = name
				}
				token, getErr := config.GetToken(account)
				if getErr != nil && !config.IsTokenNotFound(getErr) {
					return fmt.Errorf("read existing PAT for %q: %w", name, getErr)
				}
				if token == "" {
					continue
				}
				if err := config.DeleteToken(account); err != nil && !config.IsTokenNotFound(err) {
					for _, item := range deleted {
						_ = config.SetToken(item.account, item.token)
					}
					return fmt.Errorf("delete existing PAT for %q: %w", name, err)
				}
				deleted = append(deleted, removed{account, token})
			}
			if err := saveAndReconcile(&imported); err != nil {
				for _, item := range deleted {
					_ = config.SetToken(item.account, item.token)
				}
				return err
			}
		} else if err := saveAndReconcile(&imported); err != nil {
			return err
		}
		cmd.Printf("✓ Imported profiles from %s.\n", args[0])
		return nil
	})
}}

func init() {
	profilesImportCmd.Flags().BoolVar(&profilesImportReplace, "replace", false, "Replace non-empty configuration and detach existing PATs")
	profileIOCmd.AddCommand(profilesExportCmd, profilesImportCmd)
	rootCmd.AddCommand(profileIOCmd)
}
