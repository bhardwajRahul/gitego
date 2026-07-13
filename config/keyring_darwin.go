// config/keyring_darwin.go

// This file will ONLY be compiled on macOS.
//go:build darwin

package config

import (
	"errors"
	"fmt"
	"os/exec"
)

// SetGitCredential directly overwrites the keychain entry that Git's osxkeychain helper reads.
func SetGitCredential(username, token string) error {
	// Replace only this account's credential. Do not delete credentials other
	// applications or identities use for github.com.
	_ = exec.Command("security", "delete-internet-password", "-a", username, "-s", "github.com").Run()

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

// DeleteGitCredential removes the active GitHub credential installed by gitego.
func DeleteGitCredential(username string) error {
	err := exec.Command("security", "delete-internet-password", "-a", username, "-s", "github.com").Run()
	if err != nil {
		var exitErr *exec.ExitError
		// security returns 44 when no matching item exists; deletion is then
		// already complete from the caller's point of view.
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 44 {
			return nil
		}
		return fmt.Errorf("failed to delete GitHub credential: %w", err)
	}
	return nil
}
