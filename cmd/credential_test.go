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
	runner.run(dummyCmd, []string{"get"})

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

func TestCredentialCommandScopesHostAndOperation(t *testing.T) {
	runner := &credentialRunner{
		loadConfig: func() (*config.Config, error) {
			return &config.Config{Profiles: map[string]*config.Profile{
				"work": {Name: "Work User", Email: "work@example.com", Username: "work-user"},
			}, ActiveProfile: "work"}, nil
		},
		getToken: func(string) (string, error) { return "secret", nil },
	}

	for name, tc := range map[string]struct {
		args  []string
		stdin string
	}{
		"other host": {[]string{"get"}, "protocol=https\nhost=gitlab.com\n\n"},
		"erase":      {[]string{"erase"}, "protocol=https\nhost=github.com\n\n"},
		"store":      {[]string{"store"}, "protocol=https\nhost=github.com\n\n"},
	} {
		t.Run(name, func(t *testing.T) {
			var output bytes.Buffer
			runner.stdin = strings.NewReader(tc.stdin)
			runner.stdout = &output
			runner.run(&cobra.Command{}, tc.args)
			if output.Len() != 0 {
				t.Fatalf("credential helper leaked a credential: %q", output.String())
			}
		})
	}
}

func TestCredentialCommandUsesConfiguredHosts(t *testing.T) {
	runner := &credentialRunner{
		loadConfig: func() (*config.Config, error) {
			return &config.Config{Profiles: map[string]*config.Profile{
				"work": {Name: "Work User", Email: "work@example.com", Username: "work-user", Hosts: []string{"gitlab.example.com", "github.example.com:8443"}},
			}, ActiveProfile: "work"}, nil
		},
		getToken: func(string) (string, error) { return "secret", nil },
	}

	for name, tc := range map[string]struct {
		host string
		want bool
	}{
		"configured host":        {host: "gitlab.example.com", want: true},
		"configured host case":   {host: "GITLAB.EXAMPLE.COM", want: true},
		"configured host port":   {host: "github.example.com:8443", want: true},
		"default GitHub blocked": {host: "github.com", want: false},
	} {
		t.Run(name, func(t *testing.T) {
			var output bytes.Buffer
			runner.stdin = strings.NewReader("protocol=https\nhost=" + tc.host + "\n\n")
			runner.stdout = &output
			runner.run(&cobra.Command{}, []string{"get"})

			if got := output.Len() > 0; got != tc.want {
				t.Fatalf("credential output present = %t, want %t: %q", got, tc.want, output.String())
			}
		})
	}
}

func TestCredentialCommandRejectsNonHTTPSRequests(t *testing.T) {
	called := false
	runner := &credentialRunner{
		loadConfig: func() (*config.Config, error) {
			return &config.Config{Profiles: map[string]*config.Profile{
				"work": {Name: "Work User", Email: "work@example.com", Username: "work-user"},
			}, ActiveProfile: "work"}, nil
		},
		getToken: func(string) (string, error) {
			called = true
			return "secret", nil
		},
		stdin:  strings.NewReader("protocol=http\nhost=github.com\n\n"),
		stdout: &bytes.Buffer{},
	}

	runner.run(&cobra.Command{}, []string{"get"})
	if called {
		t.Fatal("credential helper looked up a token for a non-HTTPS request")
	}
}
