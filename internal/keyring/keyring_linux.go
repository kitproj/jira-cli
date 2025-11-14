//go:build linux

package keyring

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const tokenFile = "token"

func init() {
	provider = &fileProvider{}
}

// fileProvider implements token storage using files
type fileProvider struct{}

// Set stores a token in a file
func (f *fileProvider) Set(service, user, token string) error {
	tokenPath, err := getTokenFilePath()
	if err != nil {
		return err
	}

	// Create config directory if it doesn't exist
	configDirPath := filepath.Dir(tokenPath)
	if err := os.MkdirAll(configDirPath, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Store as a simple key-value map
	tokens := make(map[string]string)

	// Try to load existing tokens
	if data, err := os.ReadFile(tokenPath); err == nil {
		_ = json.Unmarshal(data, &tokens)
	}

	// Add or update token for this user
	tokens[user] = token

	// Save to file
	data, err := json.MarshalIndent(tokens, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tokens: %w", err)
	}

	// Write with 0600 permissions (only owner can read/write)
	if err := os.WriteFile(tokenPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write token file: %w", err)
	}

	return nil
}

// Get retrieves a token from a file
func (f *fileProvider) Get(service, user string) (string, error) {
	tokenPath, err := getTokenFilePath()
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(tokenPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("token not found")
		}
		return "", fmt.Errorf("failed to read token file: %w", err)
	}

	tokens := make(map[string]string)
	if err := json.Unmarshal(data, &tokens); err != nil {
		return "", fmt.Errorf("failed to parse token file: %w", err)
	}

	token, ok := tokens[user]
	if !ok {
		return "", fmt.Errorf("token not found for user: %s", user)
	}

	return token, nil
}

// getTokenFilePath returns the path to the token file
func getTokenFilePath() (string, error) {
	configDirPath, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get config directory: %w", err)
	}

	tokenPath := filepath.Join(configDirPath, "jira-cli", tokenFile)
	return tokenPath, nil
}
