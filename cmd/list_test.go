// cmd/list_test.go

package cmd

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"testing"
	"text/tabwriter"

	"github.com/bgreenwell/git-ego/config"
	"github.com/spf13/cobra"
)

// listRunner holds the dependencies for the list command for mocking.
type listRunner struct {
	load     func() (*config.Config, error)
	getToken func(string) (string, error)
}

// run executes the core logic of the list command.
func (lr *listRunner) run(cmd *cobra.Command, args []string) {
	cfg, err := lr.load()
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)

		return
	}

	if len(cfg.Profiles) == 0 {
		fmt.Println("No profiles found. Use 'gitego add <profile_name>' to create one.")

		return
	}

	profileNames := make([]string, 0, len(cfg.Profiles))
	for name := range cfg.Profiles {
		profileNames = append(profileNames, name)
	}

	sort.Strings(profileNames)

	// In the test, we write to the command's output stream.
	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 3, ' ', 0)
	defer func() {
		if err := w.Flush(); err != nil {
			fmt.Printf("Warning: Failed to flush output: %v\n", err)
		}
	}()

	if _, err := fmt.Fprintln(w, "ACTIVE\tPROFILE\tNAME\tEMAIL\tATTRIBUTES"); err != nil {
		fmt.Printf("Warning: Failed to write header: %v\n", err)
	}
	if _, err := fmt.Fprintln(w, "------\t-------\t----\t-----\t----------"); err != nil {
		fmt.Printf("Warning: Failed to write separator: %v\n", err)
	}

	for _, name := range profileNames {
		profile := cfg.Profiles[name]

		activeMarker := " "
		if name == cfg.ActiveProfile {
			activeMarker = "*"
		}

		var attributes []string
		if profile.SSHKey != "" {
			attributes = append(attributes, "[SSH]")
		}

		// Use the mocked getToken function to check for a PAT.
		if token, err := lr.getToken(name); err == nil && token != "" {
			attributes = append(attributes, "[PAT]")
		}

		if _, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			activeMarker,
			name,
			profile.Name,
			profile.Email,
			strings.Join(attributes, " "),
		); err != nil {
			fmt.Printf("Warning: Failed to write profile row: %v\n", err)
		}
	}
}

// validateListHeaders checks if all expected headers are present in the output.
func validateListHeaders(t *testing.T, output string) {
	expectedHeaders := []string{"ACTIVE", "PROFILE", "NAME", "EMAIL", "ATTRIBUTES"}
	for _, header := range expectedHeaders {
		if !strings.Contains(output, header) {
			t.Errorf("Expected output to contain header '%s', but it was missing.\nOutput:\n%s", header, output)
		}
	}
}

// extractProfileLines finds and returns the personal and work profile lines from the output.
func extractProfileLines(output string) (personalLine, workLine string) {
	lines := strings.Split(strings.TrimSpace(output), "\n")

	for _, line := range lines {
		if strings.Contains(line, "personal") {
			personalLine = line
		}

		if strings.Contains(line, "work") {
			workLine = line
		}
	}

	return personalLine, workLine
}

// validatePersonalProfileLine checks if the personal profile line contains expected content.
func validatePersonalProfileLine(t *testing.T, personalLine string) {
	if personalLine == "" {
		t.Fatal("Output did not contain a line for the 'personal' profile.")
	}

	if !strings.Contains(personalLine, "*") {
		t.Errorf("Expected 'personal' profile line to be marked as active ('*'), but it wasn't.\nLine: %s", personalLine)
	}

	if !strings.Contains(personalLine, "Personal User") {
		t.Errorf("Expected 'personal' profile line to contain 'Personal User'.\nLine: %s", personalLine)
	}
}

// validateWorkProfileLine checks if the work profile line contains expected attributes.
func validateWorkProfileLine(t *testing.T, workLine string) {
	if workLine == "" {
		t.Fatal("Output did not contain a line for the 'work' profile.")
	}

	if !strings.Contains(workLine, "[SSH]") {
		t.Errorf("Expected 'work' profile line to contain '[SSH]' attribute.\nLine: %s", workLine)
	}

	if !strings.Contains(workLine, "[PAT]") {
		t.Errorf("Expected 'work' profile line to contain '[PAT]' attribute.\nLine: %s", workLine)
	}
}

// TestListCommand verifies the output of the 'list' command.
func TestListCommand(t *testing.T) {
	// 1. Setup: Create a mock config and dependencies
	mockCfg := &config.Config{
		Profiles: map[string]*config.Profile{
			"work": {
				Name:     "Work User",
				Email:    "work@example.com",
				SSHKey:   "~/.ssh/id_work",
				Username: "workuser", // A username is present
			},
			"personal": {
				Name:  "Personal User",
				Email: "personal@example.com",
			},
		},
		ActiveProfile: "personal",
	}

	lr := &listRunner{
		load: func() (*config.Config, error) {
			return mockCfg, nil
		},
		// Mock the keychain check. Pretend the "work" profile has a PAT.
		getToken: func(profileName string) (string, error) {
			if profileName == "work" {
				return "a-test-token", nil
			}

			return "", fmt.Errorf("no token found")
		},
	}

	// 2. Redirect command's stdout to capture the output
	var buf bytes.Buffer

	listCmd := &cobra.Command{}
	listCmd.SetOut(&buf)

	// 3. Execute the command's logic
	lr.run(listCmd, []string{})

	// 4. Read the captured output
	output := buf.String()

	// 5. Assert the output is correct with robust checks
	validateListHeaders(t, output)

	personalLine, workLine := extractProfileLines(output)
	validatePersonalProfileLine(t, personalLine)
	validateWorkProfileLine(t, workLine)
}
