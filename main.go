package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

var (
	host     string
	token    string
	issueKey string
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
		fmt.Fprintln(w, "  jira get-issue")
		fmt.Fprintln(w, "  jira get-comments")
		fmt.Fprintln(w, "  jira add-comment <comment>")
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Options:")
		flag.PrintDefaults()
	}
	flag.Parse()

	if err := run(ctx, os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		flag.Usage()
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("unknown sub-command\nusage:\n - jira get-issue\n - jira get-recent-comments\n - jira add-comment <comment>")
	}

	if token == "" || host == "" || issueKey == "" {
		return fmt.Errorf("JIRA_TOKEN, JIRA_HOST, and JIRA_ISSUE_KEY must be set")
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
	url := fmt.Sprintf("https://%s/rest/api/2/issue/%s", host, issueKey)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to get issue: %s", resp.Status)
	}

	var result issue

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	fmt.Printf("Key:         %s\n", result.Key)
	fmt.Printf("Status:      %s\n", result.Fields.Status.Name)
	fmt.Printf("Summary:     %s\n", result.Fields.Summary)
	fmt.Printf("Reporter:    %s (%s)\n", result.Fields.Reporter.DisplayName, result.Fields.Reporter.Name)
	fmt.Println("Description:")
	fmt.Println(result.Fields.Description)

	return nil
}

func addComment(ctx context.Context, message string) error {
	url := fmt.Sprintf("https://%s/rest/api/2/issue/%s/comment", host, issueKey)

	comment := comment{Body: message}

	data, err := json.Marshal(comment)
	if err != nil {
		return fmt.Errorf("failed to marshal comment: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("failed to add comment: HTTP %d", resp.StatusCode)
	}

	return nil
}

func getComments(ctx context.Context) error {
	url := fmt.Sprintf("https://%s/rest/api/2/issue/%s/comment", host, issueKey)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("failed to get comments: HTTP %d", resp.StatusCode)
	}

	var result struct {
		Comments []comment `json:"comments"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	for _, comment := range result.Comments {
		fmt.Printf("%s:\n", comment.Author.Name)
		fmt.Println(comment.Body)
	}

	return nil
}
