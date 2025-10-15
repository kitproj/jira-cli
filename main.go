package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/andygrunwald/go-jira"
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

	flag.StringVar(&host, "h", os.Getenv("JIRA_HOST"), "JIRA host (e.g., your-domain.atlassian.net, defaults to JIRA_HOST env var)")
	flag.StringVar(&token, "t", os.Getenv("JIRA_TOKEN"), "JIRA API token (defaults to JIRA_TOKEN env var)")
	flag.StringVar(&issueKey, "k", os.Getenv("JIRA_ISSUE_KEY"), "JIRA issue key (e.g., PROJ-123, defaults to JIRA_ISSUE_KEY env var)")
	flag.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "Usage:")
		fmt.Fprintln(w)
		fmt.Fprintln(w, "  jira get-issue - Get details of the specified JIRA issue")
		fmt.Fprintln(w, "  jira get-comments - Get comments of the specified JIRA issue")
		fmt.Fprintln(w, "  jira add-comment <comment> - Add a comment to the specified JIRA issue")
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
	if len(args) == 0 {
		return fmt.Errorf("unknown sub-command: (none provided)")
	}

	if host == "" {
		return fmt.Errorf("host is required")
	}
	if token == "" {
		return fmt.Errorf("token is required")
	}
	if issueKey == "" {
		return fmt.Errorf("issue key is required")
	}

	tp := jira.BearerAuthTransport{Token: token}

	var err error
	client, err = jira.NewClient(tp.Client(), "https://"+host)
	if err != nil {
		return fmt.Errorf("failed to create JIRA client: %w", err)
	}

	switch args[0] {
	case "get-issue":
		return getIssue(ctx)
	case "add-comment":
		if len(args) < 2 {
			return fmt.Errorf("comment message is required")
		}
		return addComment(ctx, args[1])
	case "get-comments":
		return getComments(ctx)
	default:
		return fmt.Errorf("unknown sub-command: %s", args[0])
	}
}

func getIssue(ctx context.Context) error {
	issue, _, err := client.Issue.GetWithContext(ctx, issueKey, nil)
	if err != nil {
		return fmt.Errorf("failed to get issue: %w", err)
	}

	fmt.Printf("Key:         %s\n", issue.Key)
	fmt.Printf("Status:      %s\n", issue.Fields.Status.Name)
	fmt.Printf("Summary:     %s\n", issue.Fields.Summary)
	fmt.Printf("Reporter:    %s (%s)\n", issue.Fields.Reporter.DisplayName, issue.Fields.Reporter.Name)
	fmt.Println("Description:")
	fmt.Println(issue.Fields.Description)

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
