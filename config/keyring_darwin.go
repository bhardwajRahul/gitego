// config/keyring_darwin.go

// This file will ONLY be compiled on macOS.
//go:build darwin

package config

import (
	"fmt"
	"os/exec"
)

// SetGitCredential directly overwrites the keychain entry that Git's osxkeychain helper reads.
func SetGitCredential(username, token string) error {
	// Attempt to delete ALL existing password for github.com first.
	_ = exec.Command("security", "delete-internet-password", "-s", "github.com").Run()

	// Add the new password.
	cmd := exec.Command(
		"security",
		"add-internet-password",
		"-a", username,
		"-s", "github.com",
		"-r", "htps", // protocol
		"-w", token, // password
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to run 'security' command: %w\nOutput: %s", err, string(output))
	}

	return nil
}
