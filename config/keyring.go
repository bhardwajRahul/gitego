// config/keyring.go

package config

import (
	"errors"
	"github.com/zalando/go-keyring"
)

// Use a constant for the service name to avoid typos.
// This is the name under which gitego stores its own library of PATs.
const gitegoKeyringService = "gitego"

// SetToken securely stores a PAT for a given profile name in gitego's vault.
func SetToken(profileName, token string) error {
	return keyring.Set(gitegoKeyringService, profileName, token)
}

// GetToken securely retrieves a PAT for a given profile name from gitego's vault.
func GetToken(profileName string) (string, error) {
	return keyring.Get(gitegoKeyringService, profileName)
}

// DeleteToken securely removes a PAT for a given profile name from gitego's vault.
func DeleteToken(profileName string) error {
	return keyring.Delete(gitegoKeyringService, profileName)
}

func IsTokenNotFound(err error) bool { return errors.Is(err, keyring.ErrNotFound) }
