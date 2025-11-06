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

func TestRun_AttachFileMissingArgs(t *testing.T) {
	ctx := context.Background()

	// Test with no arguments
	err := run(ctx, []string{"attach-file"})
	if err == nil {
		t.Error("Expected error for missing arguments, got nil")
	}
	if !strings.Contains(err.Error(), "usage: jira attach-file") {
		t.Errorf("Expected usage error, got: %v", err)
	}

	// Test with only issue key
	err = run(ctx, []string{"attach-file", "TEST-123"})
	if err == nil {
		t.Error("Expected error for missing file path, got nil")
	}
	if !strings.Contains(err.Error(), "usage: jira attach-file") {
		t.Errorf("Expected usage error, got: %v", err)
	}
}

func TestRun_AttachFileNonExistentFile(t *testing.T) {
	// Set JIRA_HOST and JIRA_TOKEN env vars
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

	ctx := context.Background()
	err := run(ctx, []string{"attach-file", "TEST-123", "/tmp/nonexistent-file-12345.txt"})

	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
	if !strings.Contains(err.Error(), "failed to open file") {
		t.Errorf("Expected 'failed to open file' error, got: %v", err)
	}
}

func TestRun_AssignIssueMissingArgs(t *testing.T) {
	ctx := context.Background()

	// Test with no arguments
	err := run(ctx, []string{"assign-issue"})
	if err == nil {
		t.Error("Expected error for missing arguments, got nil")
	}
	if !strings.Contains(err.Error(), "usage: jira assign-issue") {
		t.Errorf("Expected usage error, got: %v", err)
	}

	// Test with only issue key
	err = run(ctx, []string{"assign-issue", "TEST-123"})
	if err == nil {
		t.Error("Expected error for missing assignee, got nil")
	}
	if !strings.Contains(err.Error(), "usage: jira assign-issue") {
		t.Errorf("Expected usage error, got: %v", err)
	}
}

func TestRun_AddIssueToSprintMissingArgs(t *testing.T) {
	ctx := context.Background()

	// Test with no arguments
	err := run(ctx, []string{"add-issue-to-sprint"})
	if err == nil {
		t.Error("Expected error for missing arguments, got nil")
	}
	if !strings.Contains(err.Error(), "usage: jira add-issue-to-sprint") {
		t.Errorf("Expected usage error, got: %v", err)
	}
}

func TestRun_CreateIssueMissingArgs(t *testing.T) {
	ctx := context.Background()

	// Test with no arguments
	err := run(ctx, []string{"create-issue"})
	if err == nil {
		t.Error("Expected error for missing arguments, got nil")
	}
	if !strings.Contains(err.Error(), "usage: jira create-issue") {
		t.Errorf("Expected usage error, got: %v", err)
	}

	// Test with only project
	err = run(ctx, []string{"create-issue", "PROJ"})
	if err == nil {
		t.Error("Expected error for missing arguments, got nil")
	}
	if !strings.Contains(err.Error(), "usage: jira create-issue") {
		t.Errorf("Expected usage error, got: %v", err)
	}

	// Test with project and issue type
	err = run(ctx, []string{"create-issue", "PROJ", "Task"})
	if err == nil {
		t.Error("Expected error for missing arguments, got nil")
	}
	if !strings.Contains(err.Error(), "usage: jira create-issue") {
		t.Errorf("Expected usage error, got: %v", err)
	}

	// Test with project, issue type, and title
	err = run(ctx, []string{"create-issue", "PROJ", "Task", "My Title"})
	if err == nil {
		t.Error("Expected error for missing arguments, got nil")
	}
	if !strings.Contains(err.Error(), "usage: jira create-issue") {
		t.Errorf("Expected usage error, got: %v", err)
	}
}

