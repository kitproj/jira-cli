package mcp

import (
	"context"
	"fmt"

	"github.com/andygrunwald/go-jira"
	"github.com/kitproj/jira-cli/internal/config"
	mcplib "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// CreateServer creates a new MCP server with JIRA tools
func CreateServer() *server.MCPServer {
	s := server.NewMCPServer(
		"JIRA CLI MCP Server",
		"1.0.0",
		server.WithToolCapabilities(false),
	)

	// Add get-issue tool
	getIssueTool := mcplib.NewTool("get_issue",
		mcplib.WithDescription("Get details of a JIRA issue"),
		mcplib.WithString("issue_key",
			mcplib.Required(),
			mcplib.Description("JIRA issue key (e.g., PROJ-123)"),
		),
	)
	s.AddTool(getIssueTool, getIssueHandler)

	// Add update-issue-status tool
	updateStatusTool := mcplib.NewTool("update_issue_status",
		mcplib.WithDescription("Update the status of a JIRA issue"),
		mcplib.WithString("issue_key",
			mcplib.Required(),
			mcplib.Description("JIRA issue key (e.g., PROJ-123)"),
		),
		mcplib.WithString("status",
			mcplib.Required(),
			mcplib.Description("New status name (e.g., 'In Progress', 'Closed')"),
		),
	)
	s.AddTool(updateStatusTool, updateIssueStatusHandler)

	// Add add-comment tool
	addCommentTool := mcplib.NewTool("add_comment",
		mcplib.WithDescription("Add a comment to a JIRA issue"),
		mcplib.WithString("issue_key",
			mcplib.Required(),
			mcplib.Description("JIRA issue key (e.g., PROJ-123)"),
		),
		mcplib.WithString("comment",
			mcplib.Required(),
			mcplib.Description("Comment text to add"),
		),
	)
	s.AddTool(addCommentTool, addCommentHandler)

	// Add get-comments tool
	getCommentsTool := mcplib.NewTool("get_comments",
		mcplib.WithDescription("Get comments on a JIRA issue"),
		mcplib.WithString("issue_key",
			mcplib.Required(),
			mcplib.Description("JIRA issue key (e.g., PROJ-123)"),
		),
	)
	s.AddTool(getCommentsTool, getCommentsHandler)

	// Add create-issue tool
	createIssueTool := mcplib.NewTool("create_issue",
		mcplib.WithDescription("Create a new JIRA issue"),
		mcplib.WithString("project",
			mcplib.Required(),
			mcplib.Description("JIRA project key"),
		),
		mcplib.WithString("description",
			mcplib.Required(),
			mcplib.Description("Issue description"),
		),
		mcplib.WithString("assignee",
			mcplib.Description("Optional assignee username"),
		),
	)
	s.AddTool(createIssueTool, createIssueHandler)

	return s
}

// getJiraClient creates and configures a JIRA client
func getJiraClient(ctx context.Context) (*jira.Client, error) {
	// Load host from config file
	host, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Load token from keyring
	token, err := config.LoadToken(host)
	if err != nil {
		return nil, fmt.Errorf("failed to load token: %w", err)
	}

	if host == "" {
		return nil, fmt.Errorf("host is required")
	}
	if token == "" {
		return nil, fmt.Errorf("token is required")
	}

	tp := jira.BearerAuthTransport{Token: token}
	client, err := jira.NewClient(tp.Client(), "https://"+host)
	if err != nil {
		return nil, fmt.Errorf("failed to create JIRA client: %w", err)
	}

	return client, nil
}

func getIssueHandler(ctx context.Context, request mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	issueKey, err := request.RequireString("issue_key")
	if err != nil {
		return mcplib.NewToolResultError(err.Error()), nil
	}

	client, err := getJiraClient(ctx)
	if err != nil {
		return mcplib.NewToolResultError(err.Error()), nil
	}

	issue, _, err := client.Issue.GetWithContext(ctx, issueKey, nil)
	if err != nil {
		return mcplib.NewToolResultError(fmt.Sprintf("failed to get issue: %v", err)), nil
	}

	result := fmt.Sprintf("Key: %s\nStatus: %s\nSummary: %s\nReporter: %s (%s)\nDescription: %s",
		issue.Key,
		issue.Fields.Status.Name,
		issue.Fields.Summary,
		issue.Fields.Reporter.DisplayName,
		issue.Fields.Reporter.Name,
		issue.Fields.Description,
	)

	return mcplib.NewToolResultText(result), nil
}

func updateIssueStatusHandler(ctx context.Context, request mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	issueKey, err := request.RequireString("issue_key")
	if err != nil {
		return mcplib.NewToolResultError(err.Error()), nil
	}

	statusName, err := request.RequireString("status")
	if err != nil {
		return mcplib.NewToolResultError(err.Error()), nil
	}

	client, err := getJiraClient(ctx)
	if err != nil {
		return mcplib.NewToolResultError(err.Error()), nil
	}

	// Get the issue to check current status
	issue, _, err := client.Issue.GetWithContext(ctx, issueKey, nil)
	if err != nil {
		return mcplib.NewToolResultError(fmt.Sprintf("failed to get issue: %v", err)), nil
	}

	// Check if already in the desired status
	if issue.Fields.Status.Name == statusName {
		return mcplib.NewToolResultText(fmt.Sprintf("Issue %s is already in status: %s", issueKey, statusName)), nil
	}

	// Get available transitions
	transitions, _, err := client.Issue.GetTransitionsWithContext(ctx, issueKey)
	if err != nil {
		return mcplib.NewToolResultError(fmt.Sprintf("failed to get transitions: %v", err)), nil
	}

	// Find the transition that leads to the desired status
	var targetTransition *jira.Transition
	for _, transition := range transitions {
		if transition.To.Name == statusName {
			targetTransition = &transition
			break
		}
	}

	if targetTransition == nil {
		var availableStatuses []string
		for _, transition := range transitions {
			availableStatuses = append(availableStatuses, fmt.Sprintf("%q", transition.To.Name))
		}
		return mcplib.NewToolResultError(fmt.Sprintf("no transition found to status '%s'. Available statuses: %v", statusName, availableStatuses)), nil
	}

	// Perform the transition
	_, err = client.Issue.DoTransition(issueKey, targetTransition.ID)
	if err != nil {
		return mcplib.NewToolResultError(fmt.Sprintf("failed to update issue status: %v", err)), nil
	}

	return mcplib.NewToolResultText(fmt.Sprintf("Successfully updated issue %s to status: %s", issueKey, statusName)), nil
}

func addCommentHandler(ctx context.Context, request mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	issueKey, err := request.RequireString("issue_key")
	if err != nil {
		return mcplib.NewToolResultError(err.Error()), nil
	}

	commentText, err := request.RequireString("comment")
	if err != nil {
		return mcplib.NewToolResultError(err.Error()), nil
	}

	client, err := getJiraClient(ctx)
	if err != nil {
		return mcplib.NewToolResultError(err.Error()), nil
	}

	comment := &jira.Comment{
		Body: commentText,
	}

	_, _, err = client.Issue.AddCommentWithContext(ctx, issueKey, comment)
	if err != nil {
		return mcplib.NewToolResultError(fmt.Sprintf("failed to add comment: %v", err)), nil
	}

	return mcplib.NewToolResultText(fmt.Sprintf("Successfully added comment to issue %s", issueKey)), nil
}

func getCommentsHandler(ctx context.Context, request mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	issueKey, err := request.RequireString("issue_key")
	if err != nil {
		return mcplib.NewToolResultError(err.Error()), nil
	}

	client, err := getJiraClient(ctx)
	if err != nil {
		return mcplib.NewToolResultError(err.Error()), nil
	}

	options := &jira.GetQueryOptions{
		Expand: "comments",
	}

	issue, _, err := client.Issue.GetWithContext(ctx, issueKey, options)
	if err != nil {
		return mcplib.NewToolResultError(fmt.Sprintf("failed to get issue with comments: %v", err)), nil
	}

	if issue.Fields.Comments == nil || len(issue.Fields.Comments.Comments) == 0 {
		return mcplib.NewToolResultText("No comments found"), nil
	}

	result := ""
	for i, comment := range issue.Fields.Comments.Comments {
		if i > 0 {
			result += "\n---\n"
		}
		result += fmt.Sprintf("%s (%s):\n%s", comment.Author.DisplayName, comment.Author.Name, comment.Body)
	}

	return mcplib.NewToolResultText(result), nil
}

func createIssueHandler(ctx context.Context, request mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	projectKey, err := request.RequireString("project")
	if err != nil {
		return mcplib.NewToolResultError(err.Error()), nil
	}

	description, err := request.RequireString("description")
	if err != nil {
		return mcplib.NewToolResultError(err.Error()), nil
	}

	assignee := request.GetString("assignee", "")

	client, err := getJiraClient(ctx)
	if err != nil {
		return mcplib.NewToolResultError(err.Error()), nil
	}

	// Create a new issue
	issue := &jira.Issue{
		Fields: &jira.IssueFields{
			Project: jira.Project{
				Key: projectKey,
			},
			Summary:     description,
			Description: description,
			Type: jira.IssueType{
				Name: "Task",
			},
		},
	}

	// Add assignee if provided
	if assignee != "" {
		issue.Fields.Assignee = &jira.User{
			Name: assignee,
		}
	}

	// Create the issue
	createdIssue, _, err := client.Issue.Create(issue)
	if err != nil {
		return mcplib.NewToolResultError(fmt.Sprintf("failed to create issue: %v", err)), nil
	}

	return mcplib.NewToolResultText(fmt.Sprintf("Successfully created issue: %s", createdIssue.Key)), nil
}
