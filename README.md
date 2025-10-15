# jira-cli

A Jira CLI that allows you to get issue information. Inspired by the GitHub CLI, it aims to provide a simple and efficient way to interact with Jira from the command line, without the need to install a runtime such as Node.js or Python.

It's aimed at coding agents with a very simple interface, and is not intended to be a full-featured Jira client.

## Usage

```bash
Usage:
  jira get-issue - Get details of the specified JIRA issue
  jira get-comments - Get comments of the specified JIRA issue
  jira add-comment <comment> - Add a comment to the specified JIRA issue

Options:
  -h string
    	JIRA host (e.g., your-domain.atlassian.net, defaults to JIRA_HOST env var) (default "")
  -k string
    	JIRA issue key (e.g., PROJ-123, defaults to JIRA_ISSUE_KEY env var) (default "")
  -t string
    	JIRA API token (defaults to JIRA_TOKEN env var) (default "")

```
