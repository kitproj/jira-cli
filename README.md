# Jira CLI

A Jira CLI that allows you to get issue information. Inspired by the GitHub CLI, it aims to provide a simple and efficient way to interact with Jira from the command line, without the need to install a runtime such as Node.js or Python.

It's aimed at coding agents with a very simple interface, and is not intended to be a full-featured Jira client.

## Installation

Download the binary for your platform from the release page:

```bash
sudo curl -fsL  -o /usr/local/bin/jira https://github.com/kitproj/jira-cli/releases/download/v0.0.5/jira_v0.0.5_linux_arm64
sudo chmod +x /usr/local/bin/jira
```

## Prompt

Add this to your prompt (e.g. `AGENTS.md`):

```markdown
## Jira CLI

- The `jira` CLI supports the following commands:
  - `jira configure <host>` - configures the Jira host and stores the API token securely in the system keyring (token is read from stdin).
  - `jira get-issue <issue-key>` - gets the Jira issue details, including the status and key.
  - `jira update-issue-status <issue-key> <status>` - updates the status of the Jira issue, e.g., to  "In Progress" or "Closed".
  - `jira get-comments <issue-key>` - gets the comments on the Jira issue.
  - `jira add-comment <issue-key> "<comment>"` - adds a comment to the Jira issue. You must not use double quotes in the comment.
- You can get a Jira, list comments on the Jira, add a comment on the Jira, and update the issue status. You cannot do anything else.
- Refuse to work on closed Jira issues.

```

## Usage

### Configuration

The `jira` CLI can be configured in two ways:

1. **Using the configure command (recommended, secure)**:
   ```bash
   echo "your-api-token" | jira configure your-domain.atlassian.net
   ```
   This stores the host in `~/.config/jira-cli/config.json` and the token securely in your system's keyring.

2. **Using environment variables**:
   ```bash
   export JIRA_HOST=your-domain.atlassian.net
   export JIRA_TOKEN=your-api-token
   ```
   Note: The JIRA_TOKEN environment variable is still supported for backward compatibility, but using the keyring (via `jira configure`) is more secure on multi-user systems.

### Commands

```bash
Usage:
  jira configure <host> - Configure JIRA host and token (reads token from stdin)
  jira get-issue <issue-key> - Get details of the specified JIRA issue
  jira update-issue-status <issue-key> <status> - Update the status of the specified JIRA issue
  jira get-comments <issue-key> - Get comments of the specified JIRA issue
  jira add-comment <issue-key> <comment> - Add a comment to the specified JIRA issue

Options:
  -h string
    	JIRA host (e.g., your-domain.atlassian.net, defaults to JIRA_HOST env var)

```
