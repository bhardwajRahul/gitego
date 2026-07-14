// utils/git_test.go

package utils

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// mockExecCommand remains the same.
func mockExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}

	return cmd
}

func TestGetEffectiveGitConfig(t *testing.T) {
	// Store the original function and defer its restoration.
	originalExecCommand := execCommand
	// Patch the package-level variable.
	execCommand = mockExecCommand

	defer func() { execCommand = originalExecCommand }()

	val, err := GetEffectiveGitConfig("user.email")
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}

	if val != "test@example.com" {
		t.Errorf("expected 'test@example.com', but got '%s'", val)
	}
}

func TestSetGlobalGitConfig(t *testing.T) {
	// Store the original function and defer its restoration.
	originalExecCommand := execCommand
	// Patch the package-level variable.
	execCommand = mockExecCommand

	defer func() { execCommand = originalExecCommand }()

	err := SetGlobalGitConfig("user.name", "Test User")
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
}

func TestLocalGitConfigLifecycle(t *testing.T) {
	repoDir := t.TempDir()
	if output, err := exec.Command("git", "init", repoDir).CombinedOutput(); err != nil {
		t.Fatalf("initialize Git repository: %v: %s", err, output)
	}

	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(repoDir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(originalWd); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})

	if err := SetLocalGitConfig("user.email", "local@example.com"); err != nil {
		t.Fatal(err)
	}
	if got, err := GetEffectiveGitConfig("user.email"); err != nil || got != "local@example.com" {
		t.Fatalf("effective local email = %q, %v", got, err)
	}
	if err := UnsetLocalGitConfig("user.email"); err != nil {
		t.Fatal(err)
	}
	if err := UnsetLocalGitConfig("user.email"); err != nil {
		t.Fatalf("unsetting absent local value: %v", err)
	}
}

func TestGlobalGitConfigValuesAndMissingUnset(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Setenv("USERPROFILE", homeDir)

	for _, value := range []string{"first", "second"} {
		if output, err := exec.Command("git", "config", "--global", "--add", "credential.helper", value).CombinedOutput(); err != nil {
			t.Fatalf("add global credential helper: %v: %s", err, output)
		}
	}
	values, err := GetGlobalGitConfigValues("credential.helper")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(values, ",") != "first,second" {
		t.Fatalf("global credential helpers = %v", values)
	}
	if err := UnsetGlobalGitConfig("user.signingkey"); err != nil {
		t.Fatalf("unsetting absent global value: %v", err)
	}
}

// TestHelperProcess remains the same.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	defer os.Exit(0)

	args := extractCommandArgs()
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "No command to mock\n")
		os.Exit(1)
	}

	if handleGitConfigCommands(args) {
		return
	}

	fmt.Fprintf(os.Stderr, "unhandled mock command: %s\n", strings.Join(args, " "))
	os.Exit(1)
}

func extractCommandArgs() []string {
	args := os.Args
	for len(args) > 0 {
		if args[0] == "--" {
			return args[1:]
		}

		args = args[1:]
	}

	return args
}

func handleGitConfigCommands(args []string) bool {
	if len(args) < 2 || args[0] != "git" || args[1] != "config" {
		return false
	}

	if len(args) == 3 && args[2] == "user.email" {
		if _, err := fmt.Fprint(os.Stdout, "test@example.com"); err != nil {
			panic("Failed to write to stdout: " + err.Error())
		}

		return true
	}

	if len(args) == 5 && args[2] == "--global" && args[3] == "user.name" {
		return true
	}

	return false
}
