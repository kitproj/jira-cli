package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestSaveLoadTokenFile tests the file-based token storage
func TestSaveLoadTokenFile(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Override the config directory
	origConfigDir := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer func() {
		if origConfigDir != "" {
			os.Setenv("XDG_CONFIG_HOME", origConfigDir)
		} else {
			os.Unsetenv("XDG_CONFIG_HOME")
		}
	}()

	testHost := "test.atlassian.net"
	testToken := "test-token-12345"

	// Test saving token to file
	err := saveTokenToFile(testHost, testToken)
	if err != nil {
		t.Fatalf("Failed to save token to file: %v", err)
	}

	// Test loading token from file
	loadedToken, err := loadTokenFromFile(testHost)
	if err != nil {
		t.Fatalf("Failed to load token from file: %v", err)
	}

	if loadedToken != testToken {
		t.Errorf("Expected token %q, got %q", testToken, loadedToken)
	}

	// Verify file permissions
	tokenPath, err := getTokenFilePath()
	if err != nil {
		t.Fatalf("Failed to get token file path: %v", err)
	}

	info, err := os.Stat(tokenPath)
	if err != nil {
		t.Fatalf("Failed to stat token file: %v", err)
	}

	// Check that permissions are 0600 (owner read/write only)
	expectedPerm := os.FileMode(0600)
	if info.Mode().Perm() != expectedPerm {
		t.Errorf("Expected file permissions %v, got %v", expectedPerm, info.Mode().Perm())
	}
}

// TestSaveLoadTokenFileMultipleHosts tests storing tokens for multiple hosts
func TestSaveLoadTokenFileMultipleHosts(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Override the config directory
	origConfigDir := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer func() {
		if origConfigDir != "" {
			os.Setenv("XDG_CONFIG_HOME", origConfigDir)
		} else {
			os.Unsetenv("XDG_CONFIG_HOME")
		}
	}()

	hosts := map[string]string{
		"host1.atlassian.net": "token1",
		"host2.atlassian.net": "token2",
		"host3.atlassian.net": "token3",
	}

	// Save all tokens
	for host, token := range hosts {
		err := saveTokenToFile(host, token)
		if err != nil {
			t.Fatalf("Failed to save token for %s: %v", host, err)
		}
	}

	// Load and verify all tokens
	for host, expectedToken := range hosts {
		loadedToken, err := loadTokenFromFile(host)
		if err != nil {
			t.Fatalf("Failed to load token for %s: %v", host, err)
		}
		if loadedToken != expectedToken {
			t.Errorf("For host %s, expected token %q, got %q", host, expectedToken, loadedToken)
		}
	}
}

// TestLoadTokenFileNotFound tests error handling when token file doesn't exist
func TestLoadTokenFileNotFound(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Override the config directory
	origConfigDir := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer func() {
		if origConfigDir != "" {
			os.Setenv("XDG_CONFIG_HOME", origConfigDir)
		} else {
			os.Unsetenv("XDG_CONFIG_HOME")
		}
	}()

	// Try to load from non-existent file
	_, err := loadTokenFromFile("nonexistent.atlassian.net")
	if err == nil {
		t.Error("Expected error when loading from non-existent file, got nil")
	}
}

// TestIsKeyringUnavailable tests the error detection function
func TestIsKeyringUnavailable(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "dbus error",
			err:      &testError{"failed to connect to dbus"},
			expected: true,
		},
		{
			name:     "DBus error uppercase",
			err:      &testError{"DBus connection failed"},
			expected: true,
		},
		{
			name:     "autolaunch error",
			err:      &testError{"Cannot autolaunch D-Bus"},
			expected: true,
		},
		{
			name:     "secret service error",
			err:      &testError{"secret service not available"},
			expected: true,
		},
		{
			name:     "freedesktop error",
			err:      &testError{"org.freedesktop.secrets not found"},
			expected: true,
		},
		{
			name:     "unix socket error",
			err:      &testError{"dial unix /run/user/0/bus: connect: no such file"},
			expected: true,
		},
		{
			name:     "connection refused error",
			err:      &testError{"connection refused"},
			expected: true,
		},
		{
			name:     "permission denied error",
			err:      &testError{"dial unix /run/user/0/bus: connect: permission denied"},
			expected: true,
		},
		{
			name:     "unrelated error",
			err:      &testError{"some other error"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isKeyringUnavailable(tt.err)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// testError is a simple error type for testing
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

// TestConfigDirectory tests that config directory is created with correct permissions
func TestConfigDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	origConfigDir := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer func() {
		if origConfigDir != "" {
			os.Setenv("XDG_CONFIG_HOME", origConfigDir)
		} else {
			os.Unsetenv("XDG_CONFIG_HOME")
		}
	}()

	// Save a token which should create the directory
	err := saveTokenToFile("test.atlassian.net", "test-token")
	if err != nil {
		t.Fatalf("Failed to save token: %v", err)
	}

	// Check directory exists and has correct permissions
	configDir := filepath.Join(tmpDir, "jira-cli")
	info, err := os.Stat(configDir)
	if err != nil {
		t.Fatalf("Config directory not created: %v", err)
	}

	if !info.IsDir() {
		t.Error("Config path is not a directory")
	}

	// Check directory permissions are 0700
	expectedPerm := os.FileMode(0700)
	if info.Mode().Perm() != expectedPerm {
		t.Errorf("Expected directory permissions %v, got %v", expectedPerm, info.Mode().Perm())
	}
}
