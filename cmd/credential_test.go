// cmd/credential_test.go

package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/bgreenwell/git-ego/config"
	"github.com/spf13/cobra"
)

func TestCredentialCommand(t *testing.T) {
	// 1. Setup mock config
	mockCfg := &config.Config{
		Profiles: map[string]*config.Profile{
			"work": {
				Name:     "Work User",
				Email:    "work@example.com",
				Username: "work-gh-user",
			},
		},
		ActiveProfile: "work", // Set "work" as the active profile
	}

	// 2. Create the test runner with mock dependencies
	runner := &credentialRunner{
		loadConfig: func() (*config.Config, error) {
			return mockCfg, nil
		},
		getToken: func(profileName string) (string, error) {
			if profileName == "work" {
				return "secret-work-token", nil
			}

			return "", nil
		},
		// Simulate Git providing some input, which we ignore
		stdin: strings.NewReader("protocol=https\nhost=github.com\n\n"),
	}

	// 3. Capture the stdout of the command
	var stdoutBuf bytes.Buffer

	runner.stdout = &stdoutBuf

	// 4. Execute the command's logic
	dummyCmd := &cobra.Command{}
	runner.run(dummyCmd, []string{})

	output := stdoutBuf.String()

	// 5. Assertions
	expectedUser := "username=work-gh-user"
	if !strings.Contains(output, expectedUser) {
		t.Errorf("Expected output to contain '%s', but it didn't.\nOutput:\n%s", expectedUser, output)
	}

	expectedPass := "password=secret-work-token"
	if !strings.Contains(output, expectedPass) {
		t.Errorf("Expected output to contain '%s', but it didn't.\nOutput:\n%s", expectedPass, output)
	}
}
