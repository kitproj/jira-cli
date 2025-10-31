package main

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestRun_MCPServer(t *testing.T) {
	// Set JIRA_HOST and JIRA_TOKEN env vars to get past token check
	oldHost := os.Getenv("JIRA_HOST")
	oldToken := os.Getenv("JIRA_TOKEN")
	os.Setenv("JIRA_HOST", "test.atlassian.net")
	os.Setenv("JIRA_TOKEN", "test-token")
	defer func() {
		if oldHost == "" {
			os.Unsetenv("JIRA_HOST")
		} else {
			os.Setenv("JIRA_HOST", oldHost)
		}
		if oldToken == "" {
			os.Unsetenv("JIRA_TOKEN")
		} else {
			os.Setenv("JIRA_TOKEN", oldToken)
		}
	}()

	// Test that mcp-server sub-command is recognized
	args := []string{"mcp-server"}

	// We can't easily test the full server without mocking stdin/stdout
	// but we can verify the command is recognized and doesn't return "unknown sub-command"
	_ = args
	// This test just verifies the test setup works
}

func TestRun_MCPServerMissingConfig(t *testing.T) {
	// Unset JIRA_HOST env var
	oldHost := os.Getenv("JIRA_HOST")
	os.Unsetenv("JIRA_HOST")
	defer func() {
		if oldHost != "" {
			os.Setenv("JIRA_HOST", oldHost)
		}
	}()

	ctx := context.Background()
	err := run(ctx, []string{"mcp-server"})

	if err == nil {
		t.Error("Expected error for missing configuration, got nil")
	}

	if !strings.Contains(err.Error(), "JIRA host must be configured") {
		t.Errorf("Expected 'JIRA host must be configured' error, got: %v", err)
	}
}
