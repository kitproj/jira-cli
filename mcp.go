package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/andygrunwald/go-jira"
	"github.com/kitproj/jira-cli/internal/config"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// runMCPServer starts the MCP server that communicates over stdio using the mcp-go library
func runMCPServer(ctx context.Context) error {
	// Load host from config file
	host, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("JIRA host must be configured (use 'jira configure <host>' or set JIRA_HOST env var)")
	}

	// Load token from keyring
	token, err := config.LoadToken(host)
	if err != nil {
		return fmt.Errorf("JIRA token must be set (use 'jira configure <host>' or set JIRA_TOKEN env var)")
	}

	if host == "" {
		return fmt.Errorf("JIRA host must be configured (use 'jira configure <host>')")
	}
	if token == "" {
		return fmt.Errorf("JIRA token must be set (use 'jira configure <host>')")
	}

	tp := jira.BearerAuthTransport{Token: token}
	api, err := jira.NewClient(tp.Client(), "https://"+host)
	if err != nil {
		return fmt.Errorf("failed to create JIRA client: %w", err)
	}

	// Create a new MCP server
	s := server.NewMCPServer(
		"jira-cli-mcp-server",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	// Add get-issue tool
	getIssueTool := mcp.NewTool("get_issue",
		mcp.WithDescription("Get details of a JIRA issue including status, summary, reporter, and description"),
		mcp.WithString("issue_key",
			mcp.Required(),
			mcp.Description("JIRA issue key (e.g., 'PROJ-123')"),
		),
	)
	s.AddTool(getIssueTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return getIssueHandler(ctx, api, request)
	})

	// Add update-issue-status tool
	updateStatusTool := mcp.NewTool("update_issue_status",
		mcp.WithDescription("Update the status of a JIRA issue using transitions (e.g., move to 'In Progress' or 'Closed')"),
		mcp.WithString("issue_key",
			mcp.Required(),
			mcp.Description("JIRA issue key (e.g., 'PROJ-123')"),
		),
		mcp.WithString("status",
			mcp.Required(),
			mcp.Description("New status name (e.g., 'In Progress', 'Closed')"),
		),
	)
	s.AddTool(updateStatusTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return updateIssueStatusHandler(ctx, api, request)
	})

	// Add add-comment tool
	addCommentTool := mcp.NewTool("add_comment",
		mcp.WithDescription("Add a comment to a JIRA issue"),
		mcp.WithString("issue_key",
			mcp.Required(),
			mcp.Description("JIRA issue key (e.g., 'PROJ-123')"),
		),
		mcp.WithString("comment",
			mcp.Required(),
			mcp.Description("Comment text to add"),
		),
	)
	s.AddTool(addCommentTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return addCommentHandler(ctx, api, request)
	})

	// Add get-comments tool
	getCommentsTool := mcp.NewTool("get_comments",
		mcp.WithDescription("Get all comments on a JIRA issue"),
		mcp.WithString("issue_key",
			mcp.Required(),
			mcp.Description("JIRA issue key (e.g., 'PROJ-123')"),
		),
	)
	s.AddTool(getCommentsTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return getCommentsHandler(ctx, api, request)
	})

	// Add create-issue tool
	createIssueTool := mcp.NewTool("create_issue",
		mcp.WithDescription("Create a new JIRA issue with the specified project, description, and optional assignee"),
		mcp.WithString("project",
			mcp.Required(),
			mcp.Description("JIRA project key"),
		),
		mcp.WithString("description",
			mcp.Required(),
			mcp.Description("Issue description (used as both summary and description)"),
		),
		mcp.WithString("assignee",
			mcp.Description("Optional assignee username"),
		),
	)
	s.AddTool(createIssueTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return createIssueHandler(ctx, api, request)
	})

	// Start the stdio server
	return server.ServeStdio(s)
}

func getIssueHandler(ctx context.Context, client *jira.Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	issueKey, err := request.RequireString("issue_key")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Missing or invalid 'issue_key' argument: %v", err)), nil
	}

	issue, _, err := client.Issue.GetWithContext(ctx, issueKey, nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get issue: %v", err)), nil
	}

	result := fmt.Sprintf("Key: %s\nStatus: %s\nSummary: %s\nReporter: %s (%s)\nDescription: %s",
		issue.Key,
		issue.Fields.Status.Name,
		issue.Fields.Summary,
		issue.Fields.Reporter.DisplayName,
		issue.Fields.Reporter.Name,
		issue.Fields.Description,
	)

	// Get editable custom fields
	editMetaInfo, _, err := client.Issue.GetEditMetaWithContext(ctx, issue)
	if err == nil {
		for key, value := range issue.Fields.Unknowns {
			if _, ok := editMetaInfo.Fields[key]; ok && strings.HasPrefix(key, "customfield_") && isPrimitive(value) {
				name := editMetaInfo.Fields[key].(map[string]any)["name"].(string)
				result += fmt.Sprintf("\n%s: %v", name, value)
			}
		}
	}

	return mcp.NewToolResultText(result), nil
}

func updateIssueStatusHandler(ctx context.Context, client *jira.Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	issueKey, err := request.RequireString("issue_key")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Missing or invalid 'issue_key' argument: %v", err)), nil
	}

	statusName, err := request.RequireString("status")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Missing or invalid 'status' argument: %v", err)), nil
	}

	// Get the issue to check current status
	issue, _, err := client.Issue.GetWithContext(ctx, issueKey, nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get issue: %v", err)), nil
	}

	// Check if already in the desired status
	if issue.Fields.Status.Name == statusName {
		return mcp.NewToolResultText(fmt.Sprintf("Issue %s is already in status: %s", issueKey, statusName)), nil
	}

	// Get available transitions
	transitions, _, err := client.Issue.GetTransitionsWithContext(ctx, issueKey)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get transitions: %v", err)), nil
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
		return mcp.NewToolResultError(fmt.Sprintf("No transition found to status '%s'. Available statuses: %v", statusName, strings.Join(availableStatuses, ", "))), nil
	}

	// Perform the transition
	_, err = client.Issue.DoTransition(issueKey, targetTransition.ID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to update issue status: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Successfully updated issue %s to status: %s", issueKey, statusName)), nil
}

func addCommentHandler(ctx context.Context, client *jira.Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	issueKey, err := request.RequireString("issue_key")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Missing or invalid 'issue_key' argument: %v", err)), nil
	}

	commentText, err := request.RequireString("comment")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Missing or invalid 'comment' argument: %v", err)), nil
	}

	comment := &jira.Comment{
		Body: commentText,
	}

	_, _, err = client.Issue.AddCommentWithContext(ctx, issueKey, comment)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to add comment: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Successfully added comment to issue %s", issueKey)), nil
}

func getCommentsHandler(ctx context.Context, client *jira.Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	issueKey, err := request.RequireString("issue_key")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Missing or invalid 'issue_key' argument: %v", err)), nil
	}

	options := &jira.GetQueryOptions{
		Expand: "comments",
	}

	issue, _, err := client.Issue.GetWithContext(ctx, issueKey, options)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get issue with comments: %v", err)), nil
	}

	if issue.Fields.Comments == nil || len(issue.Fields.Comments.Comments) == 0 {
		return mcp.NewToolResultText("No comments found"), nil
	}

	result := ""
	for i, comment := range issue.Fields.Comments.Comments {
		if i > 0 {
			result += "\n---\n"
		}
		result += fmt.Sprintf("%s (%s):\n%s", comment.Author.DisplayName, comment.Author.Name, comment.Body)
	}

	return mcp.NewToolResultText(result), nil
}

func createIssueHandler(ctx context.Context, client *jira.Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectKey, err := request.RequireString("project")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Missing or invalid 'project' argument: %v", err)), nil
	}

	description, err := request.RequireString("description")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Missing or invalid 'description' argument: %v", err)), nil
	}

	assignee := request.GetString("assignee", "")

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
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create issue: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Successfully created issue: %s", createdIssue.Key)), nil
}
