// cmd/list.go

package cmd

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/bgreenwell/git-ego/config"
	"github.com/spf13/cobra"
)

const (
	// minwidth, tabwidth, padding, padchar, and flags for tabwriter.
	minwidth = 0
	tabwidth = 0
	padding  = 3
	padchar  = ' '
	flags    = 0
)

// listCmd represents the list command.
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists all saved user profiles and their attributes.",
	Long: `Reads the gitego configuration file and displays a table of all saved profiles, 
including their associated user name, email, and configured credentials (SSH, PAT).
The globally active profile is marked with an asterisk (*).`,
	Aliases: []string{"ls"}, // Users can run 'gitego ls' as a shortcut for 'gitego list'
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading configuration: %v\n", err)

			return
		}

		if len(cfg.Profiles) == 0 {
			fmt.Printf("No profiles found. Use '%s add <profile_name>' to create one.\n", binaryName)

			return
		}

		profileNames := make([]string, 0, len(cfg.Profiles))
		for name := range cfg.Profiles {
			profileNames = append(profileNames, name)
		}
		sort.Strings(profileNames)

		w := tabwriter.NewWriter(os.Stdout, minwidth, tabwidth, padding, padchar, flags)
		defer func() {
			if err := w.Flush(); err != nil {
				log.Printf("Warning: Failed to flush output: %v", err)
			}
		}()

		// New, more informative header
		if _, err := fmt.Fprintln(w, "ACTIVE\tPROFILE\tNAME\tEMAIL\tATTRIBUTES"); err != nil {
			log.Printf("Warning: Failed to write header: %v", err)
		}
		if _, err := fmt.Fprintln(w, "------\t-------\t----\t-----\t----------"); err != nil {
			log.Printf("Warning: Failed to write separator: %v", err)
		}

		for _, name := range profileNames {
			profile := cfg.Profiles[name]

			// 1. Check if this is the active profile
			activeMarker := " "
			if name == cfg.ActiveProfile {
				activeMarker = "*"
			}

			// 2. Check for associated credentials
			var attributes []string
			if profile.SSHKey != "" {
				attributes = append(attributes, "[SSH]")
			}
			// Check if a PAT exists in the keychain for this profile
			if token, err := config.GetToken(name); err == nil && token != "" {
				attributes = append(attributes, "[PAT]")
			}

			// 3. Print the enhanced row
			if _, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				activeMarker,
				name,
				profile.Name,
				profile.Email,
				strings.Join(attributes, " "),
			); err != nil {
				log.Printf("Warning: Failed to write profile row: %v", err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
