// cmd/credential.go
package cmd

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/bgreenwell/git-ego/config"
	"github.com/spf13/cobra"
)

// credentialRunner holds dependencies for the credential command for mocking.
type credentialRunner struct {
	loadConfig func() (*config.Config, error)
	getToken   func(string) (string, error)
	stdin      io.Reader
	stdout     io.Writer
}

// run is the core logic for the credential command.
func (r *credentialRunner) run(cmd *cobra.Command, args []string) {
	// Git invokes helpers as "get", "store", or "erase". Only get is
	// allowed to emit a credential response.
	if len(args) != 1 || args[0] != "get" {
		return
	}

	// Parse (rather than merely consume) Git's credential context. A global
	// helper must never offer a GitHub PAT to another HTTPS host.
	request := make(map[string]string)
	scanner := bufio.NewScanner(r.stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			break
		}
		key, value, ok := strings.Cut(line, "=")
		if ok {
			request[key] = value
		}
	}
	if scanner.Err() != nil || request["protocol"] != "https" || request["host"] == "" {
		return
	}

	cfg, err := r.loadConfig()
	if err != nil {
		return // If we can't load config, we can't do anything. Exit silently.
	}

	activeProfileName, _ := cfg.GetActiveProfileForCurrentDir()

	if activeProfileName == "" {
		return // No active profile, nothing to do.
	}

	profile, exists := cfg.Profiles[activeProfileName]
	if !exists || profile.Username == "" || !profile.SupportsCredentialHost(request["host"]) {
		return // Active profile doesn't exist or has no username for auth.
	}

	token, err := r.getToken(activeProfileName)
	if err != nil || token == "" {
		return // No PAT stored for this profile.
	}

	// Print the credentials to stdout in the format Git expects.
	if _, err := fmt.Fprintf(r.stdout, "username=%s\n", profile.Username); err != nil {
		log.Printf("Warning: Failed to write username: %v", err)
	}
	if _, err := fmt.Fprintf(r.stdout, "password=%s\n", token); err != nil {
		log.Printf("Warning: Failed to write password: %v", err)
	}
}

var credentialCmd = &cobra.Command{
	Use:    "credential",
	Short:  "Internal: A Git credential helper.",
	Hidden: true, // Hide this from the standard help command.
	Run: func(cmd *cobra.Command, args []string) {
		runner := &credentialRunner{
			loadConfig: config.LoadQuiet,
			getToken:   config.GetToken,
			stdin:      os.Stdin,
			stdout:     os.Stdout,
		}
		runner.run(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(credentialCmd)
}
