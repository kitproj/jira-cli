package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"syscall"

	"github.com/andygrunwald/go-jira"
	"github.com/kitproj/jira-cli/internal/config"
	"golang.org/x/term"
)

var (
	host     string
	token    string
	issueKey string
	client   *jira.Client
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	flag.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "Usage:")
		fmt.Fprintln(w)
		fmt.Fprintln(w, "  jira configure <host> - Configure JIRA host and token (reads token from stdin)")
		fmt.Fprintln(w, "  jira create-issue <project> <description> [assignee] - Create a new JIRA issue")
		fmt.Fprintln(w, "  jira get-issue <issue-key> - Get details of the specified JIRA issue")
		fmt.Fprintln(w, "  jira list-issues - List issues assigned to the current user")
		fmt.Fprintln(w, "  jira update-issue-status <issue-key> <status> - Update the status of the specified JIRA issue")
		fmt.Fprintln(w, "  jira get-comments <issue-key> - Get comments of the specified JIRA issue")
		fmt.Fprintln(w, "  jira add-comment <issue-key> <comment> - Add a comment to the specified JIRA issue")
		fmt.Fprintln(w, "  jira mcp-server - Start MCP server (stdio transport)")
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Options:")
		flag.PrintDefaults()
	}
	flag.Parse()

	if err := run(ctx, flag.Args()); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		flag.Usage()
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: jira <command> [args...]")
	}

	// First argument is the command
	command := args[0]

	switch command {
	case "configure":
		if len(args) < 2 {
			return fmt.Errorf("usage: jira configure <host>")
		}
		return configure(args[1])
	case "create-issue":
		if len(args) < 3 {
			return fmt.Errorf("usage: jira create-issue <project> <description> [assignee]")
		}
		project := args[1]
		description := args[2]
		var assignee string
		if len(args) >= 4 {
			assignee = args[3]
		}
		return executeCommand(ctx, func(ctx context.Context) error {
			return createIssue(ctx, project, description, assignee)
		})
	case "get-issue":
		if len(args) < 2 {
			return fmt.Errorf("usage: jira <command> <issue-key> [args...]")
		}
		issueKey = args[1]
		return executeCommand(ctx, getIssue)
	case "update-issue-status":
		if len(args) < 3 {
			return fmt.Errorf("usage: jira update-issue-status <issue-key> <status>")
		}
		issueKey = args[1]
		statusName := args[2]
		return executeCommand(ctx, func(ctx context.Context) error {
			return updateIssueStatus(ctx, statusName)
		})
	case "add-comment":
		if len(args) < 3 {
			return fmt.Errorf("usage: jira add-comment <issue-key> <comment>")
		}
		issueKey = args[1]
		message := args[2]
		return executeCommand(ctx, func(ctx context.Context) error {
			return addComment(ctx, message)
		})
	case "get-comments":
		if len(args) < 2 {
			return fmt.Errorf("usage: jira <command> <issue-key> [args...]")
		}
		issueKey = args[1]
		return executeCommand(ctx, getComments)
	case "list-issues":
		return executeCommand(ctx, listIssues)
	case "mcp-server":
		return runMCPServer(ctx)
	default:
		return fmt.Errorf("unknown sub-command: %s", command)
	}
}

func executeCommand(ctx context.Context, fn func(context.Context) error) error {
	// Load host from config file, or fall back to env var
	if host == "" {
		var err error
		host, err = config.LoadConfig()
		if err != nil {
			// Fall back to environment variable
			host = os.Getenv("JIRA_HOST")
		}
	}

	// Load token from keyring, or fall back to env var
	if token == "" {
		token = os.Getenv("JIRA_TOKEN")
	}
	if token == "" {
		var err error
		token, err = config.LoadToken(host)
		if err != nil {
			return err
		}
	}

	if host == "" {
		return fmt.Errorf("host is required")
	}
	if token == "" {
		return fmt.Errorf("token is required")
	}

	tp := jira.BearerAuthTransport{Token: token}

	var err error
	client, err = jira.NewClient(tp.Client(), "https://"+host)
	if err != nil {
		return fmt.Errorf("failed to create JIRA client: %w", err)
	}

	return fn(ctx)
}

func getIssue(ctx context.Context) error {
	issue, _, err := client.Issue.GetWithContext(ctx, issueKey, nil)
	if err != nil {
		return fmt.Errorf("failed to get issue: %w", err)
	}

	printField("Key", issue.Key)
	printField("Status", issue.Fields.Status.Name)
	printField("Summary", issue.Fields.Summary)
	printField("Reporter", fmt.Sprintf("%s (%s)", issue.Fields.Reporter.DisplayName, issue.Fields.Reporter.Name))
	printField("Description", issue.Fields.Description)

	// we need to only display fields the user can set, which are the editable fields
	editMetaInfo, _, err := client.Issue.GetEditMetaWithContext(ctx, issue)
	if err != nil {
		return fmt.Errorf("failed to get edit meta: %w", err)
	}

	for key, value := range issue.Fields.Unknowns {
		if _, ok := editMetaInfo.Fields[key]; ok && strings.HasPrefix(key, "customfield_") && isPrimitive(value) {
			name := editMetaInfo.Fields[key].(map[string]any)["name"].(string)
			printField(name, value)
		}
	}

	return nil
}

func isPrimitive(v any) bool {
	kind := reflect.ValueOf(v).Kind()
	switch kind {
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64,
		reflect.Complex64, reflect.Complex128,
		reflect.String:
		return true
	default:
		return false
	}
}

func printField(key string, value any) {
	valueStr := fmt.Sprint(value)
	multiLine := strings.Contains(valueStr, "\n")
	fmt.Printf("%-20s", key+":")
	if !multiLine {
		fmt.Printf(" %s\n", valueStr)
	} else {
		fmt.Println()
		for line := range strings.SplitSeq(valueStr, "\n") {
			fmt.Printf("%-20s %s\n", "", line)
		}
	}
}

// updateIssueStatus updates the status of a Jira issue using transitions
func updateIssueStatus(ctx context.Context, statusName string) error {
	// First, get the issue to check current status
	issue, _, err := client.Issue.GetWithContext(ctx, issueKey, nil)
	if err != nil {
		return fmt.Errorf("failed to get issue: %w", err)
	}

	// Check if already in the desired status
	if issue.Fields.Status.Name == statusName {
		fmt.Printf("Issue %s is already in status: %s\n", issueKey, statusName)
		return nil
	}

	// Get available transitions for this issue
	transitions, _, err := client.Issue.GetTransitionsWithContext(ctx, issueKey)
	if err != nil {
		return fmt.Errorf("failed to get transitions: %w", err)
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
		// List available statuses for better error message
		var availableStatuses []string
		for _, transition := range transitions {
			availableStatuses = append(availableStatuses, fmt.Sprintf("%q", transition.To.Name))
		}
		return fmt.Errorf("no transition found to status '%s'. Available statuses: %v", statusName, strings.Join(availableStatuses, ", "))
	}

	// Perform the transition
	_, err = client.Issue.DoTransition(issueKey, targetTransition.ID)
	if err != nil {
		return fmt.Errorf("failed to update issue status: %w", err)
	}

	fmt.Printf("Successfully updated issue %s to status: %s\n", issueKey, statusName)
	return nil
}

func addComment(ctx context.Context, message string) error {
	comment := &jira.Comment{
		Body: message,
	}

	_, _, err := client.Issue.AddCommentWithContext(ctx, issueKey, comment)
	if err != nil {
		return fmt.Errorf("failed to add comment: %w", err)
	}

	return nil
}

func getComments(ctx context.Context) error {
	options := &jira.GetQueryOptions{
		Expand: "comments",
	}

	issue, _, err := client.Issue.GetWithContext(ctx, issueKey, options)
	if err != nil {
		return fmt.Errorf("failed to get issue with comments: %w", err)
	}

	if issue.Fields.Comments == nil {
		fmt.Println("No comments found")
		return nil
	}

	for _, comment := range issue.Fields.Comments.Comments {
		fmt.Printf("%s (%s):\n", comment.Author.DisplayName, comment.Author.Name)
		fmt.Println(comment.Body)
		fmt.Println("---")
	}

	return nil
}

// createIssue creates a new JIRA issue with the specified project, description, and optional assignee
func createIssue(ctx context.Context, projectKey, description, assignee string) error {
	// Create a new issue with the Task issue type (most common default)
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
		return fmt.Errorf("failed to create issue: %w", err)
	}

	fmt.Printf("Successfully created issue: %s\n", createdIssue.Key)
	return nil
}

// configure reads the token from stdin and saves it to the keyring
func configure(host string) error {
	if host == "" {
		return fmt.Errorf("host is required")
	}

	fmt.Fprintf(os.Stderr, "To create a personal access token, visit: https://%s/secure/ViewProfile.jspa?selectedTab=com.atlassian.pats.pats-plugin:jira-user-personal-access-tokens\n", host)
	fmt.Fprintf(os.Stderr, "The token will be stored securely in your system's keyring.\n")
	fmt.Fprintf(os.Stderr, "\nEnter JIRA API token: ")

	// Read password with hidden input
	tokenBytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Fprintln(os.Stderr) // Print newline after hidden input
	if err != nil {
		return fmt.Errorf("failed to read token: %w", err)
	}

	token := string(tokenBytes)
	if token == "" {
		return fmt.Errorf("token cannot be empty")
	}

	// Save host to config file
	if err := config.SaveConfig(host); err != nil {
		return err
	}

	// Save token to keyring
	if err := config.SaveToken(host, token); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Configuration saved successfully for host: %s\n", host)
	return nil
}

// listIssues lists issues assigned to the current user
func listIssues(ctx context.Context) error {
	// JQL to find issues assigned to the current user, excluding closed issues, updated in last 14 days
	jql := "assignee = currentUser() AND resolution = Unresolved AND updated >= -14d ORDER BY updated DESC"

	// Search for issues using JQL
	issues, _, err := client.Issue.SearchWithContext(ctx, jql, &jira.SearchOptions{
		MaxResults: 50,
		Fields:     []string{"key", "summary", "status", "sprint"},
	})
	if err != nil {
		return fmt.Errorf("failed to search issues: %w", err)
	}

	if len(issues) == 0 {
		fmt.Println("No issues assigned to you")
		return nil
	}

	fmt.Printf("Found %d issue(s):\n\n", len(issues))
	for _, issue := range issues {
		sprintName := "-"

		// Get sprint from the Sprint field
		if issue.Fields.Sprint != nil && issue.Fields.Sprint.Name != "" {
			sprintName = issue.Fields.Sprint.Name
		}

		fmt.Printf("%-15s %-20s %-25s %s\n", issue.Key, issue.Fields.Status.Name, sprintName, issue.Fields.Summary)
	}

	return nil
}
