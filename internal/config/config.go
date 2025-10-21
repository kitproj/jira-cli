package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/zalando/go-keyring"
)

const (
	serviceName = "jira-cli"
	configFile  = "config.json"
)

// Config represents the jira-cli configuration
type Config struct {
	Host string `json:"host"`
}

// GetConfigPath returns the path to the config file
func GetConfigPath() (string, error) {
	configDirPath, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get config directory: %w", err)
	}

	configPath := filepath.Join(configDirPath, "jira-cli", configFile)
	return configPath, nil
}

// SaveConfig saves the host to the config file
func SaveConfig(host string) error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	// Create config directory if it doesn't exist
	configDirPath := filepath.Dir(configPath)
	if err := os.MkdirAll(configDirPath, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	config := Config{Host: host}
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// LoadConfig loads the host from the config file
func LoadConfig() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config file not found, please run 'jira configure' first")
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// SaveToken saves the token to the keyring
func SaveToken(host, token string) error {
	if err := keyring.Set(serviceName, host, token); err != nil {
		return fmt.Errorf("failed to save token to keyring: %w", err)
	}
	return nil
}

// LoadToken loads the token from the keyring
func LoadToken(host string) (string, error) {
	token, err := keyring.Get(serviceName, host)
	if err != nil {
		if err == keyring.ErrNotFound {
			return "", fmt.Errorf("token not found for host %s, please run 'jira configure' first", host)
		}
		return "", fmt.Errorf("failed to get token from keyring: %w", err)
	}
	return token, nil
}
