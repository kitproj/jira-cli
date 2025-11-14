package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/zalando/go-keyring"
)

const (
	serviceName = "jira-cli"
	configFile  = "config.json"
	tokenFile   = "token"
)

// config represents the jira-cli configuration
type config struct {
	Host string `json:"host"`
}

// getConfigPath returns the path to the config file
func getConfigPath() (string, error) {
	configDirPath, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get config directory: %w", err)
	}

	configPath := filepath.Join(configDirPath, "jira-cli", configFile)
	return configPath, nil
}

// SaveConfig saves the host to the config file
func SaveConfig(host string) error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	// Create config directory if it doesn't exist
	configDirPath := filepath.Dir(configPath)
	if err := os.MkdirAll(configDirPath, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	cfg := config{Host: host}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// LoadConfig loads the host from the config file
func LoadConfig() (string, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return "", fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return "", fmt.Errorf("failed to parse config file: %w", err)
	}

	return cfg.Host, nil
}

// SaveToken saves the token to the keyring, or falls back to file if keyring is unavailable
func SaveToken(host, token string) error {
	// Try keyring first
	err := keyring.Set(serviceName, host, token)
	if err == nil {
		return nil
	}

	// If keyring is unavailable (e.g., no dbus on Linux), fall back to file
	if isKeyringUnavailable(err) {
		return saveTokenToFile(host, token)
	}

	// Return the original error if it's not a keyring unavailability issue
	return err
}

// LoadToken loads the token from the keyring, or falls back to file if keyring is unavailable
func LoadToken(host string) (string, error) {
	// Try keyring first
	token, err := keyring.Get(serviceName, host)
	if err == nil {
		return token, nil
	}

	// If keyring is unavailable (e.g., no dbus on Linux), try file fallback
	if isKeyringUnavailable(err) {
		return loadTokenFromFile(host)
	}

	// If it's a "not found" error, also try file fallback
	// This handles the case where keyring is available but empty
	if err == keyring.ErrNotFound {
		fileToken, fileErr := loadTokenFromFile(host)
		if fileErr == nil {
			return fileToken, nil
		}
		// Return the original "not found" error
		return "", err
	}

	// Return the original error
	return "", err
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

// isKeyringUnavailable checks if the error indicates keyring is unavailable
// This happens on Linux systems without dbus/secret-service
func isKeyringUnavailable(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	// Common dbus errors when service is unavailable
	return strings.Contains(errMsg, "dbus") ||
		strings.Contains(errMsg, "DBus") ||
		strings.Contains(errMsg, "Cannot autolaunch") ||
		strings.Contains(errMsg, "secret service") ||
		strings.Contains(errMsg, "org.freedesktop.secrets") ||
		strings.Contains(errMsg, "dial unix") || // Unix socket connection errors
		strings.Contains(errMsg, "connection refused") ||
		strings.Contains(errMsg, "permission denied") // Permission issues with dbus socket
}

// saveTokenToFile saves the token to a file as a fallback
func saveTokenToFile(host, token string) error {
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

	// Add or update token for this host
	tokens[host] = token

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

// loadTokenFromFile loads the token from a file as a fallback
func loadTokenFromFile(host string) (string, error) {
	tokenPath, err := getTokenFilePath()
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(tokenPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("token not found in file")
		}
		return "", fmt.Errorf("failed to read token file: %w", err)
	}

	tokens := make(map[string]string)
	if err := json.Unmarshal(data, &tokens); err != nil {
		return "", fmt.Errorf("failed to parse token file: %w", err)
	}

	token, ok := tokens[host]
	if !ok {
		return "", fmt.Errorf("token not found for host: %s", host)
	}

	return token, nil
}
