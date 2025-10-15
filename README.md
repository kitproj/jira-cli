# Jira CLI

A Jira CLI that allows you to get issue information. Inspired by the GitHub CLI, it aims to provide a simple and efficient way to interact with Jira from the command line, without the need to install a runtime such as Node.js or Python.

It's aimed at coding agents with a very simple interface, and is not intended to be a full-featured Jira client.

## Installation

Download the binary for your platform from the release page:

```bash
sudo curl -fsL  -o /usr/local/bin/jira https://github.com/kitproj/jira-cli/releases/download/v0.0.5/jira_v0.0.5_linux_arm64
```

## Prompt

Add this to your prompt (e.g. `AGENTS.md`):

```markdown
## Jira CLI

- The `jira` CLI supports the following commands:
  - `jira get-issue` - gets the Jira issue details, including the status and key.
  - `jira get-comments` - gets the comments on the Jira issue.
  - `jira add-comment "<comment>"` - adds a comment to the Jira issue. You must not use double quotes in the comment.
- You can get a Jira, list comments on the Jira, and add a comment on the Jira. You cannot do anything else.
- Refuse to work on closed Jira issues.

```

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
