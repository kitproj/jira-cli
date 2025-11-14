package keyring

import (
	"os"
	"path/filepath"
	"testing"
)

// TestSetGet tests basic set and get operations
func TestSetGet(t *testing.T) {
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

	testService := "test-service"
	testUser := "test.atlassian.net"
	testToken := "test-token-12345"

	// Test Set
	err := Set(testService, testUser, testToken)
	if err != nil {
		t.Fatalf("Failed to set token: %v", err)
	}

	// Test Get
	retrievedToken, err := Get(testService, testUser)
	if err != nil {
		t.Fatalf("Failed to get token: %v", err)
	}

	if retrievedToken != testToken {
		t.Errorf("Expected token %q, got %q", testToken, retrievedToken)
	}
}

// TestMultipleUsers tests storing tokens for multiple users
func TestMultipleUsers(t *testing.T) {
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

	testService := "test-service"
	users := map[string]string{
		"host1.atlassian.net": "token1",
		"host2.atlassian.net": "token2",
		"host3.atlassian.net": "token3",
	}

	// Set all tokens
	for user, token := range users {
		err := Set(testService, user, token)
		if err != nil {
			t.Fatalf("Failed to set token for %s: %v", user, err)
		}
	}

	// Get and verify all tokens
	for user, expectedToken := range users {
		retrievedToken, err := Get(testService, user)
		if err != nil {
			t.Fatalf("Failed to get token for %s: %v", user, err)
		}
		if retrievedToken != expectedToken {
			t.Errorf("For user %s, expected token %q, got %q", user, expectedToken, retrievedToken)
		}
	}
}

// TestGetNotFound tests error handling when token doesn't exist
func TestGetNotFound(t *testing.T) {
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

	// Try to get from non-existent file
	_, err := Get("test-service", "nonexistent.atlassian.net")
	if err == nil {
		t.Error("Expected error when getting non-existent token, got nil")
	}
}

// TestFilePermissions tests that token file has correct permissions (Linux only)
func TestFilePermissions(t *testing.T) {
	// Skip on non-Linux platforms since file implementation is Linux-specific
	if provider == nil {
		t.Skip("Skipping file permissions test on non-Linux platform")
	}

	// Check if provider is fileProvider (Linux)
	if _, ok := provider.(*fileProvider); !ok {
		t.Skip("Skipping file permissions test on non-Linux platform")
	}

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

	// Set a token
	err := Set("test-service", "test.atlassian.net", "test-token")
	if err != nil {
		t.Fatalf("Failed to set token: %v", err)
	}

	// Check file permissions
	tokenPath := filepath.Join(tmpDir, "jira-cli", "token")
	info, err := os.Stat(tokenPath)
	if err != nil {
		t.Fatalf("Failed to stat token file: %v", err)
	}

	expectedPerm := os.FileMode(0600)
	if info.Mode().Perm() != expectedPerm {
		t.Errorf("Expected file permissions %v, got %v", expectedPerm, info.Mode().Perm())
	}
}
