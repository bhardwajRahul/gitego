package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/bgreenwell/git-ego/config"
	"github.com/spf13/cobra"
)

// patCmd manages tokens without placing their values in shell history or the
// process argument list. Read the token from stdin, for example:
// printf %s "$GITHUB_TOKEN" | git-ego pat set work
var patCmd = &cobra.Command{
	Use:   "pat",
	Short: "Manage profile personal access tokens securely.",
}

var patSetCmd = &cobra.Command{
	Use:   "set <profile_name>",
	Short: "Store a PAT read from standard input.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.ValidateProfileName(args[0]); err != nil {
			return err
		}
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		if _, ok := cfg.Profiles[args[0]]; !ok {
			return fmt.Errorf("profile %q not found", args[0])
		}
		token, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("read PAT from stdin: %w", err)
		}
		if token := strings.TrimSpace(string(token)); token != "" {
			if err := config.SetToken(args[0], token); err != nil {
				return fmt.Errorf("store PAT: %w", err)
			}
		} else {
			return fmt.Errorf("PAT must not be empty")
		}
		cmd.Printf("✓ PAT stored for profile %q.\n", args[0])
		return nil
	},
}

var patDeleteCmd = &cobra.Command{
	Use:   "delete <profile_name>",
	Short: "Delete a profile PAT from the secure keyring.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.ValidateProfileName(args[0]); err != nil {
			return err
		}
		if err := config.DeleteToken(args[0]); err != nil {
			return fmt.Errorf("delete PAT: %w", err)
		}
		cmd.Printf("✓ PAT deleted for profile %q.\n", args[0])
		return nil
	},
}

func init() {
	patCmd.AddCommand(patSetCmd, patDeleteCmd)
	rootCmd.AddCommand(patCmd)
}
