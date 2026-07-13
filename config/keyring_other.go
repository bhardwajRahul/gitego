// config/keyring_other.go

// This file is compiled on all systems EXCEPT darwin.
//go:build !darwin

package config

import "github.com/zalando/go-keyring"

// SetGitCredential sets the PAT in the location that Git's helper will read from.
// This is the "active slot" for github.com.
func SetGitCredential(username string, token string) error {
	// The username parameter is required to match the function signature on all platforms,
	// but this library uses the full URL as the user key for Git credentials.
	// The line below tells the Go compiler that we are intentionally not using it.
	_ = username

	// The service "git" and user "https://github.com" are the standard keys
	// used by the go-keyring library for Git credentials.
	return keyring.Set("git", "https://github.com", token)
}

// DeleteGitCredential is a no-op outside macOS; gitego's own keyring token is
// removed separately by DeleteToken.
func DeleteGitCredential(username string) error {
	_ = username
	return nil
}
