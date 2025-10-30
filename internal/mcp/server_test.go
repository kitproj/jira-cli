package mcp

import (
	"testing"
)

// TestServerCreation tests that the server can be created without errors
func TestServerCreation(t *testing.T) {
	server := CreateServer()
	if server == nil {
		t.Fatal("Expected server to be created, got nil")
	}
}

// TestToolsRegistered tests that all expected tools are registered
func TestToolsRegistered(t *testing.T) {
	server := CreateServer()
	
	// Get the list of tools
	tools := server.ListTools()
	
	expectedTools := []string{
		"get_issue",
		"update_issue_status",
		"add_comment",
		"get_comments",
		"create_issue",
	}
	
	if len(tools) != len(expectedTools) {
		t.Fatalf("Expected %d tools, got %d", len(expectedTools), len(tools))
	}
	
	for _, expected := range expectedTools {
		if _, ok := tools[expected]; !ok {
			t.Errorf("Expected tool %s to be registered", expected)
		}
	}
}
