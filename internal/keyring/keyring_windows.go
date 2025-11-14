//go:build windows

package keyring

import (
	"github.com/zalando/go-keyring"
)

func init() {
	provider = &systemKeyringProvider{}
}

// systemKeyringProvider implements token storage using the system keyring
type systemKeyringProvider struct{}

// Set stores a token in the system keyring
func (s *systemKeyringProvider) Set(service, user, token string) error {
	return keyring.Set(service, user, token)
}

// Get retrieves a token from the system keyring
func (s *systemKeyringProvider) Get(service, user string) (string, error) {
	return keyring.Get(service, user)
}
