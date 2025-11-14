package keyring

// Provider defines the interface for token storage
type Provider interface {
	// Set stores a token for the given service and user
	Set(service, user, token string) error
	// Get retrieves a token for the given service and user
	Get(service, user string) (string, error)
}

// provider is the platform-specific implementation
var provider Provider

// Set stores a token using the platform-specific provider
func Set(service, user, token string) error {
	return provider.Set(service, user, token)
}

// Get retrieves a token using the platform-specific provider
func Get(service, user string) (string, error) {
	return provider.Get(service, user)
}
